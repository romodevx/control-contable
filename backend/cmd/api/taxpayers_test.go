package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/config"
	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
	"github.com/romodevx/control-contable/backend/internal/platform/logging"
	taxpayershttp "github.com/romodevx/control-contable/backend/internal/taxpayers/adapter/http"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/domain"
	"github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// withTaxpayersKeys adds fixture RFC keys (hex, all one byte), distinct
// from withAuthKeys's.
func withTaxpayersKeys(env map[string]string) map[string]string {
	env["RFC_ENC_KEY"] = strings.Repeat("5b", 32)
	env["RFC_INDEX_PEPPER"] = strings.Repeat("6b", 32)
	return env
}

// testTaxpayerKeys are fixture RFC keys (all one byte each), distinct from
// testAuthKeys, never real ones.
func testTaxpayerKeys() config.Taxpayers {
	key := func(b byte) config.Secret { return config.NewSecret(strings.Repeat(string(rune(b)), 32)) }
	return config.Taxpayers{RFCKey: key(0x55), RFCKeyID: 1, RFCIndexPepper: key(0x66)}
}

func runWith(env map[string]string) error {
	return run(context.Background(), func(k string) (string, bool) { v, ok := env[k]; return v, ok }, io.Discard)
}

// PR #59 security gate, condition 1: the RFC encryption key, the RFC index
// pepper and the email index pepper are distinct secrets. The process
// refuses to start when any two are equal, in every environment.
func TestRun_refusesSharedTaxpayerSecrets(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]string)
		wantErr string
	}{
		{"index pepper is the email pepper", func(e map[string]string) { e["RFC_INDEX_PEPPER"] = e["AUTH_EMAIL_PEPPER"] },
			"RFC_INDEX_PEPPER must differ from AUTH_EMAIL_PEPPER"},
		{"encryption key is the email pepper", func(e map[string]string) { e["RFC_ENC_KEY"] = e["AUTH_EMAIL_PEPPER"] },
			"RFC_ENC_KEY must differ from AUTH_EMAIL_PEPPER"},
		{"encryption key is the index pepper", func(e map[string]string) { e["RFC_INDEX_PEPPER"] = e["RFC_ENC_KEY"] },
			"RFC_INDEX_PEPPER must differ from RFC_ENC_KEY"},
	}
	for _, tt := range tests {
		for _, appEnv := range []string{"dev", "production"} {
			t.Run(tt.name+"/"+appEnv, func(t *testing.T) {
				env := withTaxpayersKeys(withAuthKeys(map[string]string{
					"APP_ENV": appEnv, "TRUSTED_PROXY_HOPS": "0", "DATABASE_URL": "postgres://u@127.0.0.1:1/db",
				}))
				tt.mutate(env)

				err := runWith(env)

				require.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			})
		}
	}
}

func TestRun_refusesMissingTaxpayerKeysOutsideDev(t *testing.T) {
	env := withAuthKeys(map[string]string{
		"APP_ENV": "staging", "TRUSTED_PROXY_HOPS": "0", "DATABASE_URL": "postgres://u@127.0.0.1:1/db",
	})

	err := runWith(env)

	require.Error(t, err)
	assert.ErrorContains(t, err, "RFC_ENC_KEY")
	assert.ErrorContains(t, err, "RFC_INDEX_PEPPER")
}

// In dev a missing RFC key becomes an ephemeral random one, said loudly in
// the log: the local stack starts without them.
func TestTaxpayersCrypto_devFillsMissingKeysWithEphemeralOnes(t *testing.T) {
	var logs bytes.Buffer
	w := taxpayersWiring{env: config.EnvDev, keys: config.Taxpayers{RFCKeyID: 1}, logger: logging.New(&logs, nil)}

	vault, err := newRFCVault(w)

	require.NoError(t, err)
	rfc, err := domain.NewRFCFor(domain.PersonTypeFisica, "PEXJ900215HK8")
	require.NoError(t, err)
	sealed, err := vault.Seal(rfc, usecase.RowBinding{
		TaxpayerID: "0199d4a2-8f00-7c44-a1b2-3c4d5e6f7a80", UserID: "0199d4a2-0000-7000-8000-00000000000a",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, sealed.Ciphertext)
	assert.Contains(t, logs.String(), "RFC_ENC_KEY")
	assert.Contains(t, logs.String(), "RFC_INDEX_PEPPER")
	assert.Contains(t, logs.String(), "ephemeral")
}

// errStubTaxpayers is what the stub use cases answer: the platform tests
// never reach them, and a reached one is a 500, not a panic.
var errStubTaxpayers = errors.New("stub: not wired")

type stubTaxpayerUseCases struct{}

func (stubTaxpayerUseCases) Create(context.Context, string, usecase.CreateInput) (usecase.CreateResult, error) {
	return usecase.CreateResult{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) List(context.Context, string, bool) ([]usecase.Taxpayer, error) {
	return nil, errStubTaxpayers
}

func (stubTaxpayerUseCases) Get(context.Context, string, string) (usecase.Taxpayer, error) {
	return usecase.Taxpayer{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) Update(context.Context, string, string, usecase.UpdateInput) (usecase.Taxpayer, error) {
	return usecase.Taxpayer{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) Archive(context.Context, string, string) (usecase.Taxpayer, error) {
	return usecase.Taxpayer{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) Unarchive(context.Context, string, string) (usecase.Taxpayer, error) {
	return usecase.Taxpayer{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) Delete(context.Context, string, string, bool) error {
	return errStubTaxpayers
}

func (stubTaxpayerUseCases) Add(context.Context, string, string, usecase.RegimeInput) (usecase.TaxpayerRegime, error) {
	return usecase.TaxpayerRegime{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) UpdateOptions(context.Context, string, string, string, map[string]domain.OptionValue) (usecase.TaxpayerRegime, error) {
	return usecase.TaxpayerRegime{}, errStubTaxpayers
}

func (stubTaxpayerUseCases) End(context.Context, string, string, usecase.EndRegimeInput) (usecase.TaxpayerRegime, error) {
	return usecase.TaxpayerRegime{}, errStubTaxpayers
}

type stubCatalog struct{}

func (stubCatalog) List(context.Context, string, string) ([]usecase.CatalogEntry, error) {
	return nil, errStubTaxpayers
}

// stubTaxpayersModule has stub use cases and no write budget; contract and
// integration tests wire the real module.
func stubTaxpayersModule() taxpayersModule {
	stub := stubTaxpayerUseCases{}
	return taxpayersModule{handler: taxpayershttp.NewHandler(taxpayershttp.UseCases{
		Taxpayers: stub, Regimes: stub, Catalog: stubCatalog{},
	})}
}

func TestNewHandler_refusesToServeWithoutTheTaxpayersHandler(t *testing.T) {
	_, err := newHandler(testConfig(), deps{db: pingOK{}, flags: flagtest.Static{}, auth: stubAuthModule(t), now: time.Now,
		logger: logging.New(io.Discard, nil)})

	assert.ErrorContains(t, err, "no taxpayers handler")
}

func TestTaxpayersCrypto_neverEphemeralOutsideDev(t *testing.T) {
	w := taxpayersWiring{env: config.EnvProduction, keys: config.Taxpayers{RFCKeyID: 1}, logger: logging.New(io.Discard, nil)}

	_, err := newRFCVault(w)

	require.Error(t, err)
	assert.True(t, errors.Is(err, errMissingRFCKeys), err.Error())
}
