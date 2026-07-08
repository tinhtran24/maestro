package app

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type AgentMode string

const (
	AgentModePlanner  AgentMode = "planner"
	AgentModeCoding   AgentMode = "coding"
	AgentModeReview   AgentMode = "review"
	AgentModeDebug    AgentMode = "debug"
	AgentModeResearch AgentMode = "research"
)

type AgentStatus string

const (
	AgentStatusStarting AgentStatus = "starting"
	AgentStatusRunning  AgentStatus = "running"
	AgentStatusFailed   AgentStatus = "failed"
	AgentStatusStopped  AgentStatus = "stopped"
)

type AgentEventType string

const (
	AgentEventSessionStarted AgentEventType = "session_started"
	AgentEventPromptSent     AgentEventType = "prompt_sent"
	AgentEventError          AgentEventType = "error"
	AgentEventSessionStopped AgentEventType = "session_stopped"
)

type AgentEventInfo struct {
	SessionID string         `json:"sessionId"`
	Type      AgentEventType `json:"type"`
	Payload   string         `json:"payload"`
	Time      string         `json:"time"`
}

type AgentSessionInfo struct {
	ID             string           `json:"id"`
	ProviderID     string           `json:"providerId"`
	ProviderName   string           `json:"providerName"`
	ProjectID      string           `json:"projectId"`
	TaskID         string           `json:"taskId"`
	Mode           AgentMode        `json:"mode"`
	Status         AgentStatus      `json:"status"`
	Workdir        string           `json:"workdir"`
	Prompt         string           `json:"prompt"`
	PromptHistory  []PromptRecord   `json:"promptHistory"`
	Conversation   []AgentEventInfo `json:"conversation"`
	TerminalID     string           `json:"terminalId"`
	TranscriptPath string           `json:"transcriptPath"`
	ApprovedPlan   string           `json:"approvedPlan,omitempty"`
	Result         string           `json:"result,omitempty"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
}

type AgentProviderCapabilityInfo struct {
	ProviderID        string `json:"providerId"`
	Command           string `json:"command"`
	SupportsPTY       bool   `json:"supportsPty"`
	SupportsStreaming bool   `json:"supportsStreaming"`
	SupportsInterrupt bool   `json:"supportsInterrupt"`
	SupportsResume    bool   `json:"supportsResume"`
	SetupHint         string `json:"setupHint"`
}

type StartAgentRequest struct {
	Root           string    `json:"root"`
	ProviderID     string    `json:"providerId"`
	ProjectID      string    `json:"projectId"`
	TaskID         string    `json:"taskId"`
	Mode           AgentMode `json:"mode"`
	Prompt         string    `json:"prompt"`
	Context        string    `json:"context"`
	Acceptance     string    `json:"acceptanceCriteria"`
	Constraints    string    `json:"constraints"`
	AllowedFiles   []string  `json:"allowedFiles"`
	PreviousPlan   string    `json:"previousPlan"`
	ExpectedOutput string    `json:"expectedOutput"`
	CustomCommand  string    `json:"customCommand"`
	Rows           uint16    `json:"rows"`
	Cols           uint16    `json:"cols"`
}

type SendAgentInputRequest struct {
	Root      string `json:"root"`
	SessionID string `json:"sessionId"`
	Input     string `json:"input"`
}

type AgentProvider interface {
	ID() string
	Name() string
	Capability() AgentProviderCapabilityInfo
	IsInstalled(ctx context.Context, customCommand string) bool
	Start(ctx context.Context, terminal *NativeTerminalManager, req StartAgentRequest, prompt string) (*NativeTerminalSessionInfo, error)
	Send(ctx context.Context, terminal *NativeTerminalManager, sessionID string, input string) error
	Resize(ctx context.Context, terminal *NativeTerminalManager, sessionID string, cols int, rows int) error
	Interrupt(ctx context.Context, terminal *NativeTerminalManager, sessionID string) error
	Stop(ctx context.Context, terminal *NativeTerminalManager, sessionID string) error
}

type AgentProviderRegistry struct {
	providers map[string]AgentProvider
	order     []string
}

func NewAgentProviderRegistry() *AgentProviderRegistry {
	registry := &AgentProviderRegistry{providers: map[string]AgentProvider{}}
	for _, provider := range []AgentProvider{
		NewCommandAgentProvider("claude-code", "Claude Code", "claude", "Install Claude Code and authenticate it before assigning tasks.", claudeAgentArgs),
		NewCommandAgentProvider("codex", "Codex", "codex", "Install Codex and ensure `codex` is on PATH.", codexAgentArgs),
		NewCommandAgentProvider("gemini-cli", "Gemini CLI", "gemini", "Install Gemini CLI and ensure `gemini` is on PATH.", geminiAgentArgs),
		NewCommandAgentProvider("crush", "Crush", "crush", "Install Crush and ensure `crush` is on PATH.", crushAgentArgs),
		NewCommandAgentProvider("opencode", "OpenCode", "opencode", "Install OpenCode and finish setup before assigning it.", opencodeAgentArgs),
		NewCommandAgentProvider("cursor-agent", "Cursor Agent", "cursor-agent", "Install Cursor Agent and ensure `cursor-agent` is on PATH.", cursorAgentArgs),
		NewCommandAgentProvider("aider", "Aider", "aider", "Install Aider and ensure `aider` is on PATH.", aiderAgentArgs),
		NewCommandAgentProvider("goose", "Goose", "goose", "Install Goose and ensure `goose` is on PATH.", gooseAgentArgs),
		NewCustomCommandAgentProvider(),
	} {
		registry.Add(provider)
	}
	return registry
}

func (r *AgentProviderRegistry) Add(provider AgentProvider) {
	if provider == nil {
		return
	}
	id := strings.ToLower(strings.TrimSpace(provider.ID()))
	if _, exists := r.providers[id]; !exists {
		r.order = append(r.order, id)
	}
	r.providers[id] = provider
}

func (r *AgentProviderRegistry) Get(id string) (AgentProvider, bool) {
	if r == nil {
		return nil, false
	}
	provider, ok := r.providers[strings.ToLower(strings.TrimSpace(id))]
	return provider, ok
}

func (r *AgentProviderRegistry) All() []AgentProvider {
	if r == nil {
		return nil
	}
	out := make([]AgentProvider, 0, len(r.order))
	for _, id := range r.order {
		out = append(out, r.providers[id])
	}
	return out
}

type CommandAgentProvider struct {
	id        string
	name      string
	command   string
	setupHint string
	args      func(StartAgentRequest, string) []string
	custom    bool
}

func NewCommandAgentProvider(id, name, command, setupHint string, args func(StartAgentRequest, string) []string) CommandAgentProvider {
	return CommandAgentProvider{id: id, name: name, command: command, setupHint: setupHint, args: args}
}

func NewCustomCommandAgentProvider() CommandAgentProvider {
	return CommandAgentProvider{
		id:        "custom-local",
		name:      "Custom Local Agent",
		command:   "",
		setupHint: "Enter a local agent command available on PATH.",
		args:      customAgentArgs,
		custom:    true,
	}
}

func (p CommandAgentProvider) ID() string {
	return p.id
}

func (p CommandAgentProvider) Name() string {
	return p.name
}

func (p CommandAgentProvider) Capability() AgentProviderCapabilityInfo {
	return AgentProviderCapabilityInfo{
		ProviderID:        p.ID(),
		Command:           p.command,
		SupportsPTY:       true,
		SupportsStreaming: true,
		SupportsInterrupt: true,
		SupportsResume:    false,
		SetupHint:         p.setupHint,
	}
}

func (p CommandAgentProvider) IsInstalled(_ context.Context, customCommand string) bool {
	command := p.command
	if p.custom {
		command = strings.TrimSpace(customCommand)
	}
	if command == "" {
		return false
	}
	_, err := exec.LookPath(command)
	return err == nil
}

func (p CommandAgentProvider) Start(ctx context.Context, terminal *NativeTerminalManager, req StartAgentRequest, prompt string) (*NativeTerminalSessionInfo, error) {
	command := p.command
	if p.custom {
		command = strings.TrimSpace(req.CustomCommand)
	}
	if command == "" {
		return nil, fmt.Errorf("%s command is required", p.Name())
	}
	if !p.IsInstalled(ctx, req.CustomCommand) {
		return nil, fmt.Errorf("provider not installed: %s. Install command: %s", p.Name(), p.setupHint)
	}
	args := []string(nil)
	if p.args != nil {
		args = p.args(req, prompt)
	}
	return terminal.Start(ctx, NativeTerminalRequest{
		ProviderID: p.ID(),
		TaskID:     req.TaskID,
		Step:       string(normalizeAgentMode(req.Mode)),
		Command:    command,
		Args:       args,
		CWD:        reqWorkdir(req),
		Label:      p.Name() + " " + string(normalizeAgentMode(req.Mode)),
		Rows:       req.Rows,
		Cols:       req.Cols,
	})
}

func (p CommandAgentProvider) Send(_ context.Context, terminal *NativeTerminalManager, sessionID string, input string) error {
	return terminal.Write(NativeTerminalInputRequest{SessionID: sessionID, Data: input})
}

func (p CommandAgentProvider) Resize(_ context.Context, terminal *NativeTerminalManager, sessionID string, cols int, rows int) error {
	return terminal.Resize(NativeTerminalResizeRequest{SessionID: sessionID, Cols: uint16(cols), Rows: uint16(rows)})
}

func (p CommandAgentProvider) Interrupt(_ context.Context, terminal *NativeTerminalManager, sessionID string) error {
	if !terminal.Stop(sessionID) {
		return fmt.Errorf("agent session not found: %s", sessionID)
	}
	return nil
}

func (p CommandAgentProvider) Stop(ctx context.Context, terminal *NativeTerminalManager, sessionID string) error {
	return p.Interrupt(ctx, terminal, sessionID)
}

func (p *RealProvider) ListAgentProviders(ctx context.Context) ([]ProviderInfo, error) {
	registry := NewAgentProviderRegistry()
	out := make([]ProviderInfo, 0, len(registry.All()))
	for _, adapter := range registry.All() {
		capability := adapter.Capability()
		provider := ProviderInfo{
			ID:          adapter.ID(),
			Name:        adapter.Name(),
			Command:     capability.Command,
			Status:      "not_found",
			Type:        "cli",
			SetupHint:   capability.SetupHint,
			SupportsRun: true,
		}
		if adapter.ID() == "custom-local" {
			provider.Status = "needs_setup"
			provider.Type = "custom"
		} else if adapter.IsInstalled(ctx, "") {
			provider.Status = "installed"
			if path, err := exec.LookPath(provider.Command); err == nil {
				provider.Path = &path
			}
			if version := commandVersion(ctx, provider.Command); version != "" {
				provider.Version = &version
			}
		}
		out = append(out, provider)
	}
	return out, nil
}

func (p *RealProvider) StartAgentSession(ctx context.Context, terminal *NativeTerminalManager, req StartAgentRequest) (*AgentSessionInfo, error) {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return nil, err
	}
	req.Root = root
	req.Mode = normalizeAgentMode(req.Mode)
	if strings.TrimSpace(req.ProviderID) == "" {
		return nil, fmt.Errorf("agent provider is required")
	}
	registry := NewAgentProviderRegistry()
	provider, ok := registry.Get(req.ProviderID)
	if !ok {
		return nil, fmt.Errorf("agent provider is not registered: %s", req.ProviderID)
	}
	if !provider.IsInstalled(ctx, req.CustomCommand) {
		return nil, fmt.Errorf("provider not installed: %s. Install command: %s", provider.Name(), provider.Capability().SetupHint)
	}
	prompt := buildAgentPrompt(req)
	now := p.now().UTC().Format(time.RFC3339)
	native, err := provider.Start(ctx, terminal, req, prompt)
	if err != nil {
		return nil, err
	}
	session := AgentSessionInfo{
		ID:             "agent-" + stableID(native.ID+"-"+now),
		ProviderID:     provider.ID(),
		ProviderName:   provider.Name(),
		ProjectID:      firstNonEmpty(req.ProjectID, workspaceName(root)),
		TaskID:         req.TaskID,
		Mode:           req.Mode,
		Status:         AgentStatusRunning,
		Workdir:        reqWorkdir(req),
		Prompt:         prompt,
		PromptHistory:  []PromptRecord{{At: now, Prompt: prompt}},
		TerminalID:     native.ID,
		TranscriptPath: native.TranscriptPath,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	session.Conversation = append(session.Conversation,
		AgentEventInfo{SessionID: session.ID, Type: AgentEventSessionStarted, Payload: provider.Name(), Time: now},
		AgentEventInfo{SessionID: session.ID, Type: AgentEventPromptSent, Payload: prompt, Time: now},
	)
	store := NewWorkspaceStore(root)
	if err := store.WithLock(func() error {
		if err := writeAgentSession(store, session); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-agent-start-" + stableID(session.ID), At: now, Kind: "AgentSessionStarted", Message: provider.Name() + " " + string(req.Mode)})
	}); err != nil {
		return nil, err
	}
	return &session, nil
}

func (p *RealProvider) ListAgentSessions(root string) ([]AgentSessionInfo, error) {
	root, err := normalizeWorkspaceRoot(root)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, ".thanos", "agent-sessions")
	items, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []AgentSessionInfo{}, nil
		}
		return nil, err
	}
	out := make([]AgentSessionInfo, 0, len(items))
	for _, item := range items {
		if item.IsDir() || filepath.Ext(item.Name()) != ".json" {
			continue
		}
		var session AgentSessionInfo
		if err := readJSONFile(filepath.Join(dir, item.Name()), &session); err != nil {
			return nil, err
		}
		out = append(out, session)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

func (p *RealProvider) SendAgentInput(ctx context.Context, terminal *NativeTerminalManager, req SendAgentInputRequest) error {
	root, err := normalizeWorkspaceRoot(req.Root)
	if err != nil {
		return err
	}
	session, err := loadAgentSession(root, req.SessionID)
	if err != nil {
		return err
	}
	registry := NewAgentProviderRegistry()
	provider, ok := registry.Get(session.ProviderID)
	if !ok {
		return fmt.Errorf("agent provider is not registered: %s", session.ProviderID)
	}
	input := req.Input
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}
	if err := provider.Send(ctx, terminal, session.TerminalID, input); err != nil {
		return err
	}
	now := p.now().UTC().Format(time.RFC3339)
	session.Conversation = append(session.Conversation, AgentEventInfo{SessionID: session.ID, Type: "stdin", Payload: req.Input, Time: now})
	session.UpdatedAt = now
	store := NewWorkspaceStore(root)
	return store.WithLock(func() error {
		return writeAgentSession(store, session)
	})
}

func (p *RealProvider) StopAgentSession(ctx context.Context, terminal *NativeTerminalManager, root string, sessionID string) error {
	root, err := normalizeWorkspaceRoot(root)
	if err != nil {
		return err
	}
	session, err := loadAgentSession(root, sessionID)
	if err != nil {
		return err
	}
	registry := NewAgentProviderRegistry()
	provider, ok := registry.Get(session.ProviderID)
	if !ok {
		return fmt.Errorf("agent provider is not registered: %s", session.ProviderID)
	}
	if err := provider.Stop(ctx, terminal, session.TerminalID); err != nil {
		return err
	}
	now := p.now().UTC().Format(time.RFC3339)
	session.Status = AgentStatusStopped
	session.UpdatedAt = now
	session.Conversation = append(session.Conversation, AgentEventInfo{SessionID: session.ID, Type: AgentEventSessionStopped, Payload: "Stopped by user.", Time: now})
	store := NewWorkspaceStore(root)
	return store.WithLock(func() error {
		if err := writeAgentSession(store, session); err != nil {
			return err
		}
		return store.AppendEvent(EventInfo{ID: "event-agent-stop-" + stableID(session.ID+"-"+now), At: now, Kind: "AgentSessionStopped", Message: session.ID})
	})
}

func writeAgentSession(store *WorkspaceStore, session AgentSessionInfo) error {
	return store.WriteJSON(store.Path("agent-sessions", session.ID+".json"), session)
}

func loadAgentSession(root string, sessionID string) (AgentSessionInfo, error) {
	var session AgentSessionInfo
	if strings.TrimSpace(sessionID) == "" {
		return session, fmt.Errorf("agent session ID is required")
	}
	path := filepath.Join(root, ".thanos", "agent-sessions", sessionID+".json")
	if err := readJSONFile(path, &session); err != nil {
		return session, err
	}
	return session, nil
}

func readJSONFile(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func buildAgentPrompt(req StartAgentRequest) string {
	mode := normalizeAgentMode(req.Mode)
	sections := []struct {
		title string
		body  string
	}{
		{"Task", req.Prompt},
		{"Context", req.Context},
		{"Acceptance Criteria", req.Acceptance},
		{"Constraints", req.Constraints},
		{"Allowed Files", strings.Join(req.AllowedFiles, "\n")},
		{"Previous Plan", req.PreviousPlan},
		{"Expected Output", firstNonEmpty(req.ExpectedOutput, defaultExpectedOutput(mode))},
	}
	var builder strings.Builder
	builder.WriteString("Mode: " + string(mode) + "\n\n")
	for _, section := range sections {
		body := strings.TrimSpace(section.body)
		if body == "" {
			body = "Not provided."
		}
		builder.WriteString(section.title + "\n")
		builder.WriteString(body + "\n\n")
	}
	return strings.TrimSpace(builder.String())
}

func defaultExpectedOutput(mode AgentMode) string {
	switch mode {
	case AgentModePlanner:
		return "Return a concise plan, open questions, risks, and verification strategy."
	case AgentModeReview:
		return "Return findings first with file references, then residual risk."
	case AgentModeDebug:
		return "Return the cause, evidence, fix path, and verification command."
	default:
		return "Implement the requested work and report changed files and verification results."
	}
}

func normalizeAgentMode(mode AgentMode) AgentMode {
	switch mode {
	case AgentModePlanner, AgentModeCoding, AgentModeReview, AgentModeDebug, AgentModeResearch:
		return mode
	default:
		return AgentModeCoding
	}
}

func reqWorkdir(req StartAgentRequest) string {
	root := strings.TrimSpace(req.Root)
	if root == "" {
		return "."
	}
	if strings.TrimSpace(req.TaskID) != "" {
		if tasks, err := loadTasks(root); err == nil {
			for _, task := range tasks {
				if task.ID == req.TaskID {
					if worktree := resolveTaskWorktree(root, task); worktree != "" {
						return worktree
					}
				}
			}
		}
	}
	return root
}

func claudeAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"--print", prompt}
}

func codexAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"exec", prompt}
}

func geminiAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"--prompt", prompt}
}

func crushAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"run", prompt}
}

func opencodeAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"run", prompt}
}

func cursorAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{prompt}
}

func aiderAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"--message", prompt}
}

func gooseAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{"run", "--text", prompt}
}

func customAgentArgs(_ StartAgentRequest, prompt string) []string {
	return []string{prompt}
}
