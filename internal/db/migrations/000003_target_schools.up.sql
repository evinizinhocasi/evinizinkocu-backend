-- Target schools table (Lise & Üniversite)
CREATE TABLE IF NOT EXISTS target_schools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'high_school' CHECK (type IN ('high_school', 'university')),
    city VARCHAR(100) NOT NULL DEFAULT '',
    photo_url TEXT NOT NULL DEFAULT '',
    min_score NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    percentile NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
    ranking INT NOT NULL DEFAULT 0,
    department VARCHAR(255) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_target_schools_type ON target_schools(type);
CREATE INDEX IF NOT EXISTS idx_target_schools_city ON target_schools(city);
CREATE INDEX IF NOT EXISTS idx_target_schools_is_active ON target_schools(is_active);

-- Add target school fields to students table
ALTER TABLE students ADD COLUMN IF NOT EXISTS target_school_id UUID REFERENCES target_schools(id) ON DELETE SET NULL;
ALTER TABLE students ADD COLUMN IF NOT EXISTS target_school_name VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE students ADD COLUMN IF NOT EXISTS target_school_photo TEXT NOT NULL DEFAULT '';
ALTER TABLE students ADD COLUMN IF NOT EXISTS target_school_type VARCHAR(50) NOT NULL DEFAULT '';
