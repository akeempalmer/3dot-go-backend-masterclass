package http

import (
	"context"
	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/adapters/db/dbmodels"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db *pgxpool.Pool
}

func NewHandler(db *pgxpool.Pool) Handler {
	if db == nil {
		panic("db cannot be nil")
	}

	return Handler{
		db: db,
	}
}

func Register(ctx context.Context, e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	customerUuid := common.NewUUIDv7()
	queries := dbmodels.New(h.db)

	fmt.Printf("Registering customer with UUID: %s\n", request)

	err := queries.InsertCustomer(ctx, dbmodels.InsertCustomerParams{
		CustomerUuid: customerUuid,
		Name:         request.Body.Name,
		Email:        string(request.Body.Email),
		Address:      shared.OpenapiAddressToSharedAddress(request.Body.Address),
		PhoneNumber:  request.Body.PhoneNumber,
	})

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUuid,
	}, nil
}
