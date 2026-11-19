package antigravity

import (
	"strings"

	cliproxyauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func ConfiguredUserAgent(account *cliproxyauth.Auth) string {
	return accountSetting(account, "user_agent")
}

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
