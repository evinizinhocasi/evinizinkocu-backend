CREATE TABLE IF NOT EXISTS wrong_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id UUID REFERENCES subjects(id) ON DELETE SET NULL,
    image_url TEXT NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    is_resolved BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wrong_questions_student_id ON wrong_questions(student_id);
CREATE INDEX IF NOT EXISTS idx_wrong_questions_subject_id ON wrong_questions(subject_id);
