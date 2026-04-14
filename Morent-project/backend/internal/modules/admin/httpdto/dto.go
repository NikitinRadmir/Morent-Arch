package httpdto

import "morent-backend/internal/models"

type UpdateUserRequest struct {
	models.User
}

type UpdateCommentRequest struct {
	models.Comment
}
