package cliproxy

import (
	"context"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/quotahistory"
	log "github.com/sirupsen/logrus"
)

func (s *Service) startQuotaHistory(ctx context.Context) {
	dir := quotahistory.Dir(s.configPath)
	if dir == "" || s.coreManager == nil {
		return
	}
	recorder := quotahistory.NewRecorder(s.coreManager, quotahistory.NewStore(dir), s.globalProxyURL)
	go recorder.Run(ctx)
	log.Infof("quota history recorder started (interval=%s, dir=%s)", quotahistory.Interval, dir)
}

func (s *Service) globalProxyURL() string {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	if s.cfg == nil {
		return ""
	}
	return s.cfg.ProxyURL
}
