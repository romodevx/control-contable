package main

import (
	"context"
	"reflect"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/romodevx/control-contable/backend/adapter/http/gen"
	"github.com/romodevx/control-contable/backend/internal/platform/apperr"
)

// builtOperations are the operations with a real handler. Building another
// one means deleting its stub from notbuilt.go (the compiler insists) and
// adding it here.
var builtOperations = []string{"getHealth", "listEnabledFlags", "register", "resendVerification", "verifyEmail"}

func TestNotBuilt_coversExactlyTheOperationsWithoutAHandler(t *testing.T) {
	var want []string
	for id := range gen.OperationModules() {
		if !slices.Contains(builtOperations, id) {
			want = append(want, id)
		}
	}

	assert.ElementsMatch(t, want, notBuiltOperations())
}

// P0-29 FU2: a stub reached by any path answers the hidden-module 404, never
// a nil-interface panic.
func TestNotBuilt_everyStubAnswersNotFound(t *testing.T) {
	stub := reflect.ValueOf(notBuilt{})
	for i := range stub.NumMethod() {
		method := stub.Type().Method(i)
		t.Run(method.Name, func(t *testing.T) {
			out := stub.Method(i).Call([]reflect.Value{
				reflect.ValueOf(context.Background()),
				reflect.New(method.Type.In(2)).Elem(),
			})

			require.Len(t, out, 2)
			assert.True(t, out[0].IsNil(), "no response object")
			err, ok := out[1].Interface().(error)
			require.True(t, ok)
			assert.Equal(t, apperr.CodeNotFound, apperr.Mapper{}.Resolve(err).Code)
		})
	}
}

func TestLowerFirst(t *testing.T) {
	assert.Equal(t, "receiveRevenueCatWebhook", lowerFirst("ReceiveRevenueCatWebhook"))
	assert.Equal(t, "x", lowerFirst("X"))
}
