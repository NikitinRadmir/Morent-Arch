package public

import (
	"net/http"

	"morent-backend/internal/di"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/server"
)

func Register(mux *http.ServeMux, container *di.Container) {
	groups := []server.RouteGroup{
		{Prefix: "/cars", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: container.CarHandler.GetAll},
			{Method: http.MethodGet, Path: "/:id", Handler: container.CarHandler.GetById},
			{Method: http.MethodGet, Path: "/filtered", Handler: container.CarHandler.GetFiltered},
			{Method: http.MethodPost, Path: "/", Handler: container.CarHandler.CreateCar},
			{Method: http.MethodPut, Path: "/:id", Handler: container.CarHandler.UpdateCar},
			{Method: http.MethodDelete, Path: "/:id", Handler: container.CarHandler.DeleteCar},
		}},
		{Prefix: "/Cars", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/GetAll", Handler: container.CarHandler.GetAll},
			{Method: http.MethodGet, Path: "/GetById/", Handler: container.CarHandler.GetById},
			{Method: http.MethodGet, Path: "/GetFiltered", Handler: container.CarHandler.GetFiltered},
			{Method: http.MethodPost, Path: "/Create", Handler: container.CarHandler.CreateCar},
			{Method: http.MethodPut, Path: "/Update/", Handler: container.CarHandler.UpdateCar},
			{Method: http.MethodDelete, Path: "/Delete/", Handler: container.CarHandler.DeleteCar},
		}},
		{Prefix: "/comments", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/:carId", Handler: container.CommentHandler.GetCarComments},
			{Method: http.MethodPost, Path: "/", Handler: container.CommentHandler.CreateComment},
			{Method: http.MethodPut, Path: "/:id", Handler: container.CommentHandler.UpdateComment},
			{Method: http.MethodDelete, Path: "/:id", Handler: container.CommentHandler.DeleteComment},
		}},
		{Prefix: "/Comments", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/GetCarComments/", Handler: container.CommentHandler.GetCarComments},
			{Method: http.MethodPost, Path: "/Create", Handler: container.CommentHandler.CreateComment},
			{Method: http.MethodPut, Path: "/Update/", Handler: container.CommentHandler.UpdateComment},
			{Method: http.MethodDelete, Path: "/Delete/", Handler: container.CommentHandler.DeleteComment},
		}},
		{Prefix: "/auth", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/register", Handler: container.AuthHandler.Register},
			{Method: http.MethodPost, Path: "/login", Handler: container.AuthHandler.Login},
			{Method: http.MethodGet, Path: "/profile", Handler: container.AuthHandler.Profile},
			{Method: http.MethodPut, Path: "/profile", Handler: container.AuthHandler.UpdateProfile},
			{Method: http.MethodPut, Path: "/password", Handler: container.AuthHandler.ChangePassword},
		}},
		{Prefix: "/Auth", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/Register", Handler: container.AuthHandler.Register},
			{Method: http.MethodPost, Path: "/Login", Handler: container.AuthHandler.Login},
			{Method: http.MethodGet, Path: "/Profile", Handler: container.AuthHandler.Profile},
			{Method: http.MethodPut, Path: "/Profile", Handler: container.AuthHandler.UpdateProfile},
			{Method: http.MethodPut, Path: "/Password", Handler: container.AuthHandler.ChangePassword},
		}},
		{Prefix: "/favorites", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: container.FavoriteHandler.List},
			{Method: http.MethodPost, Path: "/", Handler: container.FavoriteHandler.Add},
			{Method: http.MethodDelete, Path: "/:carId", Handler: container.FavoriteHandler.Remove},
		}},
		{Prefix: "/Favorites", Routes: []server.Route{
			{Method: http.MethodGet, Path: "", Handler: container.FavoriteHandler.List},
			{Method: http.MethodPost, Path: "", Handler: container.FavoriteHandler.Add},
			{Method: http.MethodDelete, Path: "/", Handler: container.FavoriteHandler.Remove},
		}},
		{Prefix: "/rentals", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: container.RentalHandler.List},
			{Method: http.MethodPost, Path: "/", Handler: container.RentalHandler.Create},
			{Method: http.MethodGet, Path: "/car/:id", Handler: container.RentalHandler.BookedDates},
		}},
		{Prefix: "/Rentals", Routes: []server.Route{
			{Method: http.MethodGet, Path: "", Handler: container.RentalHandler.List},
			{Method: http.MethodPost, Path: "", Handler: container.RentalHandler.Create},
			{Method: http.MethodGet, Path: "/Car/", Handler: container.RentalHandler.BookedDates},
		}},
		{Prefix: "/media", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/upload", Handler: container.MediaHandler.Upload},
		}},
		{Prefix: "/Media", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/Upload", Handler: container.MediaHandler.Upload},
		}},
	}

	server.RegisterRouteGroupsToMux(mux, groups, common.WrapCORS)
}
