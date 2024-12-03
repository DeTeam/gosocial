package main

type Connection struct {
	Source string
	Target string
}

func (c Connection) String() string {
	return c.Source + " -> " + c.Target
}

type Storage interface {
	ListConnections(handle string) ([]Connection, error)
	NewInvite(handle string) (string, error)
	UseInvite(handle, inviteToken string) error
}

type MemoryStorage struct{}

func (m *MemoryStorage) ListConnections(handle string) ([]Connection, error) {
	return nil, nil
}

func (m *MemoryStorage) NewInvite(handle string) (string, error) {
	return "test-invite-token", nil
}

func (m *MemoryStorage) UseInvite(handle, inviteToken string) error {
	return nil
}
