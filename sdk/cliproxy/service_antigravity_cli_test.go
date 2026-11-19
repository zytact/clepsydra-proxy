package cliproxy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/antigravity"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/watcher"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
)

type fakeAntigravityCLI struct {
	login     *antigravity.CLILogin
	readErr   error
	discovers int
}

func (f *fakeAntigravityCLI) source() antigravityCLILoginSource {
	return antigravityCLILoginSource{
		read: func() (*antigravity.CLILogin, error) {
			if f.readErr != nil {
				return nil, f.readErr
			}
			if f.login == nil {
				return nil, antigravity.ErrCLILoginNotFound
			}
			copied := *f.login
			return &copied, nil
		},
		discover: func(_ context.Context, _ *config.Config, auth *coreauth.Auth) (*coreauth.Auth, error) {
			f.discovers++
			prepared := auth.Clone()
			prepared.Metadata["project_id"] = "project-1"
			return prepared, nil
		},
	}
}

func newAntigravityCLILoop(cfg *config.Config, fake *fakeAntigravityCLI) (*Service, *antigravityCLILoginLoop) {
	svc := &Service{cfg: cfg, coreManager: coreauth.NewManager(nil, nil, nil)}
	return svc, &antigravityCLILoginLoop{service: svc, source: fake.source()}
}

func TestAntigravityCLILoginLoopTracksTheCLI(t *testing.T) {
	fake := &fakeAntigravityCLI{login: &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh-1", Email: "cli@example.com", Expiry: time.Now().Add(time.Hour)}}
	svc, loop := newAntigravityCLILoop(&config.Config{}, fake)
	ctx := context.Background()

	loop.sync(ctx)
	auth, ok := svc.coreManager.GetByID(antigravityCLIAuthID)
	if !ok || !auth.QuotaOnly || auth.Provider != "antigravity" || auth.MetadataString("project_id") != "project-1" || auth.MetadataString("email") != "cli@example.com" {
		t.Fatalf("registered auth = %+v", auth)
	}
	if models := registry.GetGlobalRegistry().GetModelsForClient(antigravityCLIAuthID); len(models) != 0 {
		t.Fatalf("quota-only auth registered models %v", models)
	}

	loop.sync(ctx)
	if fake.discovers != 1 {
		t.Fatalf("unchanged login re-ran project discovery %d times", fake.discovers)
	}

	fake.login.RefreshToken = "refresh-2"
	loop.sync(ctx)
	if auth, _ = svc.coreManager.GetByID(antigravityCLIAuthID); auth.MetadataString("refresh_token") != "refresh-2" {
		t.Fatalf("new CLI login not picked up: %v", auth.MetadataString("refresh_token"))
	}

	fake.login = nil
	loop.sync(ctx)
	if _, ok = svc.coreManager.GetByID(antigravityCLIAuthID); ok {
		t.Fatal("auth kept after the CLI signed out")
	}
}

func TestAntigravityCLILoginLoopHonorsConfigToggle(t *testing.T) {
	disabled := false
	fake := &fakeAntigravityCLI{login: &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh", Email: "cli@example.com"}}
	svc, loop := newAntigravityCLILoop(&config.Config{Antigravity: config.AntigravityConfig{CLILoginQuota: &disabled}}, fake)

	loop.sync(context.Background())
	if _, ok := svc.coreManager.GetByID(antigravityCLIAuthID); ok {
		t.Fatal("auth registered while cli-login-quota is false")
	}
}

func TestAntigravityCLILoginLoopDropsAnUnreadableLogin(t *testing.T) {
	fake := &fakeAntigravityCLI{login: &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh", Email: "cli@example.com"}}
	svc, loop := newAntigravityCLILoop(&config.Config{}, fake)
	loop.sync(context.Background())
	if _, ok := svc.coreManager.GetByID(antigravityCLIAuthID); !ok {
		t.Fatal("auth not registered")
	}

	fake.readErr = errors.New("parse antigravity cli login: refresh token missing")
	loop.sync(context.Background())
	if _, ok := svc.coreManager.GetByID(antigravityCLIAuthID); ok {
		t.Fatal("stale auth kept after the CLI login became unreadable")
	}
	if loop.state != antigravityCLILoginUnreadable {
		t.Fatalf("state = %q, want %q", loop.state, antigravityCLILoginUnreadable)
	}
}

func TestAntigravityCLILoginBypassesTheWatcher(t *testing.T) {
	fake := &fakeAntigravityCLI{login: &antigravity.CLILogin{AccessToken: "access", RefreshToken: "refresh", Email: "cli@example.com"}}
	svc, loop := newAntigravityCLILoop(&config.Config{}, fake)
	watcherDispatches := 0
	svc.setWatcher(&WatcherWrapper{dispatchRuntimeUpdate: func(watcher.AuthUpdate) bool {
		watcherDispatches++
		return true
	}})

	loop.sync(context.Background())
	fake.login = nil
	loop.sync(context.Background())
	if watcherDispatches != 0 {
		t.Fatalf("watcher received %d CLI login updates; its rescans would replay stale tokens", watcherDispatches)
	}
}

func TestBuilderLeavesTheCLILoginOffUnlessAsked(t *testing.T) {
	plain, errPlain := NewBuilder().WithConfig(&config.Config{}).WithConfigPath(t.TempDir() + "/config.yaml").Build()
	if errPlain != nil {
		t.Fatal(errPlain)
	}
	if plain.antigravityCLILogin.read != nil {
		t.Fatal("a plain Builder wires the real Antigravity CLI login reader")
	}
	opted, errOpted := NewBuilder().WithConfig(&config.Config{}).WithConfigPath(t.TempDir() + "/config.yaml").WithAntigravityCLILogin().Build()
	if errOpted != nil {
		t.Fatal(errOpted)
	}
	if opted.antigravityCLILogin.read == nil {
		t.Fatal("WithAntigravityCLILogin did not wire the reader")
	}
}
