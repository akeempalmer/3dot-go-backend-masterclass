package db

import (
	"context"
	"errors"
	"fmt"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"eats/backend/common"
	"eats/backend/common/log"
	"eats/backend/common/shared"
	"eats/backend/orders/adapters/db/dbmodels"
	"eats/backend/orders/app"
)

type RestaurantRepository struct {
	db *pgxpool.Pool
}

func NewRestaurantRepository(db *pgxpool.Pool) *RestaurantRepository {
	if db == nil {
		panic("db connection pool cannot be nil")
	}

	return &RestaurantRepository{
		db: db,
	}
}

func (r *RestaurantRepository) UpsertRestaurant(ctx context.Context, restaurantUUID app.RestaurantUUID, restaurant app.OnboardRestaurant) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		dbRestaurantMenu, err := queries.GetRestaurantMenu(ctx, restaurantUUID)

		log.FromContext(ctx).With("restaurant_uuid", restaurantUUID).Info("Upserting restaurant")

		dbRestaurant, err := queries.UpsertRestaurant(ctx, dbmodels.UpsertRestaurantParams{
			restaurantUUID,
			restaurant.Name,
			restaurant.Description,
			restaurant.Address,
			restaurant.Currency.Code(),
		})
		if err != nil {
			return fmt.Errorf("upsert restaurant failed: %w", err)
		}

		// Currency is immutable after creation - the upsert doesn't update it.
		// Check here catches attempts to change it and returns a clear error.
		if dbRestaurant.Currency != restaurant.Currency.Code() {
			return common.NewInvalidInputError("cannot-change-currency", "cannot change restaurant currency once set")
		}

		// TODO: upsert menu items
		for _, item := range restaurant.MenuItems {
			err = queries.UpsertRestaurantMenuItem(ctx, dbmodels.UpsertRestaurantMenuItemParams{
				RestaurantMenuItemUuid: item.MenuItemUUID,
				RestaurantUuid:         restaurantUUID,
				Name:                   item.Name,
				GrossPrice:             item.GrossPrice,
				Ordering:               item.Ordering,
				IsArchived:             false,
			})
			if err != nil {
				return fmt.Errorf("upsert restaurant menu position failed for menu position %s: %w", item.MenuItemUUID, err)
			}
		}

		currentMenuItemsUUIDs := make([]common.UUID, 0)
		for _, item := range dbRestaurantMenu {
			currentMenuItemsUUIDs = append(currentMenuItemsUUIDs, item.OrdersRestaurantMenuItem.RestaurantMenuItemUuid.UUID)
		}

		// get the list of different uuid.
		menuItemsToArchive := make([]common.UUID, 0)
		for _, currentUUID := range currentMenuItemsUUIDs {
			found := false

			for _, newItem := range restaurant.MenuItems {
				if currentUUID == newItem.MenuItemUUID.UUID {
					found = true
					break
				}
			}
			if !found {

				menuItemsToArchive = append(menuItemsToArchive, currentUUID)
			}
		}

		// Call archive
		if len(menuItemsToArchive) > 0 {
			pgUUIDs := make([]pgtype.UUID, 0, len(menuItemsToArchive))

			for _, id := range menuItemsToArchive {
				pgUUIDs = append(pgUUIDs, pgtype.UUID{
					Bytes: id,
					Valid: true,
				})
			}

			err = queries.ArchiveMenuItems(ctx, pgUUIDs)
			if err != nil {
				return fmt.Errorf("archive menu items failed: %w", err)
			}
		}
		return nil

	})
}

func (r *RestaurantRepository) GetRestaurantMenu(ctx context.Context, restaurantUUID app.RestaurantUUID) (app.RestaurantMenu, error) {
	queries := dbmodels.New(r.db)

	dbItems, err := queries.GetRestaurantMenu(ctx, restaurantUUID)
	if err != nil {
		return app.RestaurantMenu{}, fmt.Errorf("get restaurant menu failed: %w", err)
	}

	log.FromContext(ctx).With("restaurant_uuid", restaurantUUID, "count", len(dbItems)).Info("Fetched menu items")

	items := make([]app.MenuItem, len(dbItems))
	for i, dbItem := range dbItems {
		items[i] = app.MenuItem{
			MenuItemUUID: dbItem.OrdersRestaurantMenuItem.RestaurantMenuItemUuid,
			Name:         dbItem.OrdersRestaurantMenuItem.Name,
			GrossPrice:   dbItem.OrdersRestaurantMenuItem.GrossPrice,
			Ordering:     dbItem.OrdersRestaurantMenuItem.Ordering,
		}
	}

	restaurant, err := queries.GetRestaurant(ctx, restaurantUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.RestaurantMenu{}, common.NewNotFoundError("restaurant-not-found", "restaurant not found")
		}
		return app.RestaurantMenu{}, fmt.Errorf("get restaurant %s failed: %w", restaurantUUID, err)
	}

	currencyCode := shared.MustNewCurrency(restaurant.Currency)

	return app.RestaurantMenu{
		RestaurantName: restaurant.Name,
		Address:        restaurant.Address,
		Description:    restaurant.Description,
		Currency:       currencyCode,
		Positions:      items,
	}, nil
}
