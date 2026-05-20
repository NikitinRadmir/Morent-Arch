package admin

import (
	"net/http"
	"strings"

	"morent-backend/internal/di"
	"morent-backend/internal/modules/transport/http/common"
)

func Register(mux *http.ServeMux, container *di.Container) {
	adminOnly := func(h http.HandlerFunc) http.HandlerFunc {
		return common.WrapCORS(common.WithAdmin(container.AuthService, container.Config, h))
	}

	mux.HandleFunc("/Admin/Users", adminOnly(container.AdminHandler.Users))
	mux.HandleFunc("/Admin/Rentals", adminOnly(container.AdminHandler.ListRentals))
	mux.HandleFunc("/Admin/Favorites", adminOnly(container.AdminHandler.ListFavorites))
	mux.HandleFunc("/Admin/Comments", adminOnly(container.AdminHandler.Comments))
	mux.HandleFunc("/Admin/Logs", adminOnly(container.AdminHandler.ListLogs))
	mux.HandleFunc("/Admin/Aggregator/Cars", adminOnly(container.AdminHandler.ListAggregatorCars))
	mux.HandleFunc("/Admin/Aggregator/Import", adminOnly(container.AdminHandler.ImportAggregatorCar))

	mux.HandleFunc("/Admin/Users/", adminOnly(container.AdminHandler.DeleteUser))
	mux.HandleFunc("/Admin/Rentals/", adminOnly(container.AdminHandler.DeleteRental))
	mux.HandleFunc("/Admin/Favorites/", adminOnly(container.AdminHandler.DeleteFavorite))
	mux.HandleFunc("/Admin/Comments/", adminOnly(container.AdminHandler.DeleteComment))

	// Каталог и медиа — только для админа (раньше были публичными).
	mux.HandleFunc("/Admin/Cars", adminOnly(adminCars(container)))
	mux.HandleFunc("/Admin/Cars/", adminOnly(adminCars(container)))
	mux.HandleFunc("/Admin/Media/Upload", adminOnly(container.MediaHandler.Upload))
	// Legacy path для совместимости (тот же handler, только admin).
	mux.HandleFunc("/Media/Upload", adminOnly(container.MediaHandler.Upload))
	mux.HandleFunc("/media/upload", adminOnly(container.MediaHandler.Upload))
}

func adminCars(c *di.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		hasID := len(parts) >= 3 && parts[len(parts)-1] != "" && parts[len(parts)-1] != "Cars"
		switch r.Method {
		case http.MethodPost:
			if hasID {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			c.CarHandler.CreateCar(w, r)
		case http.MethodPut:
			if !hasID {
				http.Error(w, "car id required", http.StatusBadRequest)
				return
			}
			c.CarHandler.UpdateCar(w, r)
		case http.MethodDelete:
			if !hasID {
				http.Error(w, "car id required", http.StatusBadRequest)
				return
			}
			c.CarHandler.DeleteCar(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
