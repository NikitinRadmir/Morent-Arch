package httpdto

type CreateCommentRequest struct {
	CarID       uint   `json:"carId" validate:"required"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
	Rating      int    `json:"rating" validate:"required,min=1,max=5"`
}
