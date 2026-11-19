package antigravity

import (
	"testing"
	"time"
)

func TestParseCLILogin(t *testing.T) {
	raw := []byte(`{"token":{"access_token":"ya29.access","token_type":"Bearer","refresh_token":"1//refresh","expiry":"2026-11-19T20:59:51.844721+05:30","expires_in":3599},"auth_method":"consumer","id_token_email":"user@example.com"}`)
	login, errParse := ParseCLILogin(raw)
	if errParse != nil {
		t.Fatal(errParse)
	}
	wantExpiry := time.Date(2026, 11, 19, 15, 29, 51, 844721000, time.UTC)
	if login.AccessToken != "ya29.access" || login.RefreshToken != "1//refresh" || login.TokenType != "Bearer" || login.Email != "user@example.com" || !login.Expiry.Equal(wantExpiry) {
		t.Fatalf("ParseCLILogin = %+v", login)
	}
}

func TestParseCLILoginRejectsUnusablePayloads(t *testing.T) {
	for name, raw := range map[string]string{
		"not json":        `nope`,
		"no refresh":      `{"token":{"access_token":"a"}}`,
		"bad expiry":      `{"token":{"refresh_token":"r","expiry":"tomorrow"}}`,
		"empty object":    `{}`,
		"token not found": `{"id_token_email":"user@example.com"}`,
	} {
		if _, errParse := ParseCLILogin([]byte(raw)); errParse == nil {
			t.Errorf("%s: ParseCLILogin succeeded", name)
		}
	}
}
