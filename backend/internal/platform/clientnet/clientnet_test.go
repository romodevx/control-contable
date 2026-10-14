package clientnet_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/clientnet"
)

func TestFrom(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{name: "absent", ctx: context.Background(), want: ""},
		{name: "present", ctx: clientnet.With(context.Background(), "203.0.113.0/24"), want: "203.0.113.0/24"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, clientnet.From(tt.ctx))
		})
	}
}
