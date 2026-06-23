package domain

import (
	"context"
	"time"
)

type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	SenderID  string    `json:"sender_id"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationRecipient struct {
	ID             string     `json:"id"`
	NotificationID string     `json:"notification_id"`
	RecipientID    string     `json:"recipient_id"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type NotificationSchedule struct {
	ID                string     `json:"id"`
	CoachID           string     `json:"coach_id"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	TargetSelection   string     `json:"target_selection"` // 'one', 'selected', 'all'
	TargetStudentIDs  []string   `json:"target_student_ids"`
	ScheduleType      string     `json:"schedule_type"`    // 'one_time', 'daily', 'weekly'
	SelectedWeekdays  []int      `json:"selected_weekdays"` // 1 = Mon, 7 = Sun
	ScheduleTime      string     `json:"schedule_time"`     // e.g. "14:30:00"
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	NextRunAt         time.Time  `json:"next_run_at"`
	IsActive          bool       `json:"is_active"`
	LockedBy          *string    `json:"locked_by,omitempty"`
	LockedUntil       *time.Time `json:"locked_until,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type NotificationExecution struct {
	ID           string    `json:"id"`
	ScheduleID   *string   `json:"schedule_id,omitempty"`
	Status       string    `json:"status"` // 'success', 'failed'
	ErrorMessage string    `json:"error_message,omitempty"`
	ExecutedAt   time.Time `json:"executed_at"`
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, n *Notification, recipientIDs []string) error
	GetInbox(ctx context.Context, userID string) ([]*NotificationRecipient, error)
	MarkAsRead(ctx context.Context, recipientID string) error
	MarkAllAsRead(ctx context.Context, userID string) error

	// Schedules
	CreateSchedule(ctx context.Context, ns *NotificationSchedule) error
	GetScheduleByID(ctx context.Context, id string) (*NotificationSchedule, error)
	UpdateSchedule(ctx context.Context, ns *NotificationSchedule) error
	DeleteSchedule(ctx context.Context, id string) error
	ListSchedulesByCoach(ctx context.Context, coachID string) ([]*NotificationSchedule, error)

	// Worker claim
	ClaimDueSchedules(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]*NotificationSchedule, error)
	SaveExecutionRecord(ctx context.Context, exec *NotificationExecution) error
}
