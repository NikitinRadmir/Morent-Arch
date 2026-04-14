package comments

import (
	"morent-backend/internal/handlers"
	"morent-backend/internal/repository"
	"morent-backend/internal/service"

	"go.uber.org/fx"
)

type Outputs struct {
	fx.Out
	Service *service.CommentService
	Handler *handlers.CommentHandler
}

func NewModule(
	commentRepo *repository.CommentRepository,
	rentalRepo *repository.RentalRepository,
	authService *service.AuthService,
) Outputs {
	commentService := service.NewCommentService(commentRepo, rentalRepo)
	commentHandler := handlers.NewCommentHandler(commentService, authService)

	return Outputs{
		Service: commentService,
		Handler: commentHandler,
	}
}
