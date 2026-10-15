package fcm_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm"
	"github.com/romodevx/control-contable/backend/internal/reminders/infra/fcm/fcmtest"
	"github.com/romodevx/control-contable/backend/internal/reminders/usecase"
)

// The production path: the service account from the secret manager is
// exchanged for an access token, which every FCM call carries.
func TestNew_serviceAccountCredentials(t *testing.T) {
	accessToken := uuid.NewString()
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": accessToken, "token_type": "Bearer", "expires_in": int(time.Hour.Seconds()),
		})
	}))
	t.Cleanup(tokenSrv.Close)
	var authorized atomic.Bool
	fcmSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorized.Store(r.Header.Get("Authorization") == "Bearer "+accessToken)
		writeReply(w, replyOK)
	}))
	t.Cleanup(fcmSrv.Close)

	sender, err := fcm.New(t.Context(), fcm.Config{
		ProjectID: testProject, CredentialsJSON: fcmtest.ServiceAccountJSON(t, testProject, tokenSrv.URL), SendTimeout: 5 * time.Second,
	}, option.WithEndpoint(fcmSrv.URL))
	require.NoError(t, err)
	dev := target("dev")

	got, err := sender.Send(t.Context(), notification(), []usecase.PushTarget{dev})

	require.NoError(t, err)
	assert.Equal(t, []usecase.PushDelivery{{DeviceID: dev.DeviceID, Outcome: usecase.PushDelivered}}, got)
	assert.True(t, authorized.Load(), "FCM call carries the exchanged access token")
}

// Only a service account key is accepted: another credential type (an
// external account that would send the SDK to another token endpoint) fails
// at construction.
func TestNew_rejectsOtherCredentialTypes(t *testing.T) {
	creds := strings.Replace(string(fcmtest.ServiceAccountJSON(t, testProject, "http://127.0.0.1:1")),
		`"service_account"`, `"external_account"`, 1)

	_, err := fcm.New(t.Context(), fcm.Config{ProjectID: testProject, CredentialsJSON: []byte(creds)})

	require.ErrorIs(t, err, fcm.ErrConfig)
	assert.NotContains(t, err.Error(), "PRIVATE KEY")
}
