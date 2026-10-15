package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/romodevx/control-contable/backend/cmd/internal/sealing"
	authdomain "github.com/romodevx/control-contable/backend/internal/auth/domain"
	authinfra "github.com/romodevx/control-contable/backend/internal/auth/infra"
	keyinfra "github.com/romodevx/control-contable/backend/internal/keyrotation/infra"
	keyrotation "github.com/romodevx/control-contable/backend/internal/keyrotation/usecase"
	"github.com/romodevx/control-contable/backend/internal/platform/crypto"
	"github.com/romodevx/control-contable/backend/internal/platform/pgdb"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/pushtoken"
	remusecase "github.com/romodevx/control-contable/backend/internal/reminders/usecase"
	tpinfra "github.com/romodevx/control-contable/backend/internal/taxpayers/infra"
	tpusecase "github.com/romodevx/control-contable/backend/internal/taxpayers/usecase"
)

// newKeyResealJob wires the key re-seal job (ADR-0012) over the columns
// whose keys this worker holds: the push token always (the reminder job
// needs its keys), the RFC and the email only when RFC_ENC_KEY /
// AUTH_EMAIL_KEY are given to the worker for a rotation. Without any it
// still reports every column's readiness (key_usage).
func newKeyResealJob(units *pgdb.TxManager, w jobWiring) (*keyrotation.ResealJob, error) {
	columns, err := resealColumns(w)
	if err != nil {
		return nil, err
	}
	targets := make([]string, len(columns))
	for i, c := range columns {
		targets[i] = c.Target
	}
	w.logger.Info("key reseal columns", slog.String("columns", strings.Join(targets, ",")))
	return keyrotation.NewResealJob(keyrotation.ResealDeps{
		Tx: units, System: units, Usage: keyinfra.NewUsageReader(), Columns: columns, Logger: w.logger,
		BatchSize: keyrotation.DefaultBatchSize, MaxBatches: keyrotation.DefaultMaxBatches,
	})
}

// resealColumns builds one column per sealing key the worker holds.
func resealColumns(w jobWiring) ([]keyrotation.Column, error) {
	var columns []keyrotation.Column
	if email := w.resealKeys.Email; !email.IsZero() {
		keyring, err := sealing.Keyring(email.KeyID, email.Key.RevealBytes(), email.Previous)
		if err != nil {
			return nil, fmt.Errorf("key reseal: email keyring: %w", err)
		}
		columns = append(columns, keyrotation.Column{
			Target: keyinfra.TargetUserEmail, Store: keyinfra.NewUserEmailStore(),
			Sealer: emailSealer{keyring: keyring, accounts: authinfra.NewAccountStore(keyring)},
		})
	}
	if rfc := w.resealKeys.RFC; !rfc.IsZero() {
		keyring, err := sealing.Keyring(rfc.KeyID, rfc.Key.RevealBytes(), rfc.Previous)
		if err != nil {
			return nil, fmt.Errorf("key reseal: rfc keyring: %w", err)
		}
		columns = append(columns, keyrotation.Column{
			Target: keyinfra.TargetTaxpayerRFC, Store: keyinfra.NewTaxpayerRFCStore(),
			Sealer: rfcSealer{resealer: tpinfra.NewRFCResealer(keyring)},
		})
	}
	if hasReminderKeys(w.reminderKeys) {
		keyring, vault, err := newTokenKeyringAndVault(w.reminderKeys)
		if err != nil {
			return nil, fmt.Errorf("key reseal: %w", err)
		}
		columns = append(columns, keyrotation.Column{
			Target: keyinfra.TargetDeviceToken, Store: keyinfra.NewDeviceTokenStore(),
			Sealer: tokenSealer{keyring: keyring, vault: vault},
		})
	}
	return columns, nil
}

// emailSealer re-seals users.email_encrypted bound to its user (auth's own
// emailAAD); the row id is the user id.
type emailSealer struct {
	keyring  *crypto.Keyring
	accounts *authinfra.AccountStore
}

func (s emailSealer) ActiveKeyID() crypto.KeyID { return s.keyring.ActiveKeyID() }

func (s emailSealer) Reseal(ciphertext []byte, ref keyrotation.Ref) ([]byte, error) {
	id, err := authdomain.ParseUserID(ref.UserID)
	if err != nil {
		return nil, fmt.Errorf("key reseal: user id: %w", err)
	}
	return s.accounts.ResealEmail(ciphertext, id)
}

// rfcSealer re-seals taxpayers.rfc_encrypted bound to its taxpayer and user.
type rfcSealer struct {
	resealer *tpinfra.RFCResealer
}

func (s rfcSealer) ActiveKeyID() crypto.KeyID { return s.resealer.ActiveKeyID() }

func (s rfcSealer) Reseal(ciphertext []byte, ref keyrotation.Ref) ([]byte, error) {
	return s.resealer.Reseal(ciphertext, tpusecase.RowBinding{TaxpayerID: ref.ID, UserID: ref.UserID})
}

// tokenSealer re-seals devices.token_ciphertext bound to its installation
// and user.
type tokenSealer struct {
	keyring *crypto.Keyring
	vault   *pushtoken.Vault
}

func (s tokenSealer) ActiveKeyID() crypto.KeyID { return s.keyring.ActiveKeyID() }

func (s tokenSealer) Reseal(ciphertext []byte, ref keyrotation.Ref) ([]byte, error) {
	return s.vault.Reseal(ciphertext, remusecase.TokenBinding{InstallationID: ref.ID, UserID: ref.UserID})
}

var (
	_ keyrotation.Sealer = emailSealer{}
	_ keyrotation.Sealer = rfcSealer{}
	_ keyrotation.Sealer = tokenSealer{}
)
