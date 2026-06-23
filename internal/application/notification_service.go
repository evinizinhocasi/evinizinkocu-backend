package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/fcm"
)

type NotificationService struct {
	repo        domain.NotificationRepository
	studentRepo domain.StudentRepository
	userRepo    domain.UserRepository
	coachRepo   domain.CoachRepository
	fcmService  fcm.FCMService
}

func NewNotificationService(
	repo domain.NotificationRepository,
	studentRepo domain.StudentRepository,
	userRepo domain.UserRepository,
	coachRepo domain.CoachRepository,
	fcmService fcm.FCMService,
) *NotificationService {
	return &NotificationService{
		repo:        repo,
		studentRepo: studentRepo,
		userRepo:    userRepo,
		coachRepo:   coachRepo,
		fcmService:  fcmService,
	}
}

// In-app notifications inbox

func (s *NotificationService) GetInbox(ctx context.Context, userID string) ([]*domain.NotificationRecipient, error) {
	return s.repo.GetInbox(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, recipientID string) error {
	return s.repo.MarkAsRead(ctx, recipientID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

// Manual immediate notification sending
func (s *NotificationService) SendImmediateNotification(
	ctx context.Context,
	coachID, title, body, targetSelection string,
	targetStudentIDs []string,
) (*domain.Notification, error) {
	// 1. Verify Coach status and permissions
	coach, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return nil, domain.ErrCoachNotFound
	}
	if !coach.User.IsActive {
		return nil, domain.ErrPassiveAccount
	}
	if !coach.PermissionImmediatePush {
		return nil, errors.New("manuel bildirim gönderme yetkiniz yok")
	}

	// 2. Resolve recipients
	recipients, err := s.resolveStudents(ctx, coachID, targetSelection, targetStudentIDs)
	if err != nil {
		return nil, err
	}

	if len(recipients) == 0 {
		return nil, errors.New("seçilen kriterlere uygun alıcı bulunamadı")
	}

	// Save notification in database (inbox entry)
	n := &domain.Notification{
		Title:    title,
		Body:     body,
		SenderID: coachID,
	}

	err = s.repo.CreateNotification(ctx, n, recipients)
	if err != nil {
		return nil, err
	}

	// Deliver push messages asynchronously
	go func() {
		for _, rID := range recipients {
			tokens, err := s.userRepo.GetDeviceTokensByUser(context.Background(), rID)
			if err == nil && len(tokens) > 0 {
				_ = s.fcmService.SendPushToTokens(context.Background(), tokens, title, body)
			}
		}
	}()

	return n, nil
}

// Schedules management

func (s *NotificationService) CreateSchedule(
	ctx context.Context,
	coachID, title, body, targetSelection string,
	targetStudentIDs []string,
	scheduleType string,
	selectedWeekdays []int,
	scheduleTime string, // HH:MM:SS
	start, end *time.Time,
) (*domain.NotificationSchedule, error) {
	// 1. Verify permissions
	coach, err := s.coachRepo.GetCoachByID(ctx, coachID)
	if err != nil {
		return nil, domain.ErrCoachNotFound
	}
	if !coach.User.IsActive {
		return nil, domain.ErrPassiveAccount
	}
	if !coach.PermissionScheduledPush {
		return nil, errors.New("planlanmış bildirim oluşturma yetkiniz yok")
	}

	// 2. Validate parameters
	if scheduleTime == "" {
		return nil, errors.New("planlanan saat zorunludur")
	}

	// Calculate next run date/time interpreted in Europe/Istanbul timezone
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		// fallback to UTC+3 manually if load location fails
		loc = time.FixedZone("Europe/Istanbul", 3*60*60)
	}

	nowLocal := time.Now().In(loc)
	nextRun, err := s.CalculateNextRun(nowLocal, scheduleType, selectedWeekdays, scheduleTime, start)
	if err != nil {
		return nil, err
	}

	ns := &domain.NotificationSchedule{
		CoachID:          coachID,
		Title:            title,
		Body:             body,
		TargetSelection:  targetSelection,
		TargetStudentIDs: targetStudentIDs,
		ScheduleType:     scheduleType,
		SelectedWeekdays: selectedWeekdays,
		ScheduleTime:     scheduleTime,
		StartDate:        start,
		EndDate:          end,
		NextRunAt:        nextRun.UTC(), // Save in UTC
		IsActive:         true,
	}

	err = s.repo.CreateSchedule(ctx, ns)
	return ns, err
}

func (s *NotificationService) ListSchedulesByCoach(ctx context.Context, coachID string) ([]*domain.NotificationSchedule, error) {
	return s.repo.ListSchedulesByCoach(ctx, coachID)
}

func (s *NotificationService) DeleteSchedule(ctx context.Context, id string) error {
	return s.repo.DeleteSchedule(ctx, id)
}

// Helpers
func (s *NotificationService) resolveStudents(ctx context.Context, coachID, targetSelection string, targetStudentIDs []string) ([]string, error) {
	var ids []string
	if targetSelection == "one" || targetSelection == "selected" {
		// Verify ownership of the targets
		for _, sID := range targetStudentIDs {
			student, err := s.studentRepo.GetStudentByID(ctx, sID)
			if err == nil && student.CoachID == coachID && student.User.IsActive {
				ids = append(ids, sID)
			}
		}
	} else if targetSelection == "all" {
		// Fetch all students belonging to the coach
		students, err := s.studentRepo.ListStudentsByCoach(ctx, coachID, false)
		if err != nil {
			return nil, err
		}
		for _, stud := range students {
			if stud.User.IsActive {
				ids = append(ids, stud.ID)
			}
		}
	}
	return ids, nil
}

func (s *NotificationService) CalculateNextRun(
	nowLocal time.Time,
	scheduleType string,
	weekdays []int,
	schedTimeStr string,
	startDate *time.Time,
) (time.Time, error) {
	// Parse schedule time (HH:MM:SS)
	var hour, min, sec int
	_, err := fmt.Sscanf(schedTimeStr, "%d:%d:%d", &hour, &min, &sec)
	if err != nil {
		// Try HH:MM format
		_, err = fmt.Sscanf(schedTimeStr, "%d:%d", &hour, &min)
		if err != nil {
			return time.Time{}, errors.New("planlanan saat formatı geçersiz")
		}
		sec = 0
	}

	loc := nowLocal.Location()

	// Candidate day starts from nowLocal or startDate
	candidateDate := nowLocal
	if startDate != nil && startDate.After(nowLocal) {
		candidateDate = startDate.In(loc)
	}

	candidate := time.Date(candidateDate.Year(), candidateDate.Month(), candidateDate.Day(), hour, min, sec, 0, loc)

	switch scheduleType {
	case "one_time":
		if candidate.Before(nowLocal) {
			// Already passed today, push to tomorrow or error. Since one_time, it should be in the future
			return time.Time{}, errors.New("geçmiş bir zamana tek seferlik planlama yapılamaz")
		}
		return candidate, nil

	case "daily":
		if candidate.Before(nowLocal) {
			candidate = candidate.AddDate(0, 0, 1)
		}
		return candidate, nil

	case "weekly":
		if len(weekdays) == 0 {
			return time.Time{}, errors.New("haftalık planlama için gün seçimi zorunludur")
		}

		// Find the next day in selection
		// Weekday representation: 1 = Monday, 7 = Sunday
		// In Go: 0 = Sunday, 1 = Monday, 6 = Saturday
		for i := 0; i < 8; i++ {
			testDay := candidate.AddDate(0, 0, i)
			if i == 0 && testDay.Before(nowLocal) {
				continue
			}

			// Map Go weekday to Turkish 1-7
			goWD := testDay.Weekday()
			trWD := int(goWD)
			if trWD == 0 {
				trWD = 7 // Sunday
			}

			for _, wd := range weekdays {
				if wd == trWD {
					return testDay, nil
				}
			}
		}
		// Fallback
		return candidate, nil
	}

	return time.Time{}, errors.New("geçersiz planlama türü")
}

func (s *NotificationService) GetScheduleByID(ctx context.Context, id string) (*domain.NotificationSchedule, error) {
	return s.repo.GetScheduleByID(ctx, id)
}

// Custom format helper
