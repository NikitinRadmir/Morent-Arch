package service

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	morentevents "morent-events"
	"morent-backend/internal/messaging"
	"morent-backend/internal/models"
	"morent-backend/internal/repository"
)

type EmailNotifier struct {
	users  *repository.UserRepository
	emails messaging.EmailEventPublisher
	log    *slog.Logger
}

func NewEmailNotifier(users *repository.UserRepository, emails messaging.EmailEventPublisher, log *slog.Logger) *EmailNotifier {
	if emails == nil {
		emails = messaging.EmailNoopPublisher{}
	}
	if log == nil {
		log = slog.Default()
	}
	return &EmailNotifier{users: users, emails: emails, log: log}
}

func (n *EmailNotifier) NotifyWelcomeRegistered(user *models.User) {
	if user == nil || !n.emails.Enabled() {
		return
	}
	n.publishAsync(user.Email, morentevents.EmailTemplateWelcomeRegistered, map[string]string{
		"user_name":  user.Name,
		"user_email": user.Email,
	})
}

func (n *EmailNotifier) NotifyLogin(user *models.User) {
	if user == nil || !n.emails.Enabled() {
		return
	}
	n.publishAsync(user.Email, morentevents.EmailTemplateLoginNotification, map[string]string{
		"user_name":  user.Name,
		"user_email": user.Email,
		"login_time": time.Now().Format("2006-01-02 15:04 MST"),
	})
}

func (n *EmailNotifier) publishAsync(to, templateKey string, vars map[string]string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := n.emails.PublishEmailSend(ctx, messaging.EmailSendEvent{
			To:          to,
			TemplateKey: templateKey,
			Variables:   vars,
			Priority:    "Normal",
		}); err != nil {
			n.log.Warn("failed to publish email", "template", templateKey, "to", to, "error", err)
		}
	}()
}

func (n *EmailNotifier) NotifyBookingConfirmation(userID uint, rental *models.RentalResponse) {
	if rental == nil || !n.emails.Enabled() {
		return
	}
	go func() {
		user, err := n.users.GetByID(userID)
		if err != nil || user == nil {
			n.log.Warn("booking email skipped: user not found", "user_id", userID, "error", err)
			return
		}

		carName := strings.TrimSpace(rental.Car.Name)
		if carName == "" {
			carName = fmt.Sprintf("авто #%d", rental.Car.ID)
		}

		n.publishAsync(user.Email, morentevents.EmailTemplateBookingConfirmation, map[string]string{
			"user_name":   user.Name,
			"car_name":    carName,
			"start_date":  rental.StartDate.Format("2006-01-02"),
			"end_date":    rental.EndDate.Format("2006-01-02"),
			"total_price": strconv.FormatFloat(rental.TotalPrice, 'f', 2, 64),
			"rental_id":   strconv.FormatUint(uint64(rental.ID), 10),
		})
	}()
}
