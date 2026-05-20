package messaging

import "context"

// UserEventPublisher публикует события пользователя в Kafka (без plaintext-пароля).
type UserEventPublisher interface {
	PublishUserRegistered(ctx context.Context, in UserRegisteredEvent) error
	PublishUserProfileUpdated(ctx context.Context, in UserProfileUpdatedEvent) error
	PublishUserPasswordChanged(ctx context.Context, in UserPasswordChangedEvent) error
	PublishUserDeactivated(ctx context.Context, in UserDeactivatedEvent) error
	Enabled() bool
}

type UserRegisteredEvent struct {
	MorentUserID uint
	Email        string
	Name         string
	Nickname     string
	Position     string
	AvatarURL    string
	Role         string
	PasswordHash string
	CompanyName  string
}

type UserProfileUpdatedEvent struct {
	MorentUserID uint
	Email        string
	Name         *string
	Nickname     *string
	Position     *string
	AvatarURL    *string
	Role         *string
	IsActive     *bool
}

type UserPasswordChangedEvent struct {
	MorentUserID uint
	Email        string
	PasswordHash string
}

type UserDeactivatedEvent struct {
	MorentUserID uint
	Email        string
}

// NoopPublisher — заглушка, если Kafka отключён.
type NoopPublisher struct{}

func (NoopPublisher) Enabled() bool { return false }

func (NoopPublisher) PublishUserRegistered(context.Context, UserRegisteredEvent) error {
	return nil
}

func (NoopPublisher) PublishUserProfileUpdated(context.Context, UserProfileUpdatedEvent) error {
	return nil
}

func (NoopPublisher) PublishUserPasswordChanged(context.Context, UserPasswordChangedEvent) error {
	return nil
}

func (NoopPublisher) PublishUserDeactivated(context.Context, UserDeactivatedEvent) error {
	return nil
}
