package antigravity

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MoeForge/moego/json"
)

var ErrCLILoginNotFound = errors.New("antigravity cli login not found")

type CLILogin struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       time.Time
	Email        string
}

type cliLoginPayload struct {
	Token struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		RefreshToken string `json:"refresh_token"`
		Expiry       string `json:"expiry"`
	} `json:"token"`
	IDTokenEmail string `json:"id_token_email"`
}

func ReadCLILogin() (*CLILogin, error) {
	raw, errRead := readCLILoginPayload()
	if errRead != nil {
		return nil, errRead
	}
	return ParseCLILogin(raw)
}

func ParseCLILogin(raw []byte) (*CLILogin, error) {
	var payload cliLoginPayload
	if errUnmarshal := json.Unmarshal(raw, &payload); errUnmarshal != nil {
		return nil, fmt.Errorf("parse antigravity cli login: %w", errUnmarshal)
	}
	login := &CLILogin{
		AccessToken:  strings.TrimSpace(payload.Token.AccessToken),
		RefreshToken: strings.TrimSpace(payload.Token.RefreshToken),
		TokenType:    strings.TrimSpace(payload.Token.TokenType),
		Email:        strings.TrimSpace(payload.IDTokenEmail),
	}
	if login.RefreshToken == "" {
		return nil, errors.New("parse antigravity cli login: refresh token missing")
	}
	if expiry := strings.TrimSpace(payload.Token.Expiry); expiry != "" {
		parsed, errParse := time.Parse(time.RFC3339Nano, expiry)
		if errParse != nil {
			return nil, fmt.Errorf("parse antigravity cli login expiry: %w", errParse)
		}
		login.Expiry = parsed
	}
	return login, nil
}
