package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Connection struct {
	Name            string `json:"name"`
	User            string `json:"user"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
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

func LoadConnections() ([]Connection, error) {
	var connections []Connection
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return connections, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return connections, nil
	}

	err = json.Unmarshal(data, &connections)
	return connections, err
}

func SaveConnections(connections []Connection) error {
	data, err := json.MarshalIndent(connections, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, data, 0644)
}

func AddConnection(conn Connection) error {
	connections, err := LoadConnections()
	if err != nil {
		return err
	}

	for _, c := range connections {
		if c.Name == conn.Name {
			return fmt.Errorf("connection with name '%s' already exists", conn.Name)
		}
	}

	connections = append(connections, conn)
	return SaveConnections(connections)
}

func RemoveConnection(name string) error {
	connections, err := LoadConnections()
	if err != nil {
		return err
	}

	var newConnections []Connection
	var found bool
	for _, c := range connections {
		if c.Name == name {
			found = true
			continue
		}
		newConnections = append(newConnections, c)
	}

	if !found {
		return fmt.Errorf("connection with name '%s' not found", name)
	}

	return SaveConnections(newConnections)
}
