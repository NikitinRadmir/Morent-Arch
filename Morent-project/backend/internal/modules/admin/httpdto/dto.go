package httpdto

// UpdateUserRequest — только разрешённые поля (без passwordHash, role escalation через произвольный JSON).
type UpdateUserRequest struct {
	ID        uint    `json:"id"`
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
	Position  *string `json:"position,omitempty"`
	AvatarURL *string `json:"avatarUrl,omitempty"`
	Role      *string `json:"role,omitempty"`
}

// UpdateCommentRequest — только контент отзыва.
type UpdateCommentRequest struct {
	ID          uint    `json:"id"`
	Description *string `json:"description,omitempty"`
	Rating      *int    `json:"rating,omitempty"`
}
