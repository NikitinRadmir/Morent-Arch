package morentevents

// Типы событий (Morent → user-system).
const (
	EventUserRegistered     = "morent.user.registered"
	EventUserProfileUpdated = "morent.user.profile_updated"
	EventUserPasswordChanged = "morent.user.password_changed"
	EventUserDeactivated    = "morent.user.deactivated"
)

// UserRegistered — пользователь создан в Morent (пароль только как bcrypt-хеш).
type UserRegistered struct {
	MorentUserID uint   `json:"morent_user_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Nickname     string `json:"nickname,omitempty"`
	Position     string `json:"position,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
	Role         string `json:"role"`
	// PasswordHash — уже захешированный пароль (bcrypt), plaintext запрещён.
	PasswordHash string `json:"password_hash"`
	CompanyName  string `json:"company_name"`
}

// UserProfileUpdated — обновление профиля в Morent.
type UserProfileUpdated struct {
	MorentUserID uint    `json:"morent_user_id"`
	Email        string  `json:"email"`
	Name         *string `json:"name,omitempty"`
	Nickname     *string `json:"nickname,omitempty"`
	Position     *string `json:"position,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Role         *string `json:"role,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// UserPasswordChanged — смена пароля в Morent (только bcrypt-хеш).
type UserPasswordChanged struct {
	MorentUserID uint   `json:"morent_user_id"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

// UserDeactivated — деактивация/удаление в Morent.
type UserDeactivated struct {
	MorentUserID uint   `json:"morent_user_id"`
	Email        string `json:"email"`
}
