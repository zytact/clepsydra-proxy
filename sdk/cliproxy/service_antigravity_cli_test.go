package cliproxy

import (
	"context"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/antigravity"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

func init() {
	readAntigravityCLILogin = func() (*antigravity.CLILogin, error) { return nil, antigravity.ErrCLILoginNotFound }
}

func stubAntigravityCLILogin(t *testing.T, login **antigravity.CLILogin) *int {
	t.Helper()
	prepares := 0
	previousRead, previousPrepare := readAntigravityCLILogin, prepareAntigravityCLIAuth
	readAntigravityCLILogin = func() (*antigravity.CLILogin, error) {
		if *login == nil {
			return nil, antigravity.ErrCLILoginNotFound
		}
		copied := **login
		return &copied, nil
	}
	prepareAntigravityCLIAuth = func(_ context.Context, _ *config.Config, auth *coreauth.Auth) (*coreauth.Auth, error) {
		prepares++
		prepared := auth.Clone()
		prepared.Metadata["project_id"] = "project-1"
		return prepared, nil
	}
	t.Cleanup(func() { readAntigravityCLILogin, prepareAntigravityCLIAuth = previousRead, previousPrepare })
	return &prepares
}

func TestSyncAntigravityCLILoginTracksTheCLI(t *testing.T) {
	login := &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh-1", Email: "cli@example.com", Expiry: time.Now().Add(time.Hour)}
	prepares := stubAntigravityCLILogin(t, &login)
	svc := &Service{cfg: &config.Config{}, coreManager: coreauth.NewManager(nil, nil, nil)}
	ctx := context.Background()

	svc.syncAntigravityCLILogin(ctx)
	auth, ok := svc.coreManager.GetByID(antigravityCLIAuthID)
	if !ok || !auth.QuotaOnly || auth.Provider != "antigravity" || auth.Metadata["project_id"] != "project-1" || auth.Metadata["email"] != "cli@example.com" {
		t.Fatalf("registered auth = %+v", auth)
	}
	if models := registry.GetGlobalRegistry().GetModelsForClient(antigravityCLIAuthID); len(models) != 0 {
		t.Fatalf("quota-only auth registered models %v", models)
	}

	svc.syncAntigravityCLILogin(ctx)
	if *prepares != 1 {
		t.Fatalf("unchanged login re-ran project discovery %d times", *prepares)
	}

	login.RefreshToken = "refresh-2"
	svc.syncAntigravityCLILogin(ctx)
	if auth, _ = svc.coreManager.GetByID(antigravityCLIAuthID); auth.Metadata["refresh_token"] != "refresh-2" {
		t.Fatalf("new CLI login not picked up: %v", auth.Metadata["refresh_token"])
	}

	login = nil
	svc.syncAntigravityCLILogin(ctx)
	if _, ok = svc.coreManager.GetByID(antigravityCLIAuthID); ok {
		t.Fatal("auth kept after the CLI signed out")
	}
}

func TestSyncAntigravityCLILoginHonorsConfigToggle(t *testing.T) {
	login := &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh", Email: "cli@example.com"}
	stubAntigravityCLILogin(t, &login)
	disabled := false
	svc := &Service{cfg: &config.Config{Antigravity: config.AntigravityConfig{CLILoginQuota: &disabled}}, coreManager: coreauth.NewManager(nil, nil, nil)}

	svc.syncAntigravityCLILogin(context.Background())
	if _, ok := svc.coreManager.GetByID(antigravityCLIAuthID); ok {
		t.Fatal("auth registered while cli-login-quota is false")
	}
}
