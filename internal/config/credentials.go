package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Credential struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
}

var credentialsFilePath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user home directory:", err)
		os.Exit(1)
	}
	configDir := filepath.Join(home, ".config", "gossh")
	credentialsFilePath = filepath.Join(configDir, "credentials.json")
}

func LoadCredentials() ([]Credential, error) {
	var credentials []Credential
	data, err := os.ReadFile(credentialsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return credentials, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return credentials, nil
	}

	err = json.Unmarshal(data, &credentials)
	if err != nil {
		return nil, err
	}

	for i := range credentials {
		if credentials[i].Password != "" {
			decrypted, err := Decrypt(credentials[i].Password)
			if err != nil {
				return nil, fmt.Errorf("could not decrypt password for %s: %w", credentials[i].Alias, err)
			}
			credentials[i].Password = decrypted
		}
	}
	return credentials, nil
}

func SaveCredentials(credentials []Credential) error {
	credsToSave := make([]Credential, len(credentials))
	copy(credsToSave, credentials)

	for i := range credsToSave {
		if credsToSave[i].Password != "" {
			encrypted, err := Encrypt(credsToSave[i].Password)
			if err != nil {
				return fmt.Errorf("could not encrypt password for %s: %w", credsToSave[i].Alias, err)
			}
			credsToSave[i].Password = encrypted
		}
	}

	data, err := json.MarshalIndent(credsToSave, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(credentialsFilePath, data, 0600)
}

func AddCredential(cred Credential) error {
	credentials, err := LoadCredentials()
	if err != nil {
		return err
	}

	for _, c := range credentials {
		if c.Alias == cred.Alias {
			return fmt.Errorf("credential with alias '%s' already exists", cred.Alias)
		}
	}

	credentials = append(credentials, cred)
	return SaveCredentials(credentials)
}

func RemoveCredential(alias string) error {
	credentials, err := LoadCredentials()
	if err != nil {
		return err
	}

	var newCredentials []Credential
	var found bool
	for _, c := range credentials {
		if c.Alias == alias {
			found = true
			continue
		}
		newCredentials = append(newCredentials, c)
	}

	if !found {
		return fmt.Errorf("credential with alias '%s' not found", alias)
	}

	return SaveCredentials(newCredentials)
}
