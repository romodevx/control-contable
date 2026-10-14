package main

import (
	"github.com/romodevx/control-contable/backend/adapter/http/flagshttp"
	"github.com/romodevx/control-contable/backend/adapter/http/health"
	authhttp "github.com/romodevx/control-contable/backend/internal/auth/adapter/http"
	expenseshttp "github.com/romodevx/control-contable/backend/internal/expenses/adapter/http"
	taxenginehttp "github.com/romodevx/control-contable/backend/internal/taxengine/adapter/http"
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
)

// api is the single gen.StrictServerInterface, built by embedding one handler
// per module (ADR-0001). Each module adds its handler struct here and deletes
// the stubs of the operations it builds from notbuilt.go.
type api struct {
	notBuilt
	*health.Handler
	*flagsHandler
	*authHandler
	*accountHandler
	*authhttp.SessionHandler
	*taxpayersHandler
	*expensesHandler
	*taxEngineHandler
}

// flagsHandler and authHandler name the embedded handlers: health.Handler,
// flagshttp.Handler and authhttp.Handler would otherwise all be the embedded
// field "Handler". accountHandler is the auth module's account side (M1-09);
// taxpayersHandler serves module_taxpayers (M2-07), expensesHandler
// module_expenses (M4-07) and taxEngineHandler the built module_tax_engine
// operations (M5-21).
type (
	flagsHandler     = flagshttp.Handler
	authHandler      = authhttp.Handler
	accountHandler   = authhttp.AccountHandler
	taxpayersHandler = taxpayershttp.Handler
	expensesHandler  = expenseshttp.Handler
	taxEngineHandler = taxenginehttp.Handler
)
