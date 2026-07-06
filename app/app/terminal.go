package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type NativeTerminalRequest struct {
	ProviderID string   `json:"providerId"`
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	CWD        string   `json:"cwd"`
	Label      string   `json:"label"`
}

type NativeTerminalSessionInfo struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Command   string   `json:"command"`
	Args      []string `json:"args"`
	CWD       string   `json:"cwd"`
	Status    string   `json:"status"`
	StartedAt string   `json:"startedAt"`
}

type terminalSession struct {
	info   NativeTerminalSessionInfo
	cancel context.CancelFunc
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

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}

	info := NativeTerminalSessionInfo{
		ID:        sessionID,
		Label:     fallback(req.Label, command),
		Command:   command,
		Args:      append([]string(nil), req.Args...),
		CWD:       absCWD,
		Status:    "running",
		StartedAt: time.Now().Format(time.RFC3339),
	}

	m.mu.Lock()
	m.sessions[sessionID] = &terminalSession{info: info, cancel: cancel}
	m.mu.Unlock()

	if err := cmd.Start(); err != nil {
		cancel()
		m.remove(sessionID)
		return nil, err
	}

	emitTerminal(ctx, "terminal:started", info)
	go scanTerminal(ctx, sessionID, "stdout", stdout)
	go scanTerminal(ctx, sessionID, "stderr", stderr)
	go func() {
		err := cmd.Wait()
		status := "completed"
		code := cmd.ProcessState.ExitCode()
		if err != nil {
			status = "failed"
			if runCtx.Err() == context.Canceled {
				status = "stopped"
			}
		}
		m.remove(sessionID)
		emitTerminal(ctx, "terminal:exit", map[string]any{
			"id":     sessionID,
			"status": status,
			"code":   code,
		})
	}()

	return &info, nil
}

func (m *NativeTerminalManager) Stop(sessionID string) bool {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	m.mu.Unlock()
	if !ok {
		return false
	}
	session.cancel()
	return true
}

func (m *NativeTerminalManager) remove(sessionID string) {
	m.mu.Lock()
	delete(m.sessions, sessionID)
	m.mu.Unlock()
}

func scanTerminal(ctx context.Context, sessionID, stream string, pipe any) {
	reader, ok := pipe.(interface {
		Read([]byte) (int, error)
	})
	if !ok {
		return
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		emitTerminal(ctx, "terminal:output", map[string]string{
			"id":     sessionID,
			"stream": stream,
			"data":   scanner.Text() + "\n",
		})
	}
}

func emitTerminal(ctx context.Context, name string, payload any) {
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, name, payload)
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
