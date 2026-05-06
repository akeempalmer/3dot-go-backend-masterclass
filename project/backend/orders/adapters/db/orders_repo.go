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

func NewOrdersRepo(db *pgxpool.Pool) *OrdersRepository {
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

		menuItemSlice := make([]app.MenuItem, 0)
		for _, item := range menuItems {
			dbMenuItems, err := queries.GetMenuItemsByUUIDs(item.Uuid)

			if err != nil {
				return err
			}

			menuItemSlice = append(menuItemSlice, dbMenuItems)
		}

		dbRestaurant, err := queries.GetRestaurant(restaurantUUID)
		if err != nil {
			return err
		}

		quote, quoteMenuItems, err := updateFn(ctx, menuItemSlice, dbRestaurant.Currency, dbRestaurant.Address)
		if err != nil {
			return err
		}

		err = queries.AddQuote(ctx, quote)
		if err != nil {

		}

		return nil
	})

	return quote, err
}
