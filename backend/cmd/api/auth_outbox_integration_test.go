//go:build integration

package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/internal/platform/flags/flagtest"
)

// The timing oracle closed (MODULE-auth §4, M1-06 carry-over): register and
// resend answer every address alike and send nothing while the request runs.
// A new address, a pending sign-up and a verified account get the same 202
// and no email in the API's logs until the worker's delivery runs. Asserted
// on what was sent, not on wall time, so it is deterministic.
func TestAuthOutbox_noEmailIsSentDuringTheRequest(t *testing.T) {
	env := newContractEnv(t, pingOK{})
	env.apply(t, []string{"module_auth"}, flagtest.Static{"module_auth": true})
	env.signUp(t, signUpEmail, acceptableSample)
	sentSoFar := emailsInLogs(env.logs.String())

	for _, req := range []*http.Request{
		env.post("/v1/auth/register", registerJSON("nueva@example.mx", "v1")),
		env.post("/v1/auth/register", registerJSON(signUpEmail, "v1")),
		env.post("/v1/auth/verify-email/resend", `{"email":"nueva@example.mx"}`),
	} {
		rec := env.do(req)
		require.Equal(t, http.StatusAccepted, rec.Code, rec.Body.String())
		assert.JSONEq(t, verifiedSent, rec.Body.String())
	}
	assert.Equal(t, sentSoFar, emailsInLogs(env.logs.String()), "nothing sent in the requests")

	env.deliver(t)

	assert.Equal(t, sentSoFar+2, emailsInLogs(env.logs.String()), "the worker sent the code and the notice")
}

// emailsInLogs counts the emails the dev log sender wrote.
func emailsInLogs(logs string) int {
	return strings.Count(logs, `"msg":"dev email: `)
}
