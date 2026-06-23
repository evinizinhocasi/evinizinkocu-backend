package domain

import (
	"context"
	"time"
)

// --- Question Solving Tracker ---
type QuestionSolvingEntry struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	CreatorID string    `json:"creator_id"`
	Date      time.Time `json:"date"`
	SubjectID string    `json:"subject_id"`
	Subject   *Subject  `json:"subject,omitempty"`
	TopicID   *string   `json:"topic_id,omitempty"`
	Topic     *Topic    `json:"topic,omitempty"`
	Correct   int       `json:"correct"`
	Incorrect int       `json:"incorrect"`
	Blank     int       `json:"blank"`
	Net       float64   `json:"net"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type QuestionSolvingRepository interface {
	Create(ctx context.Context, entry *QuestionSolvingEntry) error
	GetByID(ctx context.Context, id string) (*QuestionSolvingEntry, error)
	Update(ctx context.Context, entry *QuestionSolvingEntry) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string, startDate, endDate *time.Time) ([]*QuestionSolvingEntry, error)
}

// --- Homework Tracker ---
type Homework struct {
	ID               string    `json:"id"`
	StudentID        string    `json:"student_id"`
	CreatorCoachID   string    `json:"creator_coach_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description,omitempty"`
	SubjectID        string    `json:"subject_id"`
	Subject          *Subject  `json:"subject,omitempty"`
	TopicID          *string   `json:"topic_id,omitempty"`
	Topic            *Topic    `json:"topic,omitempty"`
	Source           string    `json:"source,omitempty"`
	PageRange        string    `json:"page_range,omitempty"`
	URL              string    `json:"url,omitempty"`
	StartDate        time.Time `json:"start_date"`
	DueDate          time.Time `json:"due_date"`
	Status           string    `json:"status"` // 'waiting', 'started', 'awaiting_approval', 'completed'
	CoachExplanation string    `json:"coach_explanation,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type HomeworkRepository interface {
	Create(ctx context.Context, hw *Homework) error
	GetByID(ctx context.Context, id string) (*Homework, error)
	Update(ctx context.Context, hw *Homework) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string) ([]*Homework, error)
}

// --- Student Resources ---
type Resource struct {
	ID                 string    `json:"id"`
	StudentID          string    `json:"student_id"`
	Name               string    `json:"name"`
	Publisher          string    `json:"publisher"`
	SubjectID          string    `json:"subject_id"`
	Subject            *Subject  `json:"subject,omitempty"`
	Description        string    `json:"description,omitempty"`
	TotalPages         int       `json:"total_pages"`
	CompletedPages     int       `json:"completed_pages"`
	ProgressPercentage int       `json:"progress_percentage"`
	Status             string    `json:"status"` // 'planned', 'active', 'completed', 'paused'
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ResourceRepository interface {
	Create(ctx context.Context, res *Resource) error
	GetByID(ctx context.Context, id string) (*Resource, error)
	Update(ctx context.Context, res *Resource) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string) ([]*Resource, error)
}

// --- School Timetable ---
type SchoolTimetableEntry struct {
	ID          string    `json:"id"`
	StudentID   string    `json:"student_id"`
	Weekday     int       `json:"weekday"` // 1 = Monday, 7 = Sunday
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	SubjectName string    `json:"subject_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TimetableRepository interface {
	Save(ctx context.Context, entries []*SchoolTimetableEntry, studentID string) error
	GetByStudent(ctx context.Context, studentID string) ([]*SchoolTimetableEntry, error)
}

// --- Weekly Coaching Plan ---
type WeeklyPlanItem struct {
	ID        string    `json:"id"`
	StudentID string    `json:"student_id"`
	CoachID   string    `json:"coach_id"`
	Date      time.Time `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Title     string    `json:"title"`
	SubjectID *string   `json:"subject_id,omitempty"`
	Subject   *Subject  `json:"subject,omitempty"`
	TopicID   *string   `json:"topic_id,omitempty"`
	Topic     *Topic    `json:"topic,omitempty"`
	Note      string    `json:"note,omitempty"`
	URL       string    `json:"url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WeeklyPlanRepository interface {
	Create(ctx context.Context, item *WeeklyPlanItem) error
	GetByID(ctx context.Context, id string) (*WeeklyPlanItem, error)
	Update(ctx context.Context, item *WeeklyPlanItem) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string, startDate, endDate time.Time) ([]*WeeklyPlanItem, error)
	CopyWeek(ctx context.Context, studentID string, fromStart, toStart time.Time) error
}

// --- Missing Topics ---
type MissingTopic struct {
	ID           string     `json:"id"`
	StudentID    string     `json:"student_id"`
	SubjectID    string     `json:"subject_id"`
	Subject      *Subject   `json:"subject,omitempty"`
	TopicID      string     `json:"topic_id"`
	Topic        *Topic     `json:"topic,omitempty"`
	Description  string     `json:"description,omitempty"`
	Priority     string     `json:"priority"` // 'low', 'medium', 'high'
	Status       string     `json:"status"`   // 'identified', 'in_progress', 'resolved'
	TargetDate   *time.Time `json:"target_date,omitempty"`
	SolutionText string     `json:"solution_text,omitempty"`
	URL          string     `json:"url,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type MissingTopicsRepository interface {
	Create(ctx context.Context, mt *MissingTopic) error
	GetByID(ctx context.Context, id string) (*MissingTopic, error)
	Update(ctx context.Context, mt *MissingTopic) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string) ([]*MissingTopic, error)
}

// --- Goals ---
type Goal struct {
	ID           string    `json:"id"`
	StudentID    string    `json:"student_id"`
	Type         string    `json:"type"` // 'question_count', 'exam_net', 'subject_net', 'resource_completion', 'custom'
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	TargetValue  float64   `json:"target_value"`
	CurrentValue float64   `json:"current_value"`
	Unit         string    `json:"unit"`
	StartDate    time.Time `json:"start_date"`
	TargetDate   time.Time `json:"target_date"`
	Status       string    `json:"status"` // 'active', 'achieved', 'failed'
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type GoalsRepository interface {
	Create(ctx context.Context, g *Goal) error
	GetByID(ctx context.Context, id string) (*Goal, error)
	Update(ctx context.Context, g *Goal) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string) ([]*Goal, error)
}

// --- Coaching Meetings ---
type Meeting struct {
	ID              string     `json:"id"`
	StudentID       string     `json:"student_id"`
	CoachID         string     `json:"coach_id"`
	MeetingDate     time.Time  `json:"meeting_date"`
	DurationMinutes int        `json:"duration_minutes"`
	Notes           string     `json:"notes,omitempty"`
	MeetingURL      string     `json:"meeting_url,omitempty"`
	NextMeetingDate *time.Time `json:"next_meeting_date,omitempty"`
	Status          string     `json:"status"` // 'planned', 'completed', 'cancelled', 'postponed'
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type MeetingsRepository interface {
	Create(ctx context.Context, m *Meeting) error
	GetByID(ctx context.Context, id string) (*Meeting, error)
	Update(ctx context.Context, m *Meeting) error
	Delete(ctx context.Context, id string) error
	ListByStudent(ctx context.Context, studentID string) ([]*Meeting, error)
	ListByCoach(ctx context.Context, coachID string) ([]*Meeting, error)
}
