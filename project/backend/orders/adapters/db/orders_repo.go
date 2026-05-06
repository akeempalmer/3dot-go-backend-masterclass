package db

import (
	"context"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/shared"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type OrdersRepository struct {
	db *pgxpool.Pool
}

func NewOrdersRepository(db *pgxpool.Pool) *OrdersRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &OrdersRepository{
		db: db,
	}
}

func (o *OrdersRepository) CreateQuote(
	ctx context.Context,
	restaurantUUID app.RestaurantUUID,
	menuItems app.CreateQuoteItems,
	updateFn func(
		ctx context.Context,
		menuItems map[app.RestaurantMenuItemUUID]app.MenuItem,
		restaurantCurrency shared.Currency,
		restaurantAddress shared.Address,
	) (app.Quote, []app.QuoteMenuItem, error),
) (app.Quote, error) {
	var quote app.Quote

	err := common.UpdateInTx(ctx, o.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		// 1. Extract UUIDs
		menuItemUUIDs := make([]common.UUID, 0, len(menuItems))
		for _, item := range menuItems {
			menuItemUUIDs = append(menuItemUUIDs, item.MenuItemUUID.UUID)
		}

		// 2. Fetch menu items from DB
		dbMenuItems, err := queries.GetMenuItemsByUUIDs(ctx, dbmodels.GetMenuItemsByUUIDsParams{
			restaurantUUID,
			menuItemUUIDs,
		})
		if err != nil {
			return err
		}

		// 3. Convert to map for updateFn
		menuItemMap := make(map[app.RestaurantMenuItemUUID]app.MenuItem)
		for _, dbItem := range dbMenuItems {
			menuItemMap[dbItem.RestaurantMenuItemUuid] = app.MenuItem{
				MenuItemUUID: dbItem.RestaurantMenuItemUuid,
				Name:         dbItem.Name,
				Ordering:     dbItem.Ordering,
				GrossPrice:   dbItem.GrossPrice,
				IsArchived:   dbItem.IsArchived,
			}
		}

		// 4. Fetch restaurant
		dbRestaurant, err := queries.GetRestaurant(ctx, restaurantUUID)
		if err != nil {
			return err
		}

		// 5. Call domain logic
		q, quoteMenuItems, err := updateFn(ctx, menuItemMap, dbRestaurant.Currency, dbRestaurant.Address)
		if err != nil {
			return err
		}

		// 6. Persist quotes
		err = queries.AddQuote(ctx, dbmodels.AddQuoteParams{
			QuoteUuid:          q.QuoteUUID,
			CustomerUuid:       q.CustomerUUID,
			RestaurantUuid:     q.RestaurantUUID,
			DeliveryAddress:    q.DeliveryAddress,
			ItemsSubtotalGross: q.ItemsSubtotalGross,
			ServiceFeeGross:    q.ServiceFeeGross,
			DeliveryFeeGross:   q.DeliveryFeeGross,
			TotalAmountGross:   q.TotalAmountGross,
			TotalTax:           q.TotalTax,
			CreatedAt:          q.CreatedAt,
			Currency:           q.Currency,
		})
		if err != nil {
			return err
		}

		// 7. Persist quote items
		dbQuoteItems := make([]dbmodels.AddQuoteItemsParams, 0, len(quoteMenuItems))
		for _, item := range quoteMenuItems {
			dbQuoteItems = append(dbQuoteItems, dbmodels.AddQuoteItemsParams{
				QuoteItemUuid: item.MenuItemUUID.UUID,
				QuoteUuid:     q.QuoteUUID,
				MenuItemUuid:  item.MenuItemUUID,
				GrossPrice:    item.GrossPrice,
				Quantity:      int32(item.Quantity),
			})
		}

		_, err = queries.AddQuoteItems(ctx, dbQuoteItems)
		if err != nil {
			return err
		}

		quote = q
		return nil
	})

	return quote, err
}

// backend/orders/adapters/db/orders_repo.go:102:40: cannot use item (variable of struct type app.QuoteMenuItem) as dbmodels.AddQuoteItemsParams value in argument to append
