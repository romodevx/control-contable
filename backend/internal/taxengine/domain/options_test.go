package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/romodevx/control-contable/backend/internal/taxengine/domain"
)

func TestOptions(t *testing.T) {
	activities := []string{"lodging", "goods_services"}
	values := map[domain.OptionKey]domain.OptionValue{
		"definitive_option": domain.BoolOption(true),
		"deduction_mode":    domain.TextOption("blind_35"),
		"activities":        domain.ListOption(activities),
	}
	o := domain.NewOptions(values)
	activities[0] = "changed"
	delete(values, "definitive_option")

	b, ok := o.Bool("definitive_option")
	assert.True(t, ok, "NewOptions copies the map")
	assert.True(t, b)
	_, ok = o.Bool("deduction_mode")
	assert.False(t, ok)

	s, ok := o.Text("deduction_mode")
	assert.True(t, ok)
	assert.Equal(t, "blind_35", s)
	_, ok = o.Text("activities")
	assert.False(t, ok)

	list, ok := o.List("activities")
	assert.True(t, ok)
	assert.Equal(t, []string{"lodging", "goods_services"}, list, "ListOption copies the items")
	list[0] = "mutated"
	again, _ := o.List("activities")
	assert.Equal(t, "lodging", again[0], "List returns a copy")
	_, ok = o.List("missing")
	assert.False(t, ok)

	var zero domain.Options
	_, ok = zero.Bool("definitive_option")
	assert.False(t, ok)
}
