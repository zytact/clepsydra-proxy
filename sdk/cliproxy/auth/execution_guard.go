package auth

import (
	"context"
	"net/http"

	cliproxyexecutor "github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy/executor"
)

func refuseQuotaOnly(auth *Auth) error {
	if auth == nil || !auth.QuotaOnly {
		return nil
	}
	return &Error{Code: "auth_quota_only", Message: "credential is quota-only and never serves requests", HTTPStatus: http.StatusConflict}
}

func guardedExecute(ctx context.Context, executor ProviderExecutor, auth *Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	if errRefuse := refuseQuotaOnly(auth); errRefuse != nil {
		return cliproxyexecutor.Response{}, errRefuse
	}
	return executor.Execute(ctx, auth, req, opts)
}

func guardedExecuteStream(ctx context.Context, executor ProviderExecutor, auth *Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (*cliproxyexecutor.StreamResult, error) {
	if errRefuse := refuseQuotaOnly(auth); errRefuse != nil {
		return nil, errRefuse
	}
	return executor.ExecuteStream(ctx, auth, req, opts)
}

func guardedCountTokens(ctx context.Context, executor ProviderExecutor, auth *Auth, req cliproxyexecutor.Request, opts cliproxyexecutor.Options) (cliproxyexecutor.Response, error) {
	if errRefuse := refuseQuotaOnly(auth); errRefuse != nil {
		return cliproxyexecutor.Response{}, errRefuse
	}
	return executor.CountTokens(ctx, auth, req, opts)
}

func guardedHttpRequest(ctx context.Context, executor ProviderExecutor, auth *Auth, req *http.Request) (*http.Response, error) {
	if errRefuse := refuseQuotaOnly(auth); errRefuse != nil {
		return nil, errRefuse
	}
	return executor.HttpRequest(ctx, auth, req)
}

func guardedPrepareRequest(preparer RequestPreparer, req *http.Request, auth *Auth) error {
	if errRefuse := refuseQuotaOnly(auth); errRefuse != nil {
		return errRefuse
	}
	return preparer.PrepareRequest(req, auth)
}
