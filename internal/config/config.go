package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigStore struct {
	Version     string       `json:"version"`
	Connections []Connection `json:"connections"`
	Groups      []Group      `json:"groups"`
}

type Group struct {
	Name            string   `json:"name"`
	Connections     []string `json:"connections"` // List of connection names
	User            string   `json:"user,omitempty"`
	Port            int      `json:"port,omitempty"`
	KeyPath         string   `json:"key_path,omitempty"`
	CredentialAlias string   `json:"credential_alias,omitempty"`
}

type Connection struct {
	Name            string `json:"name"`
	User            string `json:"user,omitempty"`
	Host            string `json:"host"`
	Port            int    `json:"port,omitempty"`
	KeyPath         string `json:"key_path,omitempty"`
	CredentialAlias string `json:"credential_alias,omitempty"`
}

var configFilePath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".config", "gossh")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}
	configFilePath = filepath.Join(configDir, "config.json")
}

func LoadStore() (ConfigStore, error) {
	var store ConfigStore
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ConfigStore{Version: "0.1.0", Connections: []Connection{}, Groups: []Group{}}, nil
		}
		return store, err
	}
	if len(data) == 0 {
		return ConfigStore{Version: "0.1.0", Connections: []Connection{}, Groups: []Group{}}, nil
	}

	err = json.Unmarshal(data, &store)
	if err != nil {
		return store, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Basic safety check for empty slices
	if store.Connections == nil {
		store.Connections = []Connection{}
	}
	if store.Groups == nil {
		store.Groups = []Group{}
	}

	return store, nil
}

func SaveStore(store ConfigStore) error {
	store.Version = "0.1.0"
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0644)
}

// Helper wrappers for existing code logic

func LoadConnections() ([]Connection, error) {
	store, err := LoadStore()
	if err != nil {
		return nil, err
	}
	return store.Connections, nil
}

func SaveConnections(connections []Connection) error {
	store, err := LoadStore()
	if err != nil {
		return err
	}
	store.Connections = connections
	return SaveStore(store)
}

func AddConnection(conn Connection) error {
	store, err := LoadStore()
	if err != nil {
		return err
	}

	for _, c := range store.Connections {
		if c.Name == conn.Name {
			return fmt.Errorf("connection with name '%s' already exists", conn.Name)
		}
	}

	store.Connections = append(store.Connections, conn)
	return SaveStore(store)
}

func RemoveConnection(name string) error {
	store, err := LoadStore()
	if err != nil {
		return err
	}

	var newConnections []Connection
	var found bool
	for _, c := range store.Connections {
		if c.Name == name {
			found = true
			continue
		}
		newConnections = append(newConnections, c)
	}

	if !found {
		return fmt.Errorf("connection with name '%s' not found", name)
	}
	store.Connections = newConnections

	// Cleanup group memberships
	for i := range store.Groups {
		var newGroupConns []string
		for _, connName := range store.Groups[i].Connections {
			if connName != name {
				newGroupConns = append(newGroupConns, connName)
			}
		}
		store.Groups[i].Connections = newGroupConns
	}

	return SaveStore(store)
}

func ResolveConnection(name string) (*Connection, error) {
	store, err := LoadStore()
	if err != nil {
		return nil, err
	}

	var conn *Connection
	for _, c := range store.Connections {
		if c.Name == name {
			connCopy := c
			conn = &connCopy
			break
		}
	}

	if conn == nil {
		return nil, fmt.Errorf("connection '%s' not found", name)
	}

	// Search for all groups this connection belongs to
	for _, g := range store.Groups {
		for _, cName := range g.Connections {
			if cName == name {
				// Apply group-level defaults if connection field is empty
				if conn.User == "" && g.User != "" {
					conn.User = g.User
				}
				if conn.Port == 0 && g.Port != 0 {
					conn.Port = g.Port
				}
				if conn.KeyPath == "" && g.KeyPath != "" {
					conn.KeyPath = g.KeyPath
				}
				if conn.CredentialAlias == "" && g.CredentialAlias != "" {
					conn.CredentialAlias = g.CredentialAlias
				}
				break
			}
		}
	}

	if conn.Port == 0 {
		conn.Port = 22
	}

	return conn, nil
}
