package cliproxy

import (
	"context"
	"errors"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/antigravity"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/runtime/executor"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/watcher"
	coreauth "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	log "github.com/sirupsen/logrus"
)

const (
	antigravityCLIAuthID       = "antigravity-cli"
	antigravityCLILoginOrigin  = "antigravity-cli"
	antigravityCLIPollInterval = time.Minute
)

type antigravityCLILoginSource struct {
	read     func() (*antigravity.CLILogin, error)
	discover func(ctx context.Context, cfg *config.Config, auth *coreauth.Auth) (*coreauth.Auth, error)
}

func defaultAntigravityCLILoginSource() antigravityCLILoginSource {
	return antigravityCLILoginSource{
		read: antigravity.ReadCLILogin,
		discover: func(ctx context.Context, cfg *config.Config, auth *coreauth.Auth) (*coreauth.Auth, error) {
			return executor.NewAntigravityExecutor(cfg).PrepareRequestAuth(ctx, auth)
		},
	}
}

type antigravityCLILoginLoop struct {
	service *Service
	source  antigravityCLILoginSource
}

func (s *Service) startAntigravityCLILogin(ctx context.Context) {
	if s.antigravityCLILogin.read == nil {
		return
	}
	loop := &antigravityCLILoginLoop{service: s, source: s.antigravityCLILogin}
	go func() {
		ticker := time.NewTicker(antigravityCLIPollInterval)
		defer ticker.Stop()
		for {
			loop.sync(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (l *antigravityCLILoginLoop) sync(ctx context.Context) {
	s := l.service
	if s == nil || s.coreManager == nil || ctx.Err() != nil {
		return
	}
	existing, exists := s.coreManager.GetByID(antigravityCLIAuthID)
	if !s.antigravityCLILoginEnabled() {
		l.remove(ctx, exists)
		return
	}
	login, errRead := l.source.read()
	if errors.Is(errRead, antigravity.ErrCLILoginNotFound) {
		l.remove(ctx, exists)
		return
	}
	if errRead != nil {
		log.WithError(errRead).Warn("antigravity cli login: read failed")
		return
	}
	unchanged := exists && existing != nil && sameAntigravityCLILogin(existing, login)
	auth := newAntigravityCLIAuth(login, time.Now())
	if unchanged {
		if existing.MetadataString("project_id") != "" {
			return
		}
		auth = existing.Clone()
	}
	auth, projectFound := l.ensureProject(ctx, auth)
	if !projectFound && unchanged {
		return
	}
	action := watcher.AuthUpdateActionAdd
	if exists {
		action = watcher.AuthUpdateActionModify
	}
	s.emitAuthUpdate(ctx, watcher.AuthUpdate{Action: action, ID: auth.ID, Auth: auth})
}

func (l *antigravityCLILoginLoop) remove(ctx context.Context, exists bool) {
	if !exists {
		return
	}
	l.service.emitAuthUpdate(ctx, watcher.AuthUpdate{Action: watcher.AuthUpdateActionDelete, ID: antigravityCLIAuthID})
}

func (s *Service) antigravityCLILoginEnabled() bool {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.cfg == nil || s.cfg.Antigravity.CLILoginQuotaEnabled()
}

func (l *antigravityCLILoginLoop) ensureProject(ctx context.Context, auth *coreauth.Auth) (*coreauth.Auth, bool) {
	if auth.MetadataString("project_id") != "" {
		return auth, true
	}
	l.service.cfgMu.RLock()
	cfg := l.service.cfg
	l.service.cfgMu.RUnlock()
	prepared, errPrepare := l.source.discover(ctx, cfg, auth)
	if errPrepare != nil {
		log.WithError(errPrepare).Warn("antigravity cli login: project discovery failed")
		return auth, false
	}
	if prepared == nil {
		return auth, false
	}
	return prepared, prepared.MetadataString("project_id") != ""
}

func sameAntigravityCLILogin(auth *coreauth.Auth, login *antigravity.CLILogin) bool {
	return auth.MetadataString("refresh_token") == login.RefreshToken && auth.MetadataString("email") == login.Email
}

func newAntigravityCLIAuth(login *antigravity.CLILogin, now time.Time) *coreauth.Auth {
	metadata := map[string]any{
		"type":          "antigravity",
		"access_token":  login.AccessToken,
		"refresh_token": login.RefreshToken,
	}
	if login.TokenType != "" {
		metadata["token_type"] = login.TokenType
	}
	if !login.Expiry.IsZero() {
		metadata["expired"] = login.Expiry.UTC().Format(time.RFC3339)
	}
	label := "antigravity"
	if login.Email != "" {
		metadata["email"] = login.Email
		label = login.Email
	}
	return &coreauth.Auth{
		ID:        antigravityCLIAuthID,
		Provider:  "antigravity",
		Label:     label,
		Status:    coreauth.StatusActive,
		QuotaOnly: true,
		Attributes: map[string]string{
			coreauth.AttributeRuntimeOnly: "true",
			coreauth.AttributeLoginSource: antigravityCLILoginOrigin,
		},
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
