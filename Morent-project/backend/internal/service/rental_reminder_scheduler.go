package service

import (
	"context"
	"log/slog"
	"time"

	"morent-backend/internal/repository"
)

// RentalReminderScheduler отправляет напоминания в день начала аренды.
type RentalReminderScheduler struct {
	rentals *repository.RentalRepository
	users   *repository.UserRepository
	emails  *EmailNotifier
	log     *slog.Logger
}

func NewRentalReminderScheduler(
	rentals *repository.RentalRepository,
	users *repository.UserRepository,
	emails *EmailNotifier,
	log *slog.Logger,
) *RentalReminderScheduler {
	if log == nil {
		log = slog.Default()
	}
	return &RentalReminderScheduler{
		rentals: rentals,
		users:   users,
		emails:  emails,
		log:     log,
	}
}

func (s *RentalReminderScheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	s.tick()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick()
		}
	}
}

func (s *RentalReminderScheduler) tick() {
	now := time.Now()
	rentals, err := s.rentals.ListStartingTodayWithoutReminder(now)
	if err != nil {
		s.log.Warn("rental reminder query failed", "error", err)
		return
	}
	for i := range rentals {
		rental := rentals[i]
		user, err := s.users.GetByID(rental.UserID)
		if err != nil || user == nil {
			s.log.Warn("rental reminder skipped: user not found", "rental_id", rental.ID, "user_id", rental.UserID)
			continue
		}
		s.emails.NotifyRentalDayReminder(user, &rental)
		if err := s.rentals.MarkRentalDayReminderSent(rental.ID); err != nil {
			s.log.Warn("rental reminder mark failed", "rental_id", rental.ID, "error", err)
			continue
		}
		s.log.Info("rental day reminder queued", "rental_id", rental.ID, "user_id", user.ID)
	}
}
