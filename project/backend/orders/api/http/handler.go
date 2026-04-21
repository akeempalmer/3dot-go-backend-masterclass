package http

import (
	"context"

	"eats/backend/common"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func Register(ctx context.Context, e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	customerUuid := common.NewUUIDv7()
	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUuid,
	}, nil
}
