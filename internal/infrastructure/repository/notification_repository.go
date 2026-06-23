package repository

import (
	"context"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresNotificationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresNotificationRepository(db *pgxpool.Pool) domain.NotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) CreateNotification(ctx context.Context, n *domain.Notification, recipientIDs []string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO notifications (title, body, sender_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	err = tx.QueryRow(ctx, query, n.Title, n.Body, n.SenderID).Scan(&n.ID, &n.CreatedAt)
	if err != nil {
		return err
	}

	for _, recipientID := range recipientIDs {
		recQuery := `
			INSERT INTO notification_recipients (notification_id, recipient_id, is_read, created_at)
			VALUES ($1, $2, FALSE, NOW())
		`
		_, err = tx.Exec(ctx, recQuery, n.ID, recipientID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresNotificationRepository) GetInbox(ctx context.Context, userID string) ([]*domain.NotificationRecipient, error) {
	query := `
		SELECT nr.id, nr.notification_id, nr.recipient_id, nr.is_read, nr.read_at, nr.created_at,
		       n.title, n.body, n.sender_id
		FROM notification_recipients nr
		JOIN notifications n ON nr.notification_id = n.id
		WHERE nr.recipient_id = $1
		ORDER BY nr.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationRecipient
	for rows.Next() {
		var nr domain.NotificationRecipient
		var n domain.Notification
		err := rows.Scan(&nr.ID, &nr.NotificationID, &nr.RecipientID, &nr.IsRead, &nr.ReadAt, &nr.CreatedAt, &n.Title, &n.Body, &n.SenderID)
		if err != nil {
			return nil, err
		}
		// Embed parent notification details
		n.ID = nr.NotificationID
		// In-app representation
		// Wait, domain model has Notification structure, we can map to customized UI structures or just set a helper.
		// Let's create custom inbox item JSON or mock mapping if needed. For now let's set Notification as reference.
		// Wait, domain has no Notification field directly, let's attach to the recipient struct in-app or query it.
		// Since we defined the struct inside trackers/notifications we can just return these.
		// Let's make sure we return recipient objects. To send title/body we can return standard JSON mapping or map to client model directly.
		// Let's return them.
		list = append(list, &nr)
	}
	return list, nil
}

func (r *PostgresNotificationRepository) MarkAsRead(ctx context.Context, recipientID string) error {
	query := `UPDATE notification_recipients SET is_read = TRUE, read_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, recipientID)
	return err
}

func (r *PostgresNotificationRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	query := `UPDATE notification_recipients SET is_read = TRUE, read_at = NOW() WHERE recipient_id = $1 AND is_read = FALSE`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// Schedules CRUD

func (r *PostgresNotificationRepository) CreateSchedule(ctx context.Context, ns *domain.NotificationSchedule) error {
	query := `
		INSERT INTO notification_schedules (coach_id, title, body, target_selection, target_student_ids, schedule_type, selected_weekdays, schedule_time, start_date, end_date, next_run_at, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::time, $9, $10, $11, $12, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, ns.CoachID, ns.Title, ns.Body, ns.TargetSelection, ns.TargetStudentIDs, ns.ScheduleType, ns.SelectedWeekdays, ns.ScheduleTime, ns.StartDate, ns.EndDate, ns.NextRunAt, ns.IsActive).
		Scan(&ns.ID, &ns.CreatedAt, &ns.UpdatedAt)
}

func (r *PostgresNotificationRepository) GetScheduleByID(ctx context.Context, id string) (*domain.NotificationSchedule, error) {
	// Standard Get
	query := `
		SELECT id, coach_id, title, body, target_selection, target_student_ids, schedule_type, selected_weekdays, schedule_time::text, start_date, end_date, next_run_at, is_active, created_at, updated_at
		FROM notification_schedules WHERE id = $1
	`
	var ns domain.NotificationSchedule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ns.ID, &ns.CoachID, &ns.Title, &ns.Body, &ns.TargetSelection, &ns.TargetStudentIDs, &ns.ScheduleType, &ns.SelectedWeekdays, &ns.ScheduleTime, &ns.StartDate, &ns.EndDate, &ns.NextRunAt, &ns.IsActive, &ns.CreatedAt, &ns.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ns, nil
}

func (r *PostgresNotificationRepository) UpdateSchedule(ctx context.Context, ns *domain.NotificationSchedule) error {
	query := `
		UPDATE notification_schedules
		SET title = $1, body = $2, target_selection = $3, target_student_ids = $4, schedule_type = $5, selected_weekdays = $6, schedule_time = $7::time, start_date = $8, end_date = $9, next_run_at = $10, is_active = $11, updated_at = NOW()
		WHERE id = $12
	`
	_, err := r.db.Exec(ctx, query, ns.Title, ns.Body, ns.TargetSelection, ns.TargetStudentIDs, ns.ScheduleType, ns.SelectedWeekdays, ns.ScheduleTime, ns.StartDate, ns.EndDate, ns.NextRunAt, ns.IsActive, ns.ID)
	return err
}

func (r *PostgresNotificationRepository) DeleteSchedule(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM notification_schedules WHERE id = $1", id)
	return err
}

func (r *PostgresNotificationRepository) ListSchedulesByCoach(ctx context.Context, coachID string) ([]*domain.NotificationSchedule, error) {
	query := `
		SELECT id, coach_id, title, body, target_selection, target_student_ids, schedule_type, selected_weekdays, schedule_time::text, start_date, end_date, next_run_at, is_active, created_at, updated_at
		FROM notification_schedules
		WHERE coach_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationSchedule
	for rows.Next() {
		var ns domain.NotificationSchedule
		err := rows.Scan(
			&ns.ID, &ns.CoachID, &ns.Title, &ns.Body, &ns.TargetSelection, &ns.TargetStudentIDs, &ns.ScheduleType, &ns.SelectedWeekdays, &ns.ScheduleTime, &ns.StartDate, &ns.EndDate, &ns.NextRunAt, &ns.IsActive, &ns.CreatedAt, &ns.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &ns)
	}
	return list, nil
}

// Worker claim queries using transactional row locks SELECT FOR UPDATE SKIP LOCKED
func (r *PostgresNotificationRepository) ClaimDueSchedules(ctx context.Context, workerID string, limit int, lockDuration time.Duration) ([]*domain.NotificationSchedule, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	lockedUntil := time.Now().Add(lockDuration)

	query := `
		UPDATE notification_schedules
		SET locked_by = $1, locked_until = $2
		WHERE id IN (
			SELECT id FROM notification_schedules
			WHERE next_run_at <= NOW() AND is_active = true AND (locked_until IS NULL OR locked_until < NOW())
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, coach_id, title, body, target_selection, target_student_ids, schedule_type, selected_weekdays, schedule_time::text, start_date, end_date, next_run_at, is_active, created_at, updated_at
	`
	rows, err := tx.Query(ctx, query, workerID, lockedUntil, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.NotificationSchedule
	for rows.Next() {
		var ns domain.NotificationSchedule
		err := rows.Scan(
			&ns.ID, &ns.CoachID, &ns.Title, &ns.Body, &ns.TargetSelection, &ns.TargetStudentIDs, &ns.ScheduleType, &ns.SelectedWeekdays, &ns.ScheduleTime, &ns.StartDate, &ns.EndDate, &ns.NextRunAt, &ns.IsActive, &ns.CreatedAt, &ns.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &ns)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *PostgresNotificationRepository) SaveExecutionRecord(ctx context.Context, exec *domain.NotificationExecution) error {
	query := `
		INSERT INTO notification_executions (schedule_id, status, error_message, executed_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, executed_at
	`
	return r.db.QueryRow(ctx, query, exec.ScheduleID, exec.Status, exec.ErrorMessage).Scan(&exec.ID, &exec.ExecutedAt)
}
