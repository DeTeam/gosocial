package main

import (
	"database/sql"

	gonanoid "github.com/matoous/go-nanoid/v2"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type Connection struct {
	Source string
	Target string
}

type Storage interface {
	ListConnections(handle string) ([]Connection, error)
	NewInvite(handle string) (string, error)
	UseInvite(handle, inviteToken string) error
}

type memoryStorage struct {
	db *sql.DB
}

func NewMemoryStorage() (*memoryStorage, error) {
	db, err := sql.Open("sqlite3", ":memory:")

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT NOT NULL,
			target TEXT NOT NULL,
			UNIQUE(source, target)
		);
		CREATE TABLE IF NOT EXISTS Invites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			handle TEXT NOT NULL,
			token TEXT NOT NULL
		);
	`)
	if err != nil {
		return nil, err
	}

	return &memoryStorage{
		db: db,
	}, nil
}

func (m *memoryStorage) ListConnections(handle string) ([]Connection, error) {
	rows, err := m.db.Query("SELECT source, target FROM Connections WHERE source = ?", handle)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []Connection
	for rows.Next() {
		var conn Connection
		if err := rows.Scan(&conn.Source, &conn.Target); err != nil {
			return nil, err
		}
		connections = append(connections, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return connections, nil
}

func (m *memoryStorage) NewInvite(handle string) (string, error) {
	id, err := gonanoid.New()

	if err != nil {
		return "", err
	}

	_, err = m.db.Exec("INSERT INTO Invites (handle, token) VALUES (?, ?)", handle, id)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (m *memoryStorage) UseInvite(currentUser, inviteToken string) error {
	var handle string
	err := m.db.QueryRow("SELECT handle FROM Invites WHERE token = ?", inviteToken).Scan(&handle)
	if err != nil {
		return err
	}

	if handle == currentUser {
		return nil
	}

	_, err = m.db.Exec("DELETE FROM Invites WHERE token = ?", inviteToken)
	if err != nil {
		return err
	}

	// Insert or ignore to voide duplicate connections (alongside the unique constraint in the table)
	_, err = m.db.Exec("INSERT OR IGNORE INTO Connections (source, target) VALUES (?, ?), (?, ?)", currentUser, handle, handle, currentUser)
	if err != nil {
		return err
	}

	return nil
}
