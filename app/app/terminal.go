package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type NativeTerminalRequest struct {
	ProviderID string   `json:"providerId"`
	TaskID     string   `json:"taskId"`
	Step       string   `json:"step"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	CWD        string   `json:"cwd"`
	Label      string   `json:"label"`
	Rows       uint16   `json:"rows"`
	Cols       uint16   `json:"cols"`
}

type NativeTerminalInputRequest struct {
	SessionID string `json:"sessionId"`
	Data      string `json:"data"`
}

type NativeTerminalResizeRequest struct {
	SessionID string `json:"sessionId"`
	Rows      uint16 `json:"rows"`
	Cols      uint16 `json:"cols"`
}

type NativeTerminalSessionInfo struct {
	ID             string   `json:"id"`
	Label          string   `json:"label"`
	ProviderID     string   `json:"providerId"`
	TaskID         string   `json:"taskId"`
	Step           string   `json:"step"`
	Command        string   `json:"command"`
	Args           []string `json:"args"`
	CWD            string   `json:"cwd"`
	Status         string   `json:"status"`
	PTYID          string   `json:"ptyId"`
	TranscriptPath string   `json:"transcriptPath"`
	StartedAt      string   `json:"startedAt"`
	EndedAt        string   `json:"endedAt,omitempty"`
}

type terminalSession struct {
	info       NativeTerminalSessionInfo
	cancel     context.CancelFunc
	pty        *os.File
	transcript *os.File
}

type NativeTerminalManager struct {
	mu       sync.Mutex
	sessions map[string]*terminalSession
}

func NewNativeTerminalManager() *NativeTerminalManager {
	return &NativeTerminalManager{sessions: map[string]*terminalSession{}}
}

func (m *NativeTerminalManager) Start(ctx context.Context, req NativeTerminalRequest) (*NativeTerminalSessionInfo, error) {
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return nil, fmt.Errorf("%s not found on PATH", command)
	}
	cwd := strings.TrimSpace(req.CWD)
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	absCWD, err := filepath.Abs(cwd)
	if err != nil {
		return nil, err
	}
	if stat, err := os.Stat(absCWD); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("cwd is not a directory: %s", absCWD)
	}

	sessionID := fmt.Sprintf("native-%d", time.Now().UnixNano())
	runCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(runCtx, path, req.Args...)
	cmd.Dir = absCWD
	cmd.Env = scrubAuthEnv(os.Environ())

	rows, cols := normalizePTYSize(req.Rows, req.Cols)
	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		cancel()
		return nil, err
	}

	transcriptPath := terminalTranscriptPath(absCWD, sessionID)
	transcript, err := openTerminalTranscript(transcriptPath, req, sessionID)
	if err != nil {
		_ = ptyFile.Close()
		cancel()
		return nil, err
	}

	startedAt := time.Now().UTC().Format(time.RFC3339)
	info := NativeTerminalSessionInfo{
		ID:             sessionID,
		Label:          fallback(req.Label, command),
		ProviderID:     req.ProviderID,
		TaskID:         req.TaskID,
		Step:           req.Step,
		Command:        command,
		Args:           append([]string(nil), req.Args...),
		CWD:            absCWD,
		Status:         "running",
		PTYID:          sessionID,
		TranscriptPath: transcriptPath,
		StartedAt:      startedAt,
	}

	m.mu.Lock()
	m.sessions[sessionID] = &terminalSession{info: info, cancel: cancel, pty: ptyFile, transcript: transcript}
	m.mu.Unlock()

	emitTerminal(ctx, "terminal:started", info)
	go m.readPTY(ctx, sessionID, ptyFile, transcript)
	go func() {
		err := cmd.Wait()
		status := "completed"
		code := 0
		if cmd.ProcessState != nil {
			code = cmd.ProcessState.ExitCode()
		}
		if err != nil {
			status = "failed"
			if runCtx.Err() == context.Canceled {
				status = "stopped"
			}
		}
		endedAt := time.Now().UTC().Format(time.RFC3339)
		m.finish(sessionID, status, endedAt)
		emitTerminal(ctx, "terminal:exit", map[string]any{
			"id":             sessionID,
			"status":         status,
			"code":           code,
			"transcriptPath": transcriptPath,
			"endedAt":        endedAt,
		})
	}()

	return &info, nil
}

func (m *NativeTerminalManager) Write(req NativeTerminalInputRequest) error {
	m.mu.Lock()
	session, ok := m.sessions[req.SessionID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("terminal session not found: %s", req.SessionID)
	}
	if _, err := session.pty.Write([]byte(req.Data)); err != nil {
		return err
	}
	return nil
}

func (m *NativeTerminalManager) Resize(req NativeTerminalResizeRequest) error {
	m.mu.Lock()
	session, ok := m.sessions[req.SessionID]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("terminal session not found: %s", req.SessionID)
	}
	rows, cols := normalizePTYSize(req.Rows, req.Cols)
	return pty.Setsize(session.pty, &pty.Winsize{Rows: rows, Cols: cols})
}

func (m *NativeTerminalManager) Stop(sessionID string) bool {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	m.mu.Unlock()
	if !ok {
		return false
	}
	session.cancel()
	_ = session.pty.Close()
	return true
}

func (m *NativeTerminalManager) readPTY(ctx context.Context, sessionID string, ptyFile *os.File, transcript *os.File) {
	buffer := make([]byte, 8192)
	for {
		n, err := ptyFile.Read(buffer)
		if n > 0 {
			data := append([]byte(nil), buffer[:n]...)
			_, _ = transcript.Write(data)
			_ = transcript.Sync()
			emitTerminal(ctx, "terminal:output", map[string]string{
				"id":     sessionID,
				"stream": "pty",
				"data":   string(data),
			})
		}
		if err != nil {
			if err != io.EOF {
				_, _ = transcript.WriteString("\n[terminal read closed]\n")
			}
			return
		}
	}
}

func (m *NativeTerminalManager) finish(sessionID string, status string, endedAt string) {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	if ok {
		session.info.Status = status
		session.info.EndedAt = endedAt
		delete(m.sessions, sessionID)
	}
	m.mu.Unlock()
	if ok {
		_, _ = session.transcript.WriteString(fmt.Sprintf("\n[exit] %s at %s\n", status, endedAt))
		_ = session.transcript.Sync()
		_ = session.transcript.Close()
		_ = session.pty.Close()
	}
}

func emitTerminal(ctx context.Context, name string, payload any) {
	if ctx == nil || fmt.Sprintf("%T", ctx) == "context.backgroundCtx" || fmt.Sprintf("%T", ctx) == "context.todoCtx" {
		return
	}
	runtime.EventsEmit(ctx, name, payload)
}

func normalizePTYSize(rows uint16, cols uint16) (uint16, uint16) {
	if rows == 0 {
		rows = 24
	}
	if cols == 0 {
		cols = 100
	}
	return rows, cols
}

func terminalTranscriptPath(cwd string, sessionID string) string {
	return filepath.Join(cwd, ".thanos", "terminal", "sessions", sessionID+".log")
}

func openTerminalTranscript(path string, req NativeTerminalRequest, sessionID string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	header := fmt.Sprintf("[session] %s\n[provider] %s\n[task] %s\n[step] %s\n[command] %s %s\n[cwd] %s\n[start]\n",
		sessionID,
		req.ProviderID,
		req.TaskID,
		req.Step,
		req.Command,
		strings.Join(req.Args, " "),
		req.CWD,
	)
	if _, err := file.WriteString(header); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

func scrubAuthEnv(env []string) []string {
	blocked := map[string]bool{
		"ANTHROPIC_API_KEY":       true,
		"ANTHROPIC_AUTH_TOKEN":    true,
		"CLAUDE_CODE_OAUTH_TOKEN": true,
		"OPENAI_API_KEY":          true,
		"GEMINI_API_KEY":          true,
		"GOOGLE_API_KEY":          true,
	}
	out := make([]string, 0, len(env))
	for _, item := range env {
		key, _, ok := strings.Cut(item, "=")
		if ok && blocked[key] {
			continue
		}
		out = append(out, item)
	}
	return out
}
