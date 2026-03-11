package usage

import (
	"encoding/json"
	"fmt"
	"os"
)

type credentials struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

// ReadToken reads the OAuth access token from the given credentials file path.
// Returns an error if the file is missing, malformed, or the token is empty.
func ReadToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading credentials: %w", err)
	}
	var creds credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", fmt.Errorf("parsing credentials: %w", err)
	}
	if creds.ClaudeAiOauth.AccessToken == "" {
		return "", fmt.Errorf("no accessToken in credentials")
	}
	return creds.ClaudeAiOauth.AccessToken, nil
}
