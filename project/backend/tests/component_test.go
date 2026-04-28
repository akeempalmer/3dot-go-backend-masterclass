package tests_test

import (
	"eats/backend/common/testutils"
	"testing"
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx := t.Context()

	var country = testutils.GetRandomCountry()

	customerUUID := registerCustomerInCity(ctx, t, clients, country, "New York")

	t.assert.NotEmpty(t, customerUUID)

}
