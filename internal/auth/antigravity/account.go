package antigravity

import (
	"strings"

	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

// ConfiguredUserAgent returns the user agent an account overrides, or "" when it keeps the default.
func ConfiguredUserAgent(account *cliproxyauth.Auth) string {
	return accountSetting(account, "user_agent")
}

// CustomBaseURL returns the base URL an account pins its requests to, or "" when it uses the default hosts.
func CustomBaseURL(account *cliproxyauth.Auth) string {
	return strings.TrimSuffix(accountSetting(account, "base_url"), "/")
}

func accountSetting(account *cliproxyauth.Auth, key string) string {
	if account == nil {
		return ""
	}
	if value := strings.TrimSpace(account.Attributes[key]); value != "" {
		return value
	}
	value, _ := account.Metadata[key].(string)
	return strings.TrimSpace(value)
}
