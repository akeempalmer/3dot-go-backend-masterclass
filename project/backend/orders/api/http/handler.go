package http

import (
	"context"
	"eats/backend/common"
	"eats/backend/common/shared"
)

type Handler struct {
	customerRepo CustomerRepository
}

func NewHandler(customerRepo CustomerRepository) Handler {
	if customerRepo == nil {
		panic("customer repo cannot be nil")
	}

	return Handler{
		customerRepo: customerRepo,
	}
}

type CustomerRepository interface {
	RegisterCustomer(ctx context.Context, customerUUID common.UUID, customer RegisterCustomer) error
}

func Register(ctx context.Context, e common.EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	customerUuid := common.NewUUIDv7()

	h.customerRepo.RegisterCustomer(ctx, customerUuid, *request.Body)

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUuid,
	}, nil
}

func openapiAddressToSharedAddress(addr Address) (shared.Address, error) {
	return shared.NewAddress(
		addr.Line1,
		addr.Line2,
		addr.PostalCode,
		addr.City,
		addr.CountryCode,
	)
}
