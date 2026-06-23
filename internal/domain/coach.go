package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrCoachNotFound            = errors.New("coach not found")
	ErrCoachApplicationNotFound = errors.New("coach application not found")
	ErrCapacityExceeded         = errors.New("coach capacity has been reached")
	ErrCoachExpired             = errors.New("coach authorization period has expired")
)

type Coach struct {
	ID                       string    `json:"id"`
	User                     *User     `json:"user,omitempty"`
	City                     string    `json:"city"`
	Biography                string    `json:"biography"`
	Specialization           string    `json:"specialization"`
	SocialLinks              string    `json:"social_links"` // JSON string representation
	StudentCapacity          int       `json:"student_capacity"`
	AuthStartDate            time.Time `json:"auth_start_date"`
	AuthEndDate              time.Time `json:"auth_end_date"`
	PermissionImmediatePush  bool      `json:"permission_immediate_push"`
	PermissionScheduledPush  bool      `json:"permission_scheduled_push"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

type CoachApplication struct {
	ID             string    `json:"id"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Phone          string    `json:"phone"`
	Email          string    `json:"email"`
	City           string    `json:"city"`
	Specialization string    `json:"specialization"`
	Explanation    string    `json:"explanation"`
	Status         string    `json:"status"` // 'pending', 'approved', 'rejected'
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CoachAuthorizationPeriod struct {
	ID              string    `json:"id"`
	CoachID         string    `json:"coach_id"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	StudentCapacity int       `json:"student_capacity"`
	CreatedAt       time.Time `json:"created_at"`
}

type CoachRepository interface {
	CreateCoach(ctx context.Context, coach *Coach) error
	GetCoachByID(ctx context.Context, id string) (*Coach, error)
	UpdateCoach(ctx context.Context, coach *Coach) error
	ListCoaches(ctx context.Context) ([]*Coach, error)

	// Applications
	CreateApplication(ctx context.Context, app *CoachApplication) error
	GetApplicationByID(ctx context.Context, id string) (*CoachApplication, error)
	UpdateApplicationStatus(ctx context.Context, id string, status string) error
	ListApplications(ctx context.Context) ([]*CoachApplication, error)

	// History/Logs
	SaveAuthorizationPeriod(ctx context.Context, ap *CoachAuthorizationPeriod) error
	GetAuthorizationPeriods(ctx context.Context, coachID string) ([]*CoachAuthorizationPeriod, error)

	// Expiry check
	GetExpiredCoaches(ctx context.Context) ([]string, error)
	DeactivateCoaches(ctx context.Context, ids []string) error
}
