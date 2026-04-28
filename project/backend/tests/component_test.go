package tests_test

import (
	"eats/backend/common/testutils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComponent_CriticalFlow(t *testing.T) {
	t.Parallel()
	t.Helper()

	ctx := t.Context()

	var country = testutils.GenerateRandomCountry()

	var clients = newTestClients(t)

	customerUUID := registerCustomerInCity(ctx, t, clients, country, "New York")

	assert.NotEmpty(t, customerUUID)

}
