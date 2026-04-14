package admin

import (
	"net/http"

	"morent-backend/internal/di"
	"morent-backend/internal/modules/transport/http/common"
)

func Register(mux *http.ServeMux, container *di.Container) {
	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return common.WrapCORS(common.WithAdmin(container.AuthService, h))
	}

	mux.HandleFunc("/Admin/Users", adminOnly(container.AdminHandler.ListUsers))
	mux.HandleFunc("/Admin/Rentals", adminOnly(container.AdminHandler.ListRentals))
	mux.HandleFunc("/Admin/Favorites", adminOnly(container.AdminHandler.ListFavorites))
	mux.HandleFunc("/Admin/Comments", adminOnly(container.AdminHandler.ListComments))
	mux.HandleFunc("/Admin/Logs", adminOnly(container.AdminHandler.ListLogs))
	mux.HandleFunc("/Admin/Aggregator/Cars", adminOnly(container.AdminHandler.ListAggregatorCars))
	mux.HandleFunc("/Admin/Aggregator/Import", adminOnly(container.AdminHandler.ImportAggregatorCar))
	mux.HandleFunc("/Admin/Users/", adminOnly(container.AdminHandler.DeleteUser))
	mux.HandleFunc("/Admin/Rentals/", adminOnly(container.AdminHandler.DeleteRental))
	mux.HandleFunc("/Admin/Favorites/", adminOnly(container.AdminHandler.DeleteFavorite))
	mux.HandleFunc("/Admin/Comments/", adminOnly(container.AdminHandler.DeleteComment))
}
