package requestid_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/platform/requestid"
)

func TestFrom(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{name: "absent", ctx: context.Background(), want: ""},
		{name: "present", ctx: requestid.With(context.Background(), "req_01J9Z6Q4X8"), want: "req_01J9Z6Q4X8"},
		{
			name: "innermost wins",
			ctx:  requestid.With(requestid.With(context.Background(), "outer"), "inner"),
			want: "inner",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, requestid.From(tt.ctx))
		})
	}
}
