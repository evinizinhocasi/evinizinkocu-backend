package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"evinizinkocu-backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTrackersRepository struct {
	db *pgxpool.Pool
}

func NewPostgresTrackersRepository(db *pgxpool.Pool) *PostgresTrackersRepository {
	return &PostgresTrackersRepository{db: db}
}

// --- Question Solving Tracker ---

func (r *PostgresTrackersRepository) CreateQuestionSolving(ctx context.Context, e *domain.QuestionSolvingEntry) error {
	query := `
		INSERT INTO question_solving_entries (student_id, creator_id, date, subject_id, topic_id, correct, incorrect, blank, net, note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, e.StudentID, e.CreatorID, e.Date, e.SubjectID, e.TopicID, e.Correct, e.Incorrect, e.Blank, e.Net, e.Note).
		Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetQuestionSolvingByID(ctx context.Context, id string) (*domain.QuestionSolvingEntry, error) {
	query := `
		SELECT id, student_id, creator_id, date, subject_id, topic_id, correct, incorrect, blank, net, note, created_at, updated_at
		FROM question_solving_entries WHERE id = $1
	`
	var e domain.QuestionSolvingEntry
	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.StudentID, &e.CreatorID, &e.Date, &e.SubjectID, &e.TopicID, &e.Correct, &e.Incorrect, &e.Blank, &e.Net, &e.Note, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("entry not found")
		}
		return nil, err
	}
	return &e, nil
}

func (r *PostgresTrackersRepository) UpdateQuestionSolving(ctx context.Context, e *domain.QuestionSolvingEntry) error {
	query := `
		UPDATE question_solving_entries
		SET date = $1, subject_id = $2, topic_id = $3, correct = $4, incorrect = $5, blank = $6, net = $7, note = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query, e.Date, e.SubjectID, e.TopicID, e.Correct, e.Incorrect, e.Blank, e.Net, e.Note, e.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteQuestionSolving(ctx context.Context, id string) error {
	query := `DELETE FROM question_solving_entries WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListQuestionSolvingByStudent(ctx context.Context, studentID string, startDate, endDate *time.Time) ([]*domain.QuestionSolvingEntry, error) {
	filters := "student_id = $1"
	args := []interface{}{studentID}
	placeholderIdx := 2

	if startDate != nil {
		filters += fmt.Sprintf(" AND date >= $%d", placeholderIdx)
		args = append(args, *startDate)
		placeholderIdx++
	}
	if endDate != nil {
		filters += fmt.Sprintf(" AND date <= $%d", placeholderIdx)
		args = append(args, *endDate)
		placeholderIdx++
	}

	query := fmt.Sprintf(`
		SELECT q.id, q.student_id, q.creator_id, q.date, q.subject_id, q.topic_id, q.correct, q.incorrect, q.blank, q.net, q.note, q.created_at, q.updated_at,
		       s.name, t.name
		FROM question_solving_entries q
		JOIN subjects s ON q.subject_id = s.id
		LEFT JOIN topics t ON q.topic_id = t.id
		WHERE %s
		ORDER BY q.date DESC, q.created_at DESC
	`, filters)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.QuestionSolvingEntry
	for rows.Next() {
		var e domain.QuestionSolvingEntry
		var subName string
		var topName *string
		err := rows.Scan(
			&e.ID, &e.StudentID, &e.CreatorID, &e.Date, &e.SubjectID, &e.TopicID, &e.Correct, &e.Incorrect, &e.Blank, &e.Net, &e.Note, &e.CreatedAt, &e.UpdatedAt,
			&subName, &topName,
		)
		if err != nil {
			return nil, err
		}
		e.Subject = &domain.Subject{ID: e.SubjectID, Name: subName}
		if topName != nil && e.TopicID != nil {
			e.Topic = &domain.Topic{ID: *e.TopicID, Name: *topName}
		}
		list = append(list, &e)
	}
	return list, nil
}

// --- Homework Tracker ---

func (r *PostgresTrackersRepository) CreateHomework(ctx context.Context, hw *domain.Homework) error {
	query := `
		INSERT INTO homework (student_id, creator_coach_id, title, description, subject_id, topic_id, source, page_range, url, start_date, due_date, status, coach_explanation, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, hw.StudentID, hw.CreatorCoachID, hw.Title, hw.Description, hw.SubjectID, hw.TopicID, hw.Source, hw.PageRange, hw.URL, hw.StartDate, hw.DueDate, hw.Status, hw.CoachExplanation).
		Scan(&hw.ID, &hw.CreatedAt, &hw.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetHomeworkByID(ctx context.Context, id string) (*domain.Homework, error) {
	query := `
		SELECT id, student_id, creator_coach_id, title, description, subject_id, topic_id, source, page_range, url, start_date, due_date, status, coach_explanation, created_at, updated_at
		FROM homework WHERE id = $1
	`
	var hw domain.Homework
	err := r.db.QueryRow(ctx, query, id).Scan(
		&hw.ID, &hw.StudentID, &hw.CreatorCoachID, &hw.Title, &hw.Description, &hw.SubjectID, &hw.TopicID, &hw.Source, &hw.PageRange, &hw.URL, &hw.StartDate, &hw.DueDate, &hw.Status, &hw.CoachExplanation, &hw.CreatedAt, &hw.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("homework not found")
		}
		return nil, err
	}
	return &hw, nil
}

func (r *PostgresTrackersRepository) UpdateHomework(ctx context.Context, hw *domain.Homework) error {
	query := `
		UPDATE homework
		SET title = $1, description = $2, subject_id = $3, topic_id = $4, source = $5, page_range = $6, url = $7, start_date = $8, due_date = $9, status = $10, coach_explanation = $11, updated_at = NOW()
		WHERE id = $12
	`
	_, err := r.db.Exec(ctx, query, hw.Title, hw.Description, hw.SubjectID, hw.TopicID, hw.Source, hw.PageRange, hw.URL, hw.StartDate, hw.DueDate, hw.Status, hw.CoachExplanation, hw.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteHomework(ctx context.Context, id string) error {
	query := `DELETE FROM homework WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListHomeworkByStudent(ctx context.Context, studentID string) ([]*domain.Homework, error) {
	query := `
		SELECT h.id, h.student_id, h.creator_coach_id, h.title, h.description, h.subject_id, h.topic_id, h.source, h.page_range, h.url, h.start_date, h.due_date, h.status, h.coach_explanation, h.created_at, h.updated_at,
		       s.name, t.name
		FROM homework h
		JOIN subjects s ON h.subject_id = s.id
		LEFT JOIN topics t ON h.topic_id = t.id
		WHERE h.student_id = $1
		ORDER BY h.due_date ASC, h.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Homework
	for rows.Next() {
		var hw domain.Homework
		var subName string
		var topName *string
		err := rows.Scan(
			&hw.ID, &hw.StudentID, &hw.CreatorCoachID, &hw.Title, &hw.Description, &hw.SubjectID, &hw.TopicID, &hw.Source, &hw.PageRange, &hw.URL, &hw.StartDate, &hw.DueDate, &hw.Status, &hw.CoachExplanation, &hw.CreatedAt, &hw.UpdatedAt,
			&subName, &topName,
		)
		if err != nil {
			return nil, err
		}
		hw.Subject = &domain.Subject{ID: hw.SubjectID, Name: subName}
		if topName != nil && hw.TopicID != nil {
			hw.Topic = &domain.Topic{ID: *hw.TopicID, Name: *topName}
		}
		list = append(list, &hw)
	}
	return list, nil
}

// --- Student Resources ---

func (r *PostgresTrackersRepository) CreateResource(ctx context.Context, res *domain.Resource) error {
	query := `
		INSERT INTO resources (student_id, name, publisher, subject_id, description, total_pages, completed_pages, progress_percentage, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, res.StudentID, res.Name, res.Publisher, res.SubjectID, res.Description, res.TotalPages, res.CompletedPages, res.ProgressPercentage, res.Status).
		Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetResourceByID(ctx context.Context, id string) (*domain.Resource, error) {
	query := `
		SELECT id, student_id, name, publisher, subject_id, description, total_pages, completed_pages, progress_percentage, status, created_at, updated_at
		FROM resources WHERE id = $1
	`
	var res domain.Resource
	err := r.db.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.StudentID, &res.Name, &res.Publisher, &res.SubjectID, &res.Description, &res.TotalPages, &res.CompletedPages, &res.ProgressPercentage, &res.Status, &res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("resource not found")
		}
		return nil, err
	}
	return &res, nil
}

func (r *PostgresTrackersRepository) UpdateResource(ctx context.Context, res *domain.Resource) error {
	query := `
		UPDATE resources
		SET name = $1, publisher = $2, subject_id = $3, description = $4, total_pages = $5, completed_pages = $6, progress_percentage = $7, status = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query, res.Name, res.Publisher, res.SubjectID, res.Description, res.TotalPages, res.CompletedPages, res.ProgressPercentage, res.Status, res.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteResource(ctx context.Context, id string) error {
	query := `DELETE FROM resources WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListResourcesByStudent(ctx context.Context, studentID string) ([]*domain.Resource, error) {
	query := `
		SELECT r.id, r.student_id, r.name, r.publisher, r.subject_id, r.description, r.total_pages, r.completed_pages, r.progress_percentage, r.status, r.created_at, r.updated_at,
		       s.name
		FROM resources r
		JOIN subjects s ON r.subject_id = s.id
		WHERE r.student_id = $1
		ORDER BY r.status DESC, r.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Resource
	for rows.Next() {
		var res domain.Resource
		var subName string
		err := rows.Scan(
			&res.ID, &res.StudentID, &res.Name, &res.Publisher, &res.SubjectID, &res.Description, &res.TotalPages, &res.CompletedPages, &res.ProgressPercentage, &res.Status, &res.CreatedAt, &res.UpdatedAt,
			&subName,
		)
		if err != nil {
			return nil, err
		}
		res.Subject = &domain.Subject{ID: res.SubjectID, Name: subName}
		list = append(list, &res)
	}
	return list, nil
}

// --- School Timetable ---

func (r *PostgresTrackersRepository) SaveTimetable(ctx context.Context, entries []*domain.SchoolTimetableEntry, studentID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Clear previous timetable entries
	_, err = tx.Exec(ctx, "DELETE FROM school_timetable_entries WHERE student_id = $1", studentID)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		query := `
			INSERT INTO school_timetable_entries (student_id, weekday, start_time, end_time, subject_name, created_at, updated_at)
			VALUES ($1, $2, $3::time, $4::time, $5, NOW(), NOW())
		`
		_, err = tx.Exec(ctx, query, studentID, entry.Weekday, entry.StartTime, entry.EndTime, entry.SubjectName)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresTrackersRepository) GetTimetableByStudent(ctx context.Context, studentID string) ([]*domain.SchoolTimetableEntry, error) {
	query := `
		SELECT id, student_id, weekday, start_time::text, end_time::text, subject_name, created_at, updated_at
		FROM school_timetable_entries
		WHERE student_id = $1
		ORDER BY weekday, start_time
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.SchoolTimetableEntry
	for rows.Next() {
		var entry domain.SchoolTimetableEntry
		err := rows.Scan(&entry.ID, &entry.StudentID, &entry.Weekday, &entry.StartTime, &entry.EndTime, &entry.SubjectName, &entry.CreatedAt, &entry.UpdatedAt)
		if err != nil {
			return nil, err
		}
		list = append(list, &entry)
	}
	return list, nil
}

// --- Weekly Coaching Plan ---

func (r *PostgresTrackersRepository) CreateWeeklyPlanItem(ctx context.Context, item *domain.WeeklyPlanItem) error {
	query := `
		INSERT INTO weekly_plan_items (student_id, coach_id, date, start_time, end_time, title, subject_id, topic_id, note, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4::time, $5::time, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, item.StudentID, item.CoachID, item.Date, item.StartTime, item.EndTime, item.Title, item.SubjectID, item.TopicID, item.Note, item.URL).
		Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetWeeklyPlanItemByID(ctx context.Context, id string) (*domain.WeeklyPlanItem, error) {
	query := `
		SELECT id, student_id, coach_id, date, start_time::text, end_time::text, title, subject_id, topic_id, note, url, created_at, updated_at
		FROM weekly_plan_items WHERE id = $1
	`
	var item domain.WeeklyPlanItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.StudentID, &item.CoachID, &item.Date, &item.StartTime, &item.EndTime, &item.Title, &item.SubjectID, &item.TopicID, &item.Note, &item.URL, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("plan item not found")
		}
		return nil, err
	}
	return &item, nil
}

func (r *PostgresTrackersRepository) UpdateWeeklyPlanItem(ctx context.Context, item *domain.WeeklyPlanItem) error {
	query := `
		UPDATE weekly_plan_items
		SET date = $1, start_time = $2::time, end_time = $3::time, title = $4, subject_id = $5, topic_id = $6, note = $7, url = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query, item.Date, item.StartTime, item.EndTime, item.Title, item.SubjectID, item.TopicID, item.Note, item.URL, item.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteWeeklyPlanItem(ctx context.Context, id string) error {
	query := `DELETE FROM weekly_plan_items WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListWeeklyPlanItems(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*domain.WeeklyPlanItem, error) {
	query := `
		SELECT p.id, p.student_id, p.coach_id, p.date, p.start_time::text, p.end_time::text, p.title, p.subject_id, p.topic_id, p.note, p.url, p.created_at, p.updated_at,
		       s.name, t.name
		FROM weekly_plan_items p
		LEFT JOIN subjects s ON p.subject_id = s.id
		LEFT JOIN topics t ON p.topic_id = t.id
		WHERE p.student_id = $1 AND p.date >= $2 AND p.date <= $3
		ORDER BY p.date, p.start_time
	`
	rows, err := r.db.Query(ctx, query, studentID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.WeeklyPlanItem
	for rows.Next() {
		var item domain.WeeklyPlanItem
		var subName *string
		var topName *string
		err := rows.Scan(
			&item.ID, &item.StudentID, &item.CoachID, &item.Date, &item.StartTime, &item.EndTime, &item.Title, &item.SubjectID, &item.TopicID, &item.Note, &item.URL, &item.CreatedAt, &item.UpdatedAt,
			&subName, &topName,
		)
		if err != nil {
			return nil, err
		}
		if subName != nil && item.SubjectID != nil {
			item.Subject = &domain.Subject{ID: *item.SubjectID, Name: *subName}
		}
		if topName != nil && item.TopicID != nil {
			item.Topic = &domain.Topic{ID: *item.TopicID, Name: *topName}
		}
		list = append(list, &item)
	}
	return list, nil
}

func (r *PostgresTrackersRepository) CopyWeek(ctx context.Context, studentID string, fromStart, toStart time.Time) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	fromEnd := fromStart.AddDate(0, 0, 7)
	toEnd := toStart.AddDate(0, 0, 7)

	// Delete target week items first to avoid overlays
	_, err = tx.Exec(ctx, "DELETE FROM weekly_plan_items WHERE student_id = $1 AND date >= $2 AND date < $3", studentID, toStart, toEnd)
	if err != nil {
		return err
	}

	// Fetch source items
	query := `
		SELECT student_id, coach_id, date, start_time::text, end_time::text, title, subject_id, topic_id, note, url
		FROM weekly_plan_items
		WHERE student_id = $1 AND date >= $2 AND date < $3
	`
	rows, err := tx.Query(ctx, query, studentID, fromStart, fromEnd)
	if err != nil {
		return err
	}

	type rawItem struct {
		studentID string
		coachID   string
		date      time.Time
		start     string
		end       string
		title     string
		subID     *string
		topID     *string
		note      string
		url       string
	}

	var items []rawItem
	for rows.Next() {
		var i rawItem
		err := rows.Scan(&i.studentID, &i.coachID, &i.date, &i.start, &i.end, &i.title, &i.subID, &i.topID, &i.note, &i.url)
		if err != nil {
			rows.Close()
			return err
		}
		items = append(items, i)
	}
	rows.Close()

	// Offset days
	dayDiff := int(toStart.Sub(fromStart).Hours() / 24)

	// Insert duplicate plan items with offset
	for _, item := range items {
		newDate := item.date.AddDate(0, 0, dayDiff)
		insertQuery := `
			INSERT INTO weekly_plan_items (student_id, coach_id, date, start_time, end_time, title, subject_id, topic_id, note, url, created_at, updated_at)
			VALUES ($1, $2, $3, $4::time, $5::time, $6, $7, $8, $9, $10, NOW(), NOW())
		`
		_, err = tx.Exec(ctx, insertQuery, item.studentID, item.coachID, newDate, item.start, item.end, item.title, item.subID, item.topID, item.note, item.url)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// --- Missing Topics ---

func (r *PostgresTrackersRepository) CreateMissingTopic(ctx context.Context, mt *domain.MissingTopic) error {
	query := `
		INSERT INTO missing_topics (student_id, subject_id, topic_id, description, priority, status, target_date, solution_text, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, mt.StudentID, mt.SubjectID, mt.TopicID, mt.Description, mt.Priority, mt.Status, mt.TargetDate, mt.SolutionText, mt.URL).
		Scan(&mt.ID, &mt.CreatedAt, &mt.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetMissingTopicByID(ctx context.Context, id string) (*domain.MissingTopic, error) {
	query := `
		SELECT id, student_id, subject_id, topic_id, description, priority, status, target_date, solution_text, url, created_at, updated_at
		FROM missing_topics WHERE id = $1
	`
	var mt domain.MissingTopic
	err := r.db.QueryRow(ctx, query, id).Scan(
		&mt.ID, &mt.StudentID, &mt.SubjectID, &mt.TopicID, &mt.Description, &mt.Priority, &mt.Status, &mt.TargetDate, &mt.SolutionText, &mt.URL, &mt.CreatedAt, &mt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("missing topic entry not found")
		}
		return nil, err
	}
	return &mt, nil
}

func (r *PostgresTrackersRepository) UpdateMissingTopic(ctx context.Context, mt *domain.MissingTopic) error {
	query := `
		UPDATE missing_topics
		SET subject_id = $1, topic_id = $2, description = $3, priority = $4, status = $5, target_date = $6, solution_text = $7, url = $8, updated_at = NOW()
		WHERE id = $9
	`
	_, err := r.db.Exec(ctx, query, mt.SubjectID, mt.TopicID, mt.Description, mt.Priority, mt.Status, mt.TargetDate, mt.SolutionText, mt.URL, mt.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteMissingTopic(ctx context.Context, id string) error {
	query := `DELETE FROM missing_topics WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListMissingTopicsByStudent(ctx context.Context, studentID string) ([]*domain.MissingTopic, error) {
	query := `
		SELECT m.id, m.student_id, m.subject_id, m.topic_id, m.description, m.priority, m.status, m.target_date, m.solution_text, m.url, m.created_at, m.updated_at,
		       s.name, t.name
		FROM missing_topics m
		JOIN subjects s ON m.subject_id = s.id
		JOIN topics t ON m.topic_id = t.id
		WHERE m.student_id = $1
		ORDER BY m.status ASC, m.priority DESC, m.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.MissingTopic
	for rows.Next() {
		var mt domain.MissingTopic
		var subName string
		var topName string
		err := rows.Scan(
			&mt.ID, &mt.StudentID, &mt.SubjectID, &mt.TopicID, &mt.Description, &mt.Priority, &mt.Status, &mt.TargetDate, &mt.SolutionText, &mt.URL, &mt.CreatedAt, &mt.UpdatedAt,
			&subName, &topName,
		)
		if err != nil {
			return nil, err
		}
		mt.Subject = &domain.Subject{ID: mt.SubjectID, Name: subName}
		mt.Topic = &domain.Topic{ID: mt.TopicID, Name: topName}
		list = append(list, &mt)
	}
	return list, nil
}

// --- Goals ---

func (r *PostgresTrackersRepository) CreateGoal(ctx context.Context, g *domain.Goal) error {
	query := `
		INSERT INTO goals (student_id, type, title, description, target_value, current_value, unit, start_date, target_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, g.StudentID, g.Type, g.Title, g.Description, g.TargetValue, g.CurrentValue, g.Unit, g.StartDate, g.TargetDate, g.Status).
		Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetGoalByID(ctx context.Context, id string) (*domain.Goal, error) {
	query := `
		SELECT id, student_id, type, title, description, target_value, current_value, unit, start_date, target_date, status, created_at, updated_at
		FROM goals WHERE id = $1
	`
	var g domain.Goal
	err := r.db.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.StudentID, &g.Type, &g.Title, &g.Description, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.TargetDate, &g.Status, &g.CreatedAt, &g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("goal not found")
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresTrackersRepository) UpdateGoal(ctx context.Context, g *domain.Goal) error {
	query := `
		UPDATE goals
		SET type = $1, title = $2, description = $3, target_value = $4, current_value = $5, unit = $6, start_date = $7, target_date = $8, status = $9, updated_at = NOW()
		WHERE id = $10
	`
	_, err := r.db.Exec(ctx, query, g.Type, g.Title, g.Description, g.TargetValue, g.CurrentValue, g.Unit, g.StartDate, g.TargetDate, g.Status, g.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteGoal(ctx context.Context, id string) error {
	query := `DELETE FROM goals WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListGoalsByStudent(ctx context.Context, studentID string) ([]*domain.Goal, error) {
	query := `
		SELECT id, student_id, type, title, description, target_value, current_value, unit, start_date, target_date, status, created_at, updated_at
		FROM goals
		WHERE student_id = $1
		ORDER BY target_date ASC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Goal
	for rows.Next() {
		var g domain.Goal
		err := rows.Scan(
			&g.ID, &g.StudentID, &g.Type, &g.Title, &g.Description, &g.TargetValue, &g.CurrentValue, &g.Unit, &g.StartDate, &g.TargetDate, &g.Status, &g.CreatedAt, &g.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &g)
	}
	return list, nil
}

// --- Coaching Meetings ---

func (r *PostgresTrackersRepository) CreateMeeting(ctx context.Context, m *domain.Meeting) error {
	query := `
		INSERT INTO meetings (student_id, coach_id, meeting_date, duration_minutes, notes, meeting_url, next_meeting_date, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, m.StudentID, m.CoachID, m.MeetingDate, m.DurationMinutes, m.Notes, m.MeetingURL, m.NextMeetingDate, m.Status).
		Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

func (r *PostgresTrackersRepository) GetMeetingByID(ctx context.Context, id string) (*domain.Meeting, error) {
	query := `
		SELECT id, student_id, coach_id, meeting_date, duration_minutes, notes, meeting_url, next_meeting_date, status, created_at, updated_at
		FROM meetings WHERE id = $1
	`
	var m domain.Meeting
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.StudentID, &m.CoachID, &m.MeetingDate, &m.DurationMinutes, &m.Notes, &m.MeetingURL, &m.NextMeetingDate, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("meeting not found")
		}
		return nil, err
	}
	return &m, nil
}

func (r *PostgresTrackersRepository) UpdateMeeting(ctx context.Context, m *domain.Meeting) error {
	query := `
		UPDATE meetings
		SET meeting_date = $1, duration_minutes = $2, notes = $3, meeting_url = $4, next_meeting_date = $5, status = $6, updated_at = NOW()
		WHERE id = $7
	`
	_, err := r.db.Exec(ctx, query, m.MeetingDate, m.DurationMinutes, m.Notes, m.MeetingURL, m.NextMeetingDate, m.Status, m.ID)
	return err
}

func (r *PostgresTrackersRepository) DeleteMeeting(ctx context.Context, id string) error {
	query := `DELETE FROM meetings WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresTrackersRepository) ListMeetingsByStudent(ctx context.Context, studentID string) ([]*domain.Meeting, error) {
	query := `
		SELECT id, student_id, coach_id, meeting_date, duration_minutes, notes, meeting_url, next_meeting_date, status, created_at, updated_at
		FROM meetings
		WHERE student_id = $1
		ORDER BY meeting_date DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		err := rows.Scan(
			&m.ID, &m.StudentID, &m.CoachID, &m.MeetingDate, &m.DurationMinutes, &m.Notes, &m.MeetingURL, &m.NextMeetingDate, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}

func (r *PostgresTrackersRepository) ListMeetingsByCoach(ctx context.Context, coachID string) ([]*domain.Meeting, error) {
	query := `
		SELECT id, student_id, coach_id, meeting_date, duration_minutes, notes, meeting_url, next_meeting_date, status, created_at, updated_at
		FROM meetings
		WHERE coach_id = $1
		ORDER BY meeting_date DESC, created_at DESC
	`
	rows, err := r.db.Query(ctx, query, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Meeting
	for rows.Next() {
		var m domain.Meeting
		err := rows.Scan(
			&m.ID, &m.StudentID, &m.CoachID, &m.MeetingDate, &m.DurationMinutes, &m.Notes, &m.MeetingURL, &m.NextMeetingDate, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, nil
}
