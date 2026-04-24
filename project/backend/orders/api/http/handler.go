package http

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/app"
)

type CustomerRepository interface {
	RegisterCustomer(ctx context.Context, customer app.Customer) error
}

type Handler struct {
	services *app.Service
}

func NewHandler(
	services *app.Service,
) Handler {
	if services == nil {
		panic("services cannot be nil")
	}

	return Handler{
		services: services,
	}
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	customerUUID := app.CustomerUUID{common.NewUUIDv7()}

	customerAddress, _ := openapiAddressToSharedAddress(request.Body.Address)

	err := h.services.RegisterCustomer(ctx, app.Customer{
		CustomerUUID: customerUUID,
		Name:         request.Body.Name,
		Email:        string(request.Body.Email),
		PhoneNumber:  request.Body.PhoneNumber,
		Address:      customerAddress,
	})
	if err != nil {
		return nil, err
	}

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUUID,
	}, nil
}

func Register(ctx context.Context, e EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}

func openapiAddressToSharedAddress(addr Address) (shared.Address, error) {
	sharedAddr, err := shared.NewAddress(
		addr.Line1,
		addr.Line2,
		addr.PostalCode,
		addr.City,
		addr.CountryCode,
	)
	if err != nil {
		return shared.Address{}, err
	}

	return sharedAddr, nil
}
