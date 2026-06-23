-- Evinizin Koçu Database Schema Setup
-- Run as part of migrations

-- Enforce UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table: users
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(100) UNIQUE NOT NULL,
    phone VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('superadmin', 'coach', 'student')),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: refresh_tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE
);

-- Table: password_reset_codes
CREATE TABLE IF NOT EXISTS password_reset_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

-- Table: device_tokens
CREATE TABLE IF NOT EXISTS device_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    platform VARCHAR(20) NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_device_token UNIQUE(user_id, token)
);

-- Table: coaches
CREATE TABLE IF NOT EXISTS coaches (
    id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    city VARCHAR(100),
    biography TEXT,
    specialization VARCHAR(255),
    social_links JSONB DEFAULT '{}'::jsonb,
    student_capacity INT NOT NULL DEFAULT 5 CHECK (student_capacity >= 0),
    auth_start_date DATE NOT NULL,
    auth_end_date DATE NOT NULL,
    permission_immediate_push BOOLEAN NOT NULL DEFAULT FALSE,
    permission_scheduled_push BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: exam_types
CREATE TABLE IF NOT EXISTS exam_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) UNIQUE NOT NULL,
    divisor INT NOT NULL DEFAULT 4 CHECK (divisor > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: students
CREATE TABLE IF NOT EXISTS students (
    id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    coach_id UUID NOT NULL REFERENCES users(id),
    class_level VARCHAR(50) NOT NULL,
    study_track VARCHAR(50) NOT NULL CHECK (study_track IN ('Sayisal', 'Sozel', 'Esit Agirlik', 'Dil', 'LGS', 'Sayısal', 'Sözel', 'Eşit Ağırlık')),
    exam_type_id UUID NOT NULL REFERENCES exam_types(id),
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: coach_applications
CREATE TABLE IF NOT EXISTS coach_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    specialization VARCHAR(255) NOT NULL,
    explanation TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: coach_authorization_periods
CREATE TABLE IF NOT EXISTS coach_authorization_periods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    student_capacity INT NOT NULL CHECK (student_capacity >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: subjects
CREATE TABLE IF NOT EXISTS subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    exam_type_id UUID NOT NULL REFERENCES exam_types(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    question_count INT CHECK (question_count > 0),
    UNIQUE (exam_type_id, name)
);

-- Table: topics
CREATE TABLE IF NOT EXISTS topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    name VARCHAR(150) NOT NULL,
    UNIQUE (subject_id, name)
);

-- Table: trial_exams
CREATE TABLE IF NOT EXISTS trial_exams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id),
    exam_name VARCHAR(255) NOT NULL,
    exam_date DATE NOT NULL,
    exam_type_id UUID NOT NULL REFERENCES exam_types(id),
    total_net NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    score NUMERIC(8, 3),
    ranking INT CHECK (ranking > 0),
    coach_comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: trial_exam_subject_results
CREATE TABLE IF NOT EXISTS trial_exam_subject_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trial_exam_id UUID NOT NULL REFERENCES trial_exams(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    correct INT NOT NULL DEFAULT 0 CHECK (correct >= 0),
    incorrect INT NOT NULL DEFAULT 0 CHECK (incorrect >= 0),
    blank INT NOT NULL DEFAULT 0 CHECK (blank >= 0),
    net NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    UNIQUE(trial_exam_id, subject_id)
);

-- Table: question_solving_entries
CREATE TABLE IF NOT EXISTS question_solving_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    subject_id UUID NOT NULL REFERENCES subjects(id),
    topic_id UUID REFERENCES topics(id),
    correct INT NOT NULL DEFAULT 0 CHECK (correct >= 0),
    incorrect INT NOT NULL DEFAULT 0 CHECK (incorrect >= 0),
    blank INT NOT NULL DEFAULT 0 CHECK (blank >= 0),
    net NUMERIC(6, 2) NOT NULL DEFAULT 0.00,
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: homework
CREATE TABLE IF NOT EXISTS homework (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_coach_id UUID NOT NULL REFERENCES users(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    subject_id UUID NOT NULL REFERENCES subjects(id),
    topic_id UUID REFERENCES topics(id),
    source VARCHAR(255),
    page_range VARCHAR(100),
    url VARCHAR(2048),
    start_date DATE NOT NULL,
    due_date DATE NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'waiting' CHECK (status IN ('waiting', 'started', 'awaiting_approval', 'completed')),
    coach_explanation TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: resources
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    publisher VARCHAR(255) NOT NULL,
    subject_id UUID NOT NULL REFERENCES subjects(id),
    description TEXT,
    total_pages INT NOT NULL CHECK (total_pages > 0),
    completed_pages INT NOT NULL DEFAULT 0 CHECK (completed_pages >= 0),
    progress_percentage INT NOT NULL DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    status VARCHAR(20) NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'active', 'completed', 'paused')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_completed_pages CHECK (completed_pages <= total_pages)
);

-- Table: school_timetable_entries
CREATE TABLE IF NOT EXISTS school_timetable_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weekday INT NOT NULL CHECK (weekday >= 1 AND weekday <= 7),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    subject_name VARCHAR(150) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: weekly_plan_items
CREATE TABLE IF NOT EXISTS weekly_plan_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coach_id UUID NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    title VARCHAR(255) NOT NULL,
    subject_id UUID REFERENCES subjects(id),
    topic_id UUID REFERENCES topics(id),
    note TEXT,
    url VARCHAR(2048),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: missing_topics
CREATE TABLE IF NOT EXISTS missing_topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID NOT NULL REFERENCES subjects(id),
    topic_id UUID NOT NULL REFERENCES topics(id),
    description TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high')),
    status VARCHAR(20) NOT NULL DEFAULT 'identified' CHECK (status IN ('identified', 'in_progress', 'resolved')),
    target_date DATE,
    solution_text TEXT,
    url VARCHAR(2048),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: goals
CREATE TABLE IF NOT EXISTS goals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('question_count', 'exam_net', 'subject_net', 'resource_completion', 'custom')),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_value NUMERIC(10, 2) NOT NULL CHECK (target_value >= 0),
    current_value NUMERIC(10, 2) NOT NULL DEFAULT 0.00 CHECK (current_value >= 0),
    unit VARCHAR(50) NOT NULL,
    start_date DATE NOT NULL,
    target_date DATE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'achieved', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: meetings
