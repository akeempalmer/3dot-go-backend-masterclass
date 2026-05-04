package tests_test

import (
	"testing"

	"eats/backend/common/testutils"

	"github.com/stretchr/testify/assert"
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx := t.Context()

	country := testutils.GenerateRandomCountry()

	clients := newTestClients(t)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, "New York")

	assert.NotEmpty(t, customerUUID)
}
