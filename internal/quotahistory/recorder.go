package quotahistory

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MoeForge/moego/json"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/auth/antigravity"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/misc"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/proxyutil"
	log "github.com/sirupsen/logrus"
)

const (
	Interval       = 5 * time.Minute
	firstPollDelay = 30 * time.Second
	requestTimeout = 20 * time.Second
	maxBodyBytes   = 1 << 20

	claudeUsageURL       = "https://api.anthropic.com/api/oauth/usage"
	codexUsageURL        = "https://chatgpt.com/backend-api/wham/usage"
	antigravityQuotaPath = "/" + antigravity.APIVersion + ":retrieveUserQuotaSummary"
)

type usageSource struct {
	method  string
	urls    func(account *auth.Auth) []string
	body    func(account *auth.Auth) []byte
	headers func(account *auth.Auth, header http.Header)
	parse   func(body []byte, now time.Time) ([]Window, error)
}

func defaultSources() map[string]usageSource {
	return map[string]usageSource{
		"claude": {
			method: http.MethodGet,
			urls:   fixedURLs(claudeUsageURL),
			headers: func(_ *auth.Auth, header http.Header) {
				header.Set("anthropic-beta", "oauth-2025-04-20")
				header.Set("User-Agent", "claude-cli/2.1.280 (external, cli)")
			},
			parse: func(body []byte, _ time.Time) ([]Window, error) { return parseClaudeUsage(body) },
		},
		"codex": {
			method: http.MethodGet,
			urls:   fixedURLs(codexUsageURL),
			headers: func(account *auth.Auth, header http.Header) {
				header.Set("User-Agent", "codex_cli_rs")
				if accountID := metadataString(account, "account_id"); accountID != "" {
					header.Set("Chatgpt-Account-Id", accountID)
				}
			},
			parse: parseCodexUsage,
		},
		"antigravity": {
			method: http.MethodPost,
			urls:   antigravityQuotaURLs,
			body:   antigravityQuotaBody,
			headers: func(account *auth.Auth, header http.Header) {
				header.Set("Content-Type", "application/json")
				header.Set("User-Agent", misc.AntigravityRequestUserAgent(antigravity.ConfiguredUserAgent(account)))
			},
			parse: func(body []byte, _ time.Time) ([]Window, error) { return parseAntigravityQuota(body) },
		},
	}
}

func antigravityQuotaBody(account *auth.Auth) []byte {
	body, _ := json.Marshal(struct {
		Project string `json:"project,omitempty"`
	}{Project: metadataString(account, "project_id")})
	return body
}

func fixedURLs(urls ...string) func(*auth.Auth) []string {
	return func(*auth.Auth) []string { return urls }
}

func antigravityQuotaURLs(account *auth.Auth) []string {
	if base := antigravity.CustomBaseURL(account); base != "" {
		return []string{base + antigravityQuotaPath}
	}
	return []string{
		antigravity.DailyAPIEndpoint + antigravityQuotaPath,
		antigravity.APIEndpoint + antigravityQuotaPath,
	}
}

type statusError struct {
	status int
}

func (e statusError) Error() string {
	return fmt.Sprintf("usage request returned status %d", e.status)
}

func isAuthError(err error) bool {
	var status statusError
	return errors.As(err, &status) && (status.status == http.StatusUnauthorized || status.status == http.StatusForbidden)
}

type AuthLister interface {
	List() []*auth.Auth
}

type Recorder struct {
	auths       AuthLister
	store       *Store
	globalProxy func() string
	now         func() time.Time
	sources     map[string]usageSource
}

func NewRecorder(auths AuthLister, store *Store, globalProxy func() string) *Recorder {
	return &Recorder{
		auths:       auths,
		store:       store,
		globalProxy: globalProxy,
		now:         time.Now,
		sources:     defaultSources(),
	}
}

func (r *Recorder) Run(ctx context.Context) {
	timer := time.NewTimer(firstPollDelay)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			r.Poll(ctx)
			timer.Reset(Interval)
		}
	}
}

