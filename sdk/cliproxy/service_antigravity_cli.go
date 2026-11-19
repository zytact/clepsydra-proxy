package cliproxy

import (
	"context"
	"errors"
	"strings"
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
	antigravityCLILoginSource  = "antigravity-cli"
	antigravityCLIPollInterval = time.Minute
)

var (
	readAntigravityCLILogin   = antigravity.ReadCLILogin
	prepareAntigravityCLIAuth = func(ctx context.Context, cfg *config.Config, auth *coreauth.Auth) (*coreauth.Auth, error) {
		return executor.NewAntigravityExecutor(cfg).PrepareRequestAuth(ctx, auth)
	}
)

func (s *Service) startAntigravityCLILogin(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(antigravityCLIPollInterval)
		defer ticker.Stop()
		for {
			s.syncAntigravityCLILogin(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (s *Service) syncAntigravityCLILogin(ctx context.Context) {
	if s == nil || s.coreManager == nil || ctx.Err() != nil {
		return
	}
	existing, exists := s.coreManager.GetByID(antigravityCLIAuthID)
	if !s.antigravityCLILoginEnabled() {
		s.removeAntigravityCLILogin(ctx, exists)
		return
	}
	login, errRead := readAntigravityCLILogin()
	if errors.Is(errRead, antigravity.ErrCLILoginNotFound) {
		s.removeAntigravityCLILogin(ctx, exists)
		return
	}
	if errRead != nil {
		log.WithError(errRead).Warn("antigravity cli login: read failed")
		return
	}
	unchanged := exists && existing != nil && sameAntigravityCLILogin(existing, login)
	auth := newAntigravityCLIAuth(login, time.Now())
	if unchanged {
		if metadataString(existing.Metadata, "project_id") != "" {
			return
		}
		auth = existing.Clone()
	}
	auth, projectFound := s.ensureAntigravityCLIProject(ctx, auth)
	if !projectFound && unchanged {
		return
	}
	action := watcher.AuthUpdateActionAdd
	if exists {
		action = watcher.AuthUpdateActionModify
	}
	s.emitAuthUpdate(ctx, watcher.AuthUpdate{Action: action, ID: auth.ID, Auth: auth})
}

func (s *Service) removeAntigravityCLILogin(ctx context.Context, exists bool) {
	if !exists {
		return
	}
	s.emitAuthUpdate(ctx, watcher.AuthUpdate{Action: watcher.AuthUpdateActionDelete, ID: antigravityCLIAuthID})
}

func (s *Service) antigravityCLILoginEnabled() bool {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.cfg == nil || s.cfg.Antigravity.CLILoginQuotaEnabled()
}

func (s *Service) ensureAntigravityCLIProject(ctx context.Context, auth *coreauth.Auth) (*coreauth.Auth, bool) {
	if metadataString(auth.Metadata, "project_id") != "" {
		return auth, true
	}
	s.cfgMu.RLock()
	cfg := s.cfg
	s.cfgMu.RUnlock()
	prepared, errPrepare := prepareAntigravityCLIAuth(ctx, cfg, auth)
	if errPrepare != nil {
		log.WithError(errPrepare).Warn("antigravity cli login: project discovery failed")
		return auth, false
	}
	if prepared == nil {
		return auth, false
	}
	return prepared, metadataString(prepared.Metadata, "project_id") != ""
}

func sameAntigravityCLILogin(auth *coreauth.Auth, login *antigravity.CLILogin) bool {
	return metadataString(auth.Metadata, "refresh_token") == login.RefreshToken &&
		metadataString(auth.Metadata, "email") == login.Email
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
		ID:       antigravityCLIAuthID,
		Provider: "antigravity",
		Label:    label,
		Status:   coreauth.StatusActive,
		Attributes: map[string]string{
			coreauth.AttributeRuntimeOnly: "true",
			coreauth.AttributeQuotaOnly:   "true",
			coreauth.AttributeLoginSource: antigravityCLILoginSource,
		},
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func metadataString(metadata map[string]any, key string) string {
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}
