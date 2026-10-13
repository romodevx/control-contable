package main

import (
	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
)

// api is the single gen.StrictServerInterface, built by embedding one handler
// per module (ADR-0001). Each module adds its handler struct here.
type api struct {
	notBuilt
	*health.Handler
}

// notBuilt stands in for the operations whose module is not built yet. Its
// interface is nil, so calling one panics and middleware.Recover answers the
// 500 envelope. Flagged modules never get that far (their flags are OFF);
// the platform operation listEnabledFlags arrives with P0-15. Being embedded
// one level deeper than the module handlers, it never shadows them.
type notBuilt struct {
	gen.StrictServerInterface
}
