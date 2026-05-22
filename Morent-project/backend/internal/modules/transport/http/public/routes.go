package public

import (
	"net/http"

	"morent-backend/internal/di"
	"morent-backend/internal/modules/transport/http/common"
	"morent-backend/internal/server"
)

func Register(mux *http.ServeMux, container *di.Container) {
	auth := func(h http.HandlerFunc) http.HandlerFunc {
		return common.WrapCORS(common.WithAuth(container.AuthService, container.Config, h))
	}

	groups := []server.RouteGroup{
		{Prefix: "/cars", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: container.CarHandler.GetAll},
			{Method: http.MethodGet, Path: "/:id", Handler: container.CarHandler.GetById},
			{Method: http.MethodGet, Path: "/filtered", Handler: container.CarHandler.GetFiltered},
		}},
		{Prefix: "/Cars", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/GetAll", Handler: container.CarHandler.GetAll},
			{Method: http.MethodGet, Path: "/GetById/", Handler: container.CarHandler.GetById},
			{Method: http.MethodGet, Path: "/GetFiltered", Handler: container.CarHandler.GetFiltered},
		}},
		{Prefix: "/comments", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/:carId", Handler: container.CommentHandler.GetCarComments},
			{Method: http.MethodPost, Path: "/", Handler: auth(container.CommentHandler.CreateComment)},
			{Method: http.MethodPut, Path: "/:id", Handler: auth(container.CommentHandler.UpdateComment)},
			{Method: http.MethodDelete, Path: "/:id", Handler: auth(container.CommentHandler.DeleteComment)},
		}},
		{Prefix: "/Comments", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/GetCarComments/", Handler: container.CommentHandler.GetCarComments},
			{Method: http.MethodPost, Path: "/Create", Handler: auth(container.CommentHandler.CreateComment)},
			{Method: http.MethodPut, Path: "/Update/", Handler: auth(container.CommentHandler.UpdateComment)},
			{Method: http.MethodDelete, Path: "/Delete/", Handler: auth(container.CommentHandler.DeleteComment)},
		}},
		{Prefix: "/auth", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/password/generate", Handler: container.PasswordHandler.Generate},
			{Method: http.MethodPost, Path: "/password/validate", Handler: container.PasswordHandler.Validate},
			{Method: http.MethodPost, Path: "/register", Handler: container.AuthHandler.Register},
			{Method: http.MethodPost, Path: "/login", Handler: container.AuthHandler.Login},
			{Method: http.MethodPost, Path: "/logout", Handler: container.AuthHandler.Logout},
			{Method: http.MethodGet, Path: "/profile", Handler: container.AuthHandler.Profile},
			{Method: http.MethodPut, Path: "/profile", Handler: container.AuthHandler.UpdateProfile},
			{Method: http.MethodPut, Path: "/password", Handler: container.AuthHandler.ChangePassword},
			{Method: http.MethodPost, Path: "/verify-email", Handler: container.AuthHandler.VerifyEmail},
			{Method: http.MethodPost, Path: "/verify-email/resend", Handler: container.AuthHandler.ResendVerificationEmail},
		}},
		{Prefix: "/Auth", Routes: []server.Route{
			{Method: http.MethodPost, Path: "/Register", Handler: container.AuthHandler.Register},
			{Method: http.MethodPost, Path: "/Login", Handler: container.AuthHandler.Login},
			{Method: http.MethodPost, Path: "/Logout", Handler: container.AuthHandler.Logout},
			{Method: http.MethodGet, Path: "/Profile", Handler: container.AuthHandler.Profile},
			{Method: http.MethodPut, Path: "/Profile", Handler: container.AuthHandler.UpdateProfile},
			{Method: http.MethodPut, Path: "/Password", Handler: container.AuthHandler.ChangePassword},
		}},
		{Prefix: "/favorites", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: auth(container.FavoriteHandler.List)},
			{Method: http.MethodPost, Path: "/", Handler: auth(container.FavoriteHandler.Add)},
			{Method: http.MethodDelete, Path: "/:carId", Handler: auth(container.FavoriteHandler.Remove)},
		}},
		{Prefix: "/Favorites", Routes: []server.Route{
			{Method: http.MethodGet, Path: "", Handler: auth(container.FavoriteHandler.List)},
			{Method: http.MethodPost, Path: "", Handler: auth(container.FavoriteHandler.Add)},
			{Method: http.MethodDelete, Path: "/", Handler: auth(container.FavoriteHandler.Remove)},
		}},
		{Prefix: "/rentals", Routes: []server.Route{
			{Method: http.MethodGet, Path: "/", Handler: auth(container.RentalHandler.List)},
			{Method: http.MethodPost, Path: "/", Handler: auth(container.RentalHandler.Create)},
			{Method: http.MethodGet, Path: "/car/:id", Handler: container.RentalHandler.BookedDates},
		}},
		{Prefix: "/Rentals", Routes: []server.Route{
			{Method: http.MethodGet, Path: "", Handler: auth(container.RentalHandler.List)},
			{Method: http.MethodPost, Path: "", Handler: auth(container.RentalHandler.Create)},
			{Method: http.MethodGet, Path: "/Car/", Handler: container.RentalHandler.BookedDates},
		}},
	}

	server.RegisterRouteGroupsToMux(mux, groups, common.WrapCORS)
}
