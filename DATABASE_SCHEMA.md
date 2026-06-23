# Database Schema - Evinizin Koçu

All tables must reside in PostgreSQL. Primary keys are UUID v4. Timestamps are stored in UTC.

## 1. Table Definitions

### `users`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `email` VARCHAR(255) UNIQUE NOT NULL -- normalized to lowercase
* `username` VARCHAR(100) UNIQUE NOT NULL -- normalized to lowercase
* `phone` VARCHAR(50) UNIQUE NOT NULL -- normalized (only digits/E.164)
* `password_hash` VARCHAR(255) NOT NULL
* `first_name` VARCHAR(100) NOT NULL
* `last_name` VARCHAR(100) NOT NULL
* `role` VARCHAR(20) NOT NULL -- 'superadmin', 'coach', 'student'
* `is_active` BOOLEAN NOT NULL DEFAULT TRUE
* `must_change_password` BOOLEAN NOT NULL DEFAULT FALSE
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `refresh_tokens`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `user_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `token` VARCHAR(255) UNIQUE NOT NULL
* `expires_at` TIMESTAMPTZ NOT NULL
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `is_revoked` BOOLEAN NOT NULL DEFAULT FALSE

### `password_reset_codes`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `user_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `code_hash` VARCHAR(255) NOT NULL
* `expires_at` TIMESTAMPTZ NOT NULL
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `used_at` TIMESTAMPTZ

### `device_tokens`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `user_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `token` VARCHAR(255) NOT NULL
* `platform` VARCHAR(20) NOT NULL -- 'android', 'ios', 'web'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* CONSTRAINT unique_user_device_token UNIQUE(user_id, token)

