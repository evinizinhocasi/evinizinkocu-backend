package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"evinizinkocu-backend/internal/application"
	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/fcm"
)

type NotificationWorker struct {
	repo         domain.NotificationRepository
	studentRepo  domain.StudentRepository
	userRepo     domain.UserRepository
	coachRepo    domain.CoachRepository
	fcmService   fcm.FCMService
	notifService *application.NotificationService
	workerID     string
	stopChan     chan struct{}
}

func NewNotificationWorker(
	repo domain.NotificationRepository,
	studentRepo domain.StudentRepository,
	userRepo domain.UserRepository,
	coachRepo domain.CoachRepository,
	fcmService fcm.FCMService,
	notifService *application.NotificationService,
) *NotificationWorker {
	// Simple random worker ID to support multi-instance safely
	rand.Seed(time.Now().UnixNano())
	workerID := fmt.Sprintf("worker-%d", rand.Intn(100000))

	return &NotificationWorker{
		repo:         repo,
		studentRepo:  studentRepo,
		userRepo:     userRepo,
		coachRepo:    coachRepo,
		fcmService:   fcmService,
		notifService: notifService,
		workerID:     workerID,
		stopChan:     make(chan struct{}),
	}
}

func (w *NotificationWorker) Start(interval time.Duration) {
	log.Printf("Starting Notification Scheduled Worker [%s]...\n", w.workerID)
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				w.ProcessSchedules()
			case <-w.stopChan:
				ticker.Stop()
				log.Printf("Stopping Scheduled Worker [%s]...\n", w.workerID)
				return
			}
		}
	}()
}

func (w *NotificationWorker) Stop() {
	close(w.stopChan)
}

func (w *NotificationWorker) ProcessSchedules() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Claim up to 10 due schedules at a time, lock them for 1 minute
	schedules, err := w.repo.ClaimDueSchedules(ctx, w.workerID, 10, 1*time.Minute)
	if err != nil {
		log.Printf("Worker [%s] failed claiming due schedules: %v\n", w.workerID, err)
		return
	}

	if len(schedules) == 0 {
		return
	}

	log.Printf("Worker [%s] claimed %d due scheduled notifications\n", w.workerID, len(schedules))

	for _, sched := range schedules {
		w.processSingleSchedule(ctx, sched)
	}
}

func (w *NotificationWorker) processSingleSchedule(ctx context.Context, sched *domain.NotificationSchedule) {
	log.Printf("Worker [%s] processing schedule ID: %s (%s)\n", w.workerID, sched.ID, sched.Title)

	// 1. Verify Coach status and notification permissions
	coach, err := w.coachRepo.GetCoachByID(ctx, sched.CoachID)
	if err != nil {
		w.handleScheduleError(ctx, sched, fmt.Sprintf("Coach not found: %v", err))
		return
	}

	if !coach.User.IsActive {
		// Stop schedule permanently if coach is passive
		sched.IsActive = false
		w.handleScheduleError(ctx, sched, "Coach is inactive")
		return
	}

	if !coach.PermissionScheduledPush {
		// Stop schedule permanently if scheduled permission is disabled
		sched.IsActive = false
		w.handleScheduleError(ctx, sched, "Scheduled notifications permission revoked")
		return
	}

	// 2. Resolve target students (excl. passive or archived)
	var activeRecipients []string
	var deviceTokens []string

	if sched.TargetSelection == "one" || sched.TargetSelection == "selected" {
		for _, sID := range sched.TargetStudentIDs {
			student, err := w.studentRepo.GetStudentByID(ctx, sID)
			if err == nil && student.CoachID == sched.CoachID && student.User.IsActive && !student.IsArchived {
				activeRecipients = append(activeRecipients, sID)
				tokens, _ := w.userRepo.GetDeviceTokensByUser(ctx, sID)
				deviceTokens = append(deviceTokens, tokens...)
			}
		}
	} else if sched.TargetSelection == "all" {
		students, err := w.studentRepo.ListStudentsByCoach(ctx, sched.CoachID, false)
		if err == nil {
			for _, student := range students {
				if student.User.IsActive && !student.IsArchived {
					activeRecipients = append(activeRecipients, student.ID)
					tokens, _ := w.userRepo.GetDeviceTokensByUser(ctx, student.ID)
					deviceTokens = append(deviceTokens, tokens...)
				}
			}
		}
	}

	if len(activeRecipients) == 0 {
		// Skip delivery, update next run
		w.updateNextScheduleRun(ctx, sched, "Success (no active recipients)")
		return
	}

	// 3. Create In-App Notification (Inbox item)
	n := &domain.Notification{
		Title:    sched.Title,
		Body:     sched.Body,
		SenderID: sched.CoachID,
	}

	err = w.repo.CreateNotification(ctx, n, activeRecipients)
	if err != nil {
		w.handleScheduleError(ctx, sched, fmt.Sprintf("Failed saving in-app notification: %v", err))
		return
	}

	// 4. Send Push Notifications via FCM
	if len(deviceTokens) > 0 {
		// Deliver FCM in background
		go func(tokens []string, title, body string) {
			_ = w.fcmService.SendPushToTokens(context.Background(), tokens, title, body)
		}(deviceTokens, sched.Title, sched.Body)
	}

	w.updateNextScheduleRun(ctx, sched, "Success")
}

func (w *NotificationWorker) handleScheduleError(ctx context.Context, sched *domain.NotificationSchedule, errMsg string) {
	log.Printf("Worker [%s] schedule error for %s: %s\n", w.workerID, sched.ID, errMsg)

	exec := &domain.NotificationExecution{
		ScheduleID:   &sched.ID,
		Status:       "failed",
		ErrorMessage: errMsg,
	}
	_ = w.repo.SaveExecutionRecord(ctx, exec)

	// Clean locks and update
	sched.LockedBy = nil
	sched.LockedUntil = nil
	_ = w.repo.UpdateSchedule(ctx, sched)
}

func (w *NotificationWorker) updateNextScheduleRun(ctx context.Context, sched *domain.NotificationSchedule, status string) {
	exec := &domain.NotificationExecution{
		ScheduleID: &sched.ID,
		Status:     "success",
	}
	if status != "Success" {
		exec.ErrorMessage = status
	}
	_ = w.repo.SaveExecutionRecord(ctx, exec)

	// Calculate next execution run
	loc, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		loc = time.FixedZone("Europe/Istanbul", 3*60*60)
	}

	nowLocal := time.Now().In(loc)

	// Calculate next local run time based on type
	nextRun, err := w.notifService.CalculateNextRun(nowLocal, sched.ScheduleType, sched.SelectedWeekdays, sched.ScheduleTime, sched.StartDate)
	if err != nil || (sched.EndDate != nil && nextRun.After(*sched.EndDate)) || sched.ScheduleType == "one_time" {
		// Disable schedule permanently
		sched.IsActive = false
	} else {
		sched.NextRunAt = nextRun.UTC()
	}

	sched.LockedBy = nil
	sched.LockedUntil = nil
	_ = w.repo.UpdateSchedule(ctx, sched)
}
