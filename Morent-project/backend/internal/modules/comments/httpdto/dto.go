package httpdto

type CreateCommentRequest struct {
	CarID       uint   `json:"carId" validate:"required"`
	Description string `json:"description" validate:"required,min=1,max=1000"`
	Rating      int    `json:"rating" validate:"required,min=1,max=5"`
}

type UpdateCommentBody struct {
	Description *string `json:"description,omitempty"`
	Rating      *int    `json:"rating,omitempty"`
}