### `coaches`
* `id` UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
* `city` VARCHAR(100)
* `biography` TEXT
* `specialization` VARCHAR(255)
* `social_links` JSONB -- e.g. {"instagram": "...", "linkedin": "..."}
* `student_capacity` INT NOT NULL DEFAULT 5
* `auth_start_date` DATE NOT NULL
* `auth_end_date` DATE NOT NULL
* `permission_immediate_push` BOOLEAN NOT NULL DEFAULT FALSE
* `permission_scheduled_push` BOOLEAN NOT NULL DEFAULT FALSE
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `students`
* `id` UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE
* `coach_id` UUID NOT NULL REFERENCES users(id) -- The current coach. Must resolve to a user with role = 'coach'
* `class_level` VARCHAR(50) NOT NULL -- e.g. '12', 'mezun'
* `study_track` VARCHAR(50) NOT NULL -- 'Sayisal', 'Sozel', 'Esit Agirlik', 'Dil'
* `exam_type_id` UUID NOT NULL -- references exam_types(id)
* `is_archived` BOOLEAN NOT NULL DEFAULT FALSE
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `coach_applications`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `first_name` VARCHAR(100) NOT NULL
* `last_name` VARCHAR(100) NOT NULL
* `phone` VARCHAR(50) NOT NULL
* `email` VARCHAR(255) NOT NULL
* `city` VARCHAR(100) NOT NULL
* `specialization` VARCHAR(255) NOT NULL
* `explanation` TEXT NOT NULL
* `status` VARCHAR(20) NOT NULL DEFAULT 'pending' -- 'pending', 'approved', 'rejected'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `coach_authorization_periods`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `coach_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `start_date` DATE NOT NULL
* `end_date` DATE NOT NULL
* `student_capacity` INT NOT NULL
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `exam_types`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `name` VARCHAR(100) UNIQUE NOT NULL -- e.g. 'LGS', 'TYT', 'AYT', 'YDT'
* `divisor` INT NOT NULL DEFAULT 4 -- Net divisor: e.g. 4 for 4-wrong-1-right penalty (LGS is 3, YKS is 4)
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `subjects`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `exam_type_id` UUID NOT NULL REFERENCES exam_types(id) ON DELETE CASCADE
* `name` VARCHAR(100) NOT NULL
* `question_count` INT -- config for total questions in this exam subject
* UNIQUE (exam_type_id, name)

### `topics`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `subject_id` UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE
* `name` VARCHAR(150) NOT NULL
* UNIQUE (subject_id, name)

### `trial_exams`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `creator_id` UUID NOT NULL REFERENCES users(id)
* `exam_name` VARCHAR(255) NOT NULL
* `exam_date` DATE NOT NULL
* `exam_type_id` UUID NOT NULL REFERENCES exam_types(id)
* `total_net` NUMERIC(6, 2) NOT NULL DEFAULT 0.00
* `score` NUMERIC(8, 3)
* `ranking` INT
* `coach_comment` TEXT
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `trial_exam_subject_results`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `trial_exam_id` UUID NOT NULL REFERENCES trial_exams(id) ON DELETE CASCADE
* `subject_id` UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE
* `correct` INT NOT NULL DEFAULT 0
* `incorrect` INT NOT NULL DEFAULT 0
* `blank` INT NOT NULL DEFAULT 0
* `net` NUMERIC(6, 2) NOT NULL DEFAULT 0.00
* UNIQUE(trial_exam_id, subject_id)

### `question_solving_entries`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `creator_id` UUID NOT NULL REFERENCES users(id)
* `date` DATE NOT NULL
* `subject_id` UUID NOT NULL REFERENCES subjects(id)
* `topic_id` UUID REFERENCES topics(id)
* `correct` INT NOT NULL DEFAULT 0
* `incorrect` INT NOT NULL DEFAULT 0
* `blank` INT NOT NULL DEFAULT 0
* `net` NUMERIC(6, 2) NOT NULL DEFAULT 0.00
* `note` TEXT
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `homework`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `creator_coach_id` UUID NOT NULL REFERENCES users(id)
* `title` VARCHAR(255) NOT NULL
* `description` TEXT
* `subject_id` UUID NOT NULL REFERENCES subjects(id)
* `topic_id` UUID REFERENCES topics(id)
* `source` VARCHAR(255)
* `page_range` VARCHAR(100)
* `url` VARCHAR(2048)
* `start_date` DATE NOT NULL
* `due_date` DATE NOT NULL
* `status` VARCHAR(30) NOT NULL DEFAULT 'waiting' -- 'waiting', 'started', 'awaiting_approval', 'completed'
* `coach_explanation` TEXT -- for reject reasons
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `resources`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `name` VARCHAR(255) NOT NULL
* `publisher` VARCHAR(255) NOT NULL
* `subject_id` UUID NOT NULL REFERENCES subjects(id)
* `description` TEXT
* `total_pages` INT NOT NULL
* `completed_pages` INT NOT NULL DEFAULT 0
* `progress_percentage` INT NOT NULL DEFAULT 0 -- manual override or calculated
* `status` VARCHAR(20) NOT NULL DEFAULT 'planned' -- 'planned', 'active', 'completed', 'paused'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `school_timetable_entries`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `weekday` INT NOT NULL -- 1 = Pazartesi, 7 = Pazar
* `start_time` TIME NOT NULL
* `end_time` TIME NOT NULL
* `subject_name` VARCHAR(150) NOT NULL
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `weekly_plan_items`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `coach_id` UUID NOT NULL REFERENCES users(id)
* `date` DATE NOT NULL
* `start_time` TIME NOT NULL
* `end_time` TIME NOT NULL
* `title` VARCHAR(255) NOT NULL
* `subject_id` UUID REFERENCES subjects(id)
* `topic_id` UUID REFERENCES topics(id)
* `note` TEXT
* `url` VARCHAR(2048)
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `missing_topics`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `subject_id` UUID NOT NULL REFERENCES subjects(id)
* `topic_id` UUID NOT NULL REFERENCES topics(id)
* `description` TEXT
* `priority` VARCHAR(20) NOT NULL DEFAULT 'medium' -- 'low', 'medium', 'high'
* `status` VARCHAR(20) NOT NULL DEFAULT 'identified' -- 'identified', 'in_progress', 'resolved'
* `target_date` DATE
* `solution_text` TEXT
* `url` VARCHAR(2048)
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `goals`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `type` VARCHAR(50) NOT NULL -- 'question_count', 'exam_net', 'subject_net', 'resource_completion', 'custom'
* `title` VARCHAR(255) NOT NULL
* `description` TEXT
* `target_value` NUMERIC(10, 2) NOT NULL
* `current_value` NUMERIC(10, 2) NOT NULL DEFAULT 0.00
* `unit` VARCHAR(50) NOT NULL
* `start_date` DATE NOT NULL
* `target_date` DATE NOT NULL
* `status` VARCHAR(20) NOT NULL DEFAULT 'active' -- 'active', 'achieved', 'failed'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `meetings`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `coach_id` UUID NOT NULL REFERENCES users(id)
* `meeting_date` TIMESTAMPTZ NOT NULL
* `duration_minutes` INT NOT NULL
* `notes` TEXT
* `meeting_url` VARCHAR(2048)
* `next_meeting_date` TIMESTAMPTZ
* `status` VARCHAR(20) NOT NULL DEFAULT 'planned' -- 'planned', 'completed', 'cancelled', 'postponed'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `coach_payments`
* `id` PRIMARY KEY UUID DEFAULT gen_random_uuid()
* `coach_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `amount` NUMERIC(12, 2) NOT NULL CHECK(amount >= 0)
* `payment_date` DATE NOT NULL
* `description` TEXT
* `status` VARCHAR(20) NOT NULL DEFAULT 'pending' -- 'paid', 'pending', 'overdue', 'cancelled'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `student_payments`
* `id` PRIMARY KEY UUID DEFAULT gen_random_uuid()
* `student_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `coach_id` UUID NOT NULL REFERENCES users(id)
* `amount` NUMERIC(12, 2) NOT NULL CHECK(amount >= 0)
* `payment_date` DATE NOT NULL
* `description` TEXT
* `status` VARCHAR(20) NOT NULL DEFAULT 'pending' -- 'paid', 'pending', 'overdue', 'cancelled'
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `notifications`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `title` VARCHAR(255) NOT NULL
* `body` TEXT NOT NULL
* `sender_id` UUID NOT NULL REFERENCES users(id)
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `notification_recipients`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `notification_id` UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE
* `recipient_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `is_read` BOOLEAN NOT NULL DEFAULT FALSE
* `read_at` TIMESTAMPTZ
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* UNIQUE(notification_id, recipient_id)

