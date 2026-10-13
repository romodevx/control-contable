package main

import (
	"github.com/romodevx/control-contable/backend/adapter/http/flagshttp"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
)

// api is the single gen.StrictServerInterface, built by embedding one handler
// per module (ADR-0001). Each module adds its handler struct here and deletes
// the stubs of the operations it builds from notbuilt.go.
type api struct {
	notBuilt
	*health.Handler
	*flagsHandler
	*authHandler
	*authhttp.SessionHandler
}

// flagsHandler and authHandler name the embedded handlers: health.Handler,
// flagshttp.Handler and authhttp.Handler would otherwise all be the embedded
// field "Handler".
type (
	flagsHandler = flagshttp.Handler
	authHandler  = authhttp.Handler
)
