package http

import (
	"context"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/app"
)

type Handler struct {
	service *app.Service
}

func NewHandler(
	service *app.Service,
) Handler {
	if service == nil {
		panic("service cannot be nil")
	}

	return Handler{
		service: service,
	}
}

func (h Handler) RegisterCustomer(ctx context.Context, request RegisterCustomerRequestObject) (RegisterCustomerResponseObject, error) {
	addr, err := openapiAddressToSharedAddress(request.Body.Address)
	if err != nil {
		return nil, common.NewInvalidInputError("invalid-address", "invalid address: %s", err)
	}

	customerUUID := CustomerUUID{common.NewUUIDv7()}

	err = h.service.RegisterCustomer(ctx, app.Customer{
		CustomerUUID: customerUUID,
		Name:         request.Body.Name,
		Email:        string(request.Body.Email),
		// address should be ideally normalized to ensure consistent city names and postal codes
		// across customers, restaurants, and delivery addresses
		Address:     addr,
		PhoneNumber: request.Body.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}

	return RegisterCustomer201JSONResponse{
		CustomerUuid: customerUUID,
	}, nil
}

func (h Handler) OnboardRestaurant(ctx context.Context, request OnboardRestaurantRequestObject) (OnboardRestaurantResponseObject, error) {

	// MAP the request body fields to app.Onboard Resaurant and []app.menuItem.
	var menuItemList = make([]app.MenuItem, 0)

	for _, item := range request.Body.MenuItems {

		menuItemList = append(menuItemList, app.MenuItem{
			MenuItemUUID: item.Uuid,
			Name:         item.Name,
			Ordering:     float64(item.Ordering),
			GrossPrice:   item.GrossPrice,
			IsArchived:   false,
		})
	}

	addr, err := openapiAddressToSharedAddress(request.Body.Address)
	if err != nil {
		return nil, common.NewInvalidInputError("invalid-address", "invalid address: %s", err)
	}

	appRequest := app.OnboardRestaurant{
		Name:        request.Body.Name,
		Address:     addr,
		Currency:    request.Body.Currency,
		Description: request.Body.Description,
		MenuItems:   menuItemList,
	}

	err := h.service.OnboardRestaurant(ctx, request.RestaurantUuid, appRequest)

	if err != nil {
		// handle error
		panic("error parsing appmenu")
	}

	return OnboardRestuarant204Response{}, nil

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

func Register(ctx context.Context, e EchoRouter, handler Handler) error {
	RegisterHandlers(e, NewStrictHandler(handler, nil))

	return nil
}