CREATE TABLE IF NOT EXISTS meetings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coach_id UUID NOT NULL REFERENCES users(id),
    meeting_date TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    notes TEXT,
    meeting_url VARCHAR(2048),
    next_meeting_date TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'planned' CHECK (status IN ('planned', 'completed', 'cancelled', 'postponed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: coach_payments
CREATE TABLE IF NOT EXISTS coach_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    payment_date DATE NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('paid', 'pending', 'overdue', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: student_payments
CREATE TABLE IF NOT EXISTS student_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coach_id UUID NOT NULL REFERENCES users(id),
    amount NUMERIC(12, 2) NOT NULL CHECK (amount >= 0),
    payment_date DATE NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('paid', 'pending', 'overdue', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    sender_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: notification_recipients
CREATE TABLE IF NOT EXISTS notification_recipients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(notification_id, recipient_id)
);

-- Table: notification_schedules
CREATE TABLE IF NOT EXISTS notification_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coach_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    target_selection VARCHAR(50) NOT NULL CHECK (target_selection IN ('one', 'selected', 'all')),
    target_student_ids UUID[] DEFAULT '{}'::uuid[],
    schedule_type VARCHAR(20) NOT NULL CHECK (schedule_type IN ('one_time', 'daily', 'weekly')),
    selected_weekdays INT[] DEFAULT '{}'::int[],
    schedule_time TIME NOT NULL,
    start_date DATE,
    end_date DATE,
    next_run_at TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    locked_by VARCHAR(100),
    locked_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: notification_executions
CREATE TABLE IF NOT EXISTS notification_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID REFERENCES notification_schedules(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failed')),
    error_message TEXT,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: cms_settings
CREATE TABLE IF NOT EXISTS cms_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: faqs
CREATE TABLE IF NOT EXISTS faqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table: legal_documents
CREATE TABLE IF NOT EXISTS legal_documents (
    slug VARCHAR(100) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create essential indexes for performance and multi-tenant lookup isolation
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_students_coach ON students(coach_id);
CREATE INDEX IF NOT EXISTS idx_trial_exams_student ON trial_exams(student_id);
CREATE INDEX IF NOT EXISTS idx_question_solving_student ON question_solving_entries(student_id);
CREATE INDEX IF NOT EXISTS idx_homework_student ON homework(student_id);
CREATE INDEX IF NOT EXISTS idx_resources_student ON resources(student_id);
CREATE INDEX IF NOT EXISTS idx_weekly_plan_student ON weekly_plan_items(student_id);
CREATE INDEX IF NOT EXISTS idx_missing_topics_student ON missing_topics(student_id);
CREATE INDEX IF NOT EXISTS idx_goals_student ON goals(student_id);
CREATE INDEX IF NOT EXISTS idx_meetings_student ON meetings(student_id);
CREATE INDEX IF NOT EXISTS idx_meetings_coach ON meetings(coach_id);
CREATE INDEX IF NOT EXISTS idx_student_payments_student ON student_payments(student_id);
CREATE INDEX IF NOT EXISTS idx_coach_payments_coach ON coach_payments(coach_id);
CREATE INDEX IF NOT EXISTS idx_notification_recipients_user ON notification_recipients(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notification_schedules_run ON notification_schedules(next_run_at) WHERE is_active = TRUE;

-- Dynamic patch to drop and recreate constraint for backwards-compatibility on existing databases
ALTER TABLE students DROP CONSTRAINT IF EXISTS students_study_track_check;
ALTER TABLE students ADD CONSTRAINT students_study_track_check CHECK (study_track IN ('Sayisal', 'Sozel', 'Esit Agirlik', 'Dil', 'LGS', 'Sayısal', 'Sözel', 'Eşit Ağırlık'));