func (r *Recorder) Poll(ctx context.Context) {
	readings := []Reading{}
	for _, account := range r.auths.List() {
		if account == nil || account.Disabled {
			continue
		}
		source, ok := r.sources[account.Provider]
		if !ok {
			continue
		}
		readings = append(readings, r.read(ctx, account, source))
	}
	if ctx.Err() != nil {
		return
	}
	if len(readings) > 0 {
		if errAppend := r.store.Append(readings...); errAppend != nil {
			log.WithError(errAppend).Warn("quota history: failed to store readings")
		}
	}
	if errPrune := r.store.Prune(r.now().Add(-Retention)); errPrune != nil {
		log.WithError(errPrune).Warn("quota history: failed to prune old readings")
	}
}

func (r *Recorder) read(ctx context.Context, account *auth.Auth, source usageSource) Reading {
	reading := Reading{
		At:        r.now().UTC(),
		Provider:  account.Provider,
		AuthID:    account.ID,
		AuthIndex: account.EnsureIndex(),
		Email:     metadataString(account, "email"),
	}
	windows, errRead := r.readWindows(ctx, account, source, reading.At)
	if errRead != nil {
		reading.Error = errRead.Error()
		return reading
	}
	reading.Windows = windows
	return reading
}

func (r *Recorder) readWindows(ctx context.Context, account *auth.Auth, source usageSource, at time.Time) ([]Window, error) {
	token := account.CurrentAccessToken()
	if token == "" {
		return nil, fmt.Errorf("no access token")
	}
	if expiry, ok := account.AccessTokenExpirationTime(); ok && !expiry.IsZero() && !expiry.After(r.now()) {
		return nil, fmt.Errorf("access token expired")
	}
	var payload []byte
	if source.body != nil {
		payload = source.body(account)
	}
	var errFirst error
	for _, url := range source.urls(account) {
		windows, errURL := r.readURL(ctx, account, source, url, token, payload, at)
		if errURL == nil {
			return windows, nil
		}
		if errFirst == nil {
			errFirst = errURL
		}
		if isAuthError(errURL) {
			return nil, errURL
		}
	}
	return nil, errFirst
}

func (r *Recorder) readURL(ctx context.Context, account *auth.Auth, source usageSource, url, token string, payload []byte, at time.Time) ([]Window, error) {
	body, errFetch := r.fetchURL(ctx, account, source, url, token, payload)
	if errFetch != nil {
		return nil, errFetch
	}
	windows, errParse := source.parse(body, at)
	if errParse != nil {
		return nil, errors.New("unexpected usage response")
	}
	return windows, nil
}

func (r *Recorder) fetchURL(ctx context.Context, account *auth.Auth, source usageSource, url, token string, payload []byte) ([]byte, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	var requestBody io.Reader
	if payload != nil {
		requestBody = bytes.NewReader(payload)
	}
	request, errRequest := http.NewRequestWithContext(requestCtx, source.method, url, requestBody)
	if errRequest != nil {
		return nil, fmt.Errorf("build usage request: %w", errRequest)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Accept", "application/json")
	source.headers(account, request.Header)
	transport := r.transport(account)
	defer transport.CloseIdleConnections()
	response, errDo := (&http.Client{Transport: transport}).Do(request)
	if errDo != nil {
		return nil, fmt.Errorf("usage request failed: %w", errDo)
	}
	defer func() {
		if errClose := response.Body.Close(); errClose != nil {
			log.WithError(errClose).Debug("quota history: close usage response")
		}
	}()
	body, errRead := io.ReadAll(io.LimitReader(response.Body, maxBodyBytes))
	if errRead != nil {
		return nil, fmt.Errorf("read usage response: %w", errRead)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, statusError{status: response.StatusCode}
	}
	return body, nil
}

func (r *Recorder) transport(account *auth.Auth) *http.Transport {
	for _, candidate := range []string{account.ProxyURL, r.globalProxy()} {
		if strings.TrimSpace(candidate) == "" {
			continue
		}
		if transport, _, errBuild := proxyutil.BuildHTTPTransport(candidate); errBuild == nil && transport != nil {
			return transport
		}
	}
	return proxyutil.NewDirectTransport()
}

func metadataString(account *auth.Auth, key string) string {
	if account.Metadata == nil {
		return ""
	}
	value, _ := account.Metadata[key].(string)
	return strings.TrimSpace(value)
}
