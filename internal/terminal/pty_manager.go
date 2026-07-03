package terminal

import "fmt"

type PTYManager struct {
	sessions map[string]Session
}

func NewPTYManager() *PTYManager {
	return &PTYManager{sessions: map[string]Session{}}
}

func (m *PTYManager) Add(session Session) error {
	if session.ID == "" {
		return fmt.Errorf("session id is required")
	}
	m.sessions[session.ID] = session
	return nil
}
