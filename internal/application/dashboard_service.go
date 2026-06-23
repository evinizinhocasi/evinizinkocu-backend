package application

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardService struct {
	db *pgxpool.Pool
}

func NewDashboardService(db *pgxpool.Pool) *DashboardService {
	return &DashboardService{db: db}
}

type SuperadminDashboard struct {
	TotalCoaches             int     `json:"total_coaches"`
	ActiveCoaches            int     `json:"active_coaches"`
	PassiveCoaches           int     `json:"passive_coaches"`
	TotalStudents            int     `json:"total_students"`
	ActiveStudents           int     `json:"active_students"`
	PassiveStudents          int     `json:"passive_students"`
	CoachesNearExpiration    int     `json:"coaches_near_expiration"`
	CoachesAtOrOverCapacity  int     `json:"coaches_at_or_over_capacity"`
	PendingApplicationsCount int     `json:"pending_applications_count"`
	TotalPaymentsReceived    float64 `json:"total_payments_received"`
}

func (s *DashboardService) GetSuperadminDashboard(ctx context.Context) (*SuperadminDashboard, error) {
	var dbInfo SuperadminDashboard

	// Coaches counts
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM coaches").Scan(&dbInfo.TotalCoaches)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM coaches c JOIN users u ON c.id = u.id WHERE u.is_active = TRUE").Scan(&dbInfo.ActiveCoaches)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM coaches c JOIN users u ON c.id = u.id WHERE u.is_active = FALSE").Scan(&dbInfo.PassiveCoaches)

	// Students counts
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students").Scan(&dbInfo.TotalStudents)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students s JOIN users u ON s.id = u.id WHERE u.is_active = TRUE").Scan(&dbInfo.ActiveStudents)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students s JOIN users u ON s.id = u.id WHERE u.is_active = FALSE").Scan(&dbInfo.PassiveStudents)

	// Coaches near expiration (<= 7 days remaining)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM coaches WHERE auth_end_date <= CURRENT_DATE + 7 AND auth_end_date >= CURRENT_DATE").Scan(&dbInfo.CoachesNearExpiration)

	// Coaches at or over capacity
	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM coaches c
		WHERE c.student_capacity <= (SELECT COUNT(*) FROM students s WHERE s.coach_id = c.id)
	`).Scan(&dbInfo.CoachesAtOrOverCapacity)

	// Applications pending
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM coach_applications WHERE status = 'pending'").Scan(&dbInfo.PendingApplicationsCount)

	// Total Payments (TRY TRY!)
	var tryPayments *float64
	_ = s.db.QueryRow(ctx, "SELECT SUM(amount) FROM coach_payments WHERE status = 'paid'").Scan(&tryPayments)
	if tryPayments != nil {
		dbInfo.TotalPaymentsReceived = *tryPayments
	}

	return &dbInfo, nil
}

type CoachDashboard struct {
	CapacityUsed       int     `json:"capacity_used"`
	CapacityLimit      int     `json:"capacity_limit"`
	AuthEndDate        string  `json:"auth_end_date"`
	DaysRemaining      int     `json:"days_remaining"`
	ActiveStudents     int     `json:"active_students"`
	PassiveStudents    int     `json:"passive_students"`
	TodaySolvedCount   int     `json:"today_solved_count"`
	WeeklySolvedCount  int     `json:"weekly_solved_count"`
	PendingHomeworks   int     `json:"pending_homeworks"`
	AwaitingHwApproval int     `json:"awaiting_hw_approval"`
	CompletedHomeworks int     `json:"completed_homeworks"`
	OverdueHomeworks   int     `json:"overdue_homeworks"`
	TotalPaymentsDue   float64 `json:"total_payments_due"`
}

func (s *DashboardService) GetCoachDashboard(ctx context.Context, coachID string) (*CoachDashboard, error) {
	var d CoachDashboard

	// Get capacity & expiry info
	var authEnd time.Time
	err := s.db.QueryRow(ctx, "SELECT student_capacity, auth_end_date FROM coaches WHERE id = $1", coachID).Scan(&d.CapacityLimit, &authEnd)
	if err != nil {
		return nil, err
	}
	d.AuthEndDate = authEnd.Format("2006-01-02")
	d.DaysRemaining = int(time.Until(authEnd).Hours() / 24)
	if d.DaysRemaining < 0 {
		d.DaysRemaining = 0
	}

	// Current student capacity used
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students WHERE coach_id = $1", coachID).Scan(&d.CapacityUsed)

	// Active/passive students
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students s JOIN users u ON s.id = u.id WHERE s.coach_id = $1 AND u.is_active = TRUE", coachID).Scan(&d.ActiveStudents)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM students s JOIN users u ON s.id = u.id WHERE s.coach_id = $1 AND u.is_active = FALSE", coachID).Scan(&d.PassiveStudents)

	// Solve stats (Today & Weekly)
	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(correct + incorrect + blank), 0)
		FROM question_solving_entries
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1)
		AND date = CURRENT_DATE
	`, coachID).Scan(&d.TodaySolvedCount)

	_ = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(correct + incorrect + blank), 0)
		FROM question_solving_entries
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1)
		AND date >= CURRENT_DATE - 7
	`, coachID).Scan(&d.WeeklySolvedCount)

	// Homework stats
	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homework
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1) AND status = 'waiting'
	`, coachID).Scan(&d.PendingHomeworks)

	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homework
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1) AND status = 'awaiting_approval'
	`, coachID).Scan(&d.AwaitingHwApproval)

	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homework
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1) AND status = 'completed'
	`, coachID).Scan(&d.CompletedHomeworks)

	_ = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM homework
		WHERE student_id IN (SELECT id FROM students WHERE coach_id = $1) AND status != 'completed' AND due_date < CURRENT_DATE
	`, coachID).Scan(&d.OverdueHomeworks)

	// Unpaid/Pending student payments sum
	var dueAmt *float64
	_ = s.db.QueryRow(ctx, `
		SELECT SUM(amount) FROM student_payments
		WHERE coach_id = $1 AND status IN ('pending', 'overdue')
	`, coachID).Scan(&dueAmt)
	if dueAmt != nil {
		d.TotalPaymentsDue = *dueAmt
	}

	return &d, nil
}

type StudentDashboard struct {
	TodaySolvedCount   int      `json:"today_solved_count"`
	WeeklySolvedCount  int      `json:"weekly_solved_count"`
	PendingHomeworks   int      `json:"pending_homeworks"`
	AwaitingHwApproval int      `json:"awaiting_hw_approval"`
	OverdueHomeworks   int      `json:"overdue_homeworks"`
	LatestExamNet      float64  `json:"latest_exam_net"`
	LatestExamName     string   `json:"latest_exam_name"`
	NextMeetingDate    *string  `json:"next_meeting_date"`
	UnreadNotifCount   int      `json:"unread_notif_count"`
	ResourcesProgress  float64  `json:"resources_progress"`
}

func (s *DashboardService) GetStudentDashboard(ctx context.Context, studentID string) (*StudentDashboard, error) {
	var d StudentDashboard

	// solved count
	_ = s.db.QueryRow(ctx, "SELECT COALESCE(SUM(correct + incorrect + blank), 0) FROM question_solving_entries WHERE student_id = $1 AND date = CURRENT_DATE", studentID).Scan(&d.TodaySolvedCount)
	_ = s.db.QueryRow(ctx, "SELECT COALESCE(SUM(correct + incorrect + blank), 0) FROM question_solving_entries WHERE student_id = $1 AND date >= CURRENT_DATE - 7", studentID).Scan(&d.WeeklySolvedCount)

	// homework count
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM homework WHERE student_id = $1 AND status = 'waiting'", studentID).Scan(&d.PendingHomeworks)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM homework WHERE student_id = $1 AND status = 'awaiting_approval'", studentID).Scan(&d.AwaitingHwApproval)
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM homework WHERE student_id = $1 AND status != 'completed' AND due_date < CURRENT_DATE", studentID).Scan(&d.OverdueHomeworks)

	// latest exam
	var name *string
	var net *float64
	_ = s.db.QueryRow(ctx, `
		SELECT exam_name, total_net FROM trial_exams
		WHERE student_id = $1
		ORDER BY exam_date DESC, created_at DESC
		LIMIT 1
	`, studentID).Scan(&name, &net)
	if name != nil && net != nil {
		d.LatestExamName = *name
		d.LatestExamNet = *net
	}

	// next meeting
	var mtDate *time.Time
	_ = s.db.QueryRow(ctx, `
		SELECT meeting_date FROM meetings
		WHERE student_id = $1 AND status = 'planned' AND meeting_date >= NOW()
		ORDER BY meeting_date ASC
		LIMIT 1
	`, studentID).Scan(&mtDate)
	if mtDate != nil {
		str := mtDate.Format(time.RFC3339)
		d.NextMeetingDate = &str
	}

	// unread notif count
	_ = s.db.QueryRow(ctx, "SELECT COUNT(*) FROM notification_recipients WHERE recipient_id = $1 AND is_read = FALSE", studentID).Scan(&d.UnreadNotifCount)

	// average resource progress
	var avgProgress *float64
	_ = s.db.QueryRow(ctx, "SELECT AVG(progress_percentage) FROM resources WHERE student_id = $1", studentID).Scan(&avgProgress)
	if avgProgress != nil {
		d.ResourcesProgress = *avgProgress
	}

	return &d, nil
}