### `notification_schedules`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `coach_id` UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE
* `title` VARCHAR(255) NOT NULL
* `body` TEXT NOT NULL
* `target_selection` VARCHAR(50) NOT NULL -- 'one', 'selected', 'all'
* `target_student_ids` UUID[] -- array of student IDs
* `schedule_type` VARCHAR(20) NOT NULL -- 'one_time', 'daily', 'weekly'
* `selected_weekdays` INT[] -- 1-7 representing Pazartesi-Pazar
* `schedule_time` TIME NOT NULL -- e.g. '14:30:00' (interpreted in Europe/Istanbul, saved in UTC)
* `start_date` DATE
* `end_date` DATE
* `next_run_at` TIMESTAMPTZ NOT NULL
* `is_active` BOOLEAN NOT NULL DEFAULT TRUE
* `locked_by` VARCHAR(100)
* `locked_until` TIMESTAMPTZ
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `notification_executions`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `schedule_id` UUID REFERENCES notification_schedules(id) ON DELETE SET NULL
* `status` VARCHAR(20) NOT NULL -- 'success', 'failed'
* `error_message` TEXT
* `executed_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `cms_settings`
* `key` VARCHAR(100) PRIMARY KEY
* `value` JSONB NOT NULL
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `faqs`
* `id` UUID PRIMARY KEY DEFAULT gen_random_uuid()
* `question` TEXT NOT NULL
* `answer` TEXT NOT NULL
* `display_order` INT NOT NULL DEFAULT 0
* `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

### `legal_documents`
* `slug` VARCHAR(100) PRIMARY KEY -- e.g. 'privacy-policy', 'terms-of-use', 'kvkk'
* `title` VARCHAR(255) NOT NULL
* `content` TEXT NOT NULL
* `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()

---

## 2. Key Constraints & Guarantees
1. **Uniqueness:** Lowercase email, username, and digits-only phone numbers must be globally unique across all roles.
2. **Tenant Isolation:** All queries for coach data MUST filter by `coach_id`. All student-specific tables have `student_id`.
3. **Validations:** Academic correct/incorrect/blank fields have `CHECK(correct >= 0 AND incorrect >= 0 AND blank >= 0)`.
4. **Capacities:** Checking coach limits is protected using atomic transactions and row locks.
5. **No deletion bypass:** Soft delete/archive fields are used to avoid direct data loss, except superadmin-initiated final student removal.
