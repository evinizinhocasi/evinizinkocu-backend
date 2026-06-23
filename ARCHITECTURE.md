# Architecture Design - Evinizin Koçu

This document outlines the architecture, directory structure, state management, and multi-tenant isolation design for both the backend and frontend.

## 1. Directory Structure

### Backend (`evinizinkocu-backend`)
We use Clean Architecture.
```
evinizinkocu-backend/
├── cmd/
│   └── api/                # Application entrypoint
│       └── main.go
├── internal/
│   ├── config/             # Configuration & environment variables
│   ├── db/                 # DB connection and migrations
│   ├── middleware/         # Auth, tenant, logger, rate-limiter, cors
│   ├── worker/             # PostgreSQL claim worker (scheduler)
│   ├── domain/             # Entities, custom errors, repository interfaces
│   │   ├── user.go
│   │   ├── coach.go
│   │   ├── student.go
│   │   ├── catalog.go
│   │   └── notification.go
│   ├── transport/          # HTTP handlers and JSON encoders/decoders
│   │   ├── http/
│   │   │   ├── handler.go  # Core handler with route setups
│   │   │   ├── auth.go
│   │   │   ├── student.go
│   │   │   └── ...
│   ├── application/        # Use cases / services implementing business logic
│   │   ├── auth_service.go
│   │   ├── coach_service.go
│   │   └── ...
│   └── infrastructure/     # SQL DB queries, SMTP client, FCM service
│   │   ├── repository/     # SQL-based repo implementations
│   │   ├── mailer/         # SMTP-based mailer
│   │   └── fcm/            # FCM HTTP v1 implementation
├── migrations/             # SQL Migration files (.sql)
├── scripts/                # Utility & seed scripts
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

### Frontend (`evinizinkocu`)
We organize the frontend by features alongside a `core` directory containing shared logic.
```
evinizinkocu/
├── android/
├── ios/
├── web/
├── assets/
│   ├── fonts/
│   └── icons/
├── lib/
│   ├── main.dart
│   ├── core/               # Shared logic across features
│   │   ├── theme/          # Typography, colors, breakpoints, borders
│   │   ├── widgets/        # Reusable design system widgets (inputs, cards, tables, etc.)
│   │   ├── network/        # Dio client, interceptors, error mapping
│   │   ├── router/         # go_router configurations, guards
│   │   ├── storage/        # secure storage service
│   │   └── utils/          # dates, numbers, validators
│   └── features/           # Feature-based clean architecture
│       ├── public/         # Hero, features, CMS contact page, coach application form
│       ├── auth/           # Login, change password, forgot password
│       ├── superadmin/     # Coach management, capacity, payments, settings
│       ├── coach/          # Student management, plans, goals, homework, meetings
│       ├── student/        # Solved questions, trial exams, weekly plan view, resources
│       └── notifications/  # Notification center, inbox
├── pubspec.yaml
└── analysis_options.yaml
```

---

## 2. Multi-Tenant Isolation (Coach Level)

Every student belongs to exactly one coach.
* A coach is treated as a tenant.
* A coach MUST only access their own students and related database rows (homework, trial exams, goals, timetable, plan, payments, etc.).
* Tenant isolation is enforced in three places:
  1. **Token Claims:** The JWT token contains `user_id`, `role`, and `coach_id` (if the logged-in user is a coach or student).
  2. **Database Queries:** Repository queries targeting tenant data must filter by `coach_id` (for coach queries) or direct `student_id` (for student queries).
  3. **Middleware checks:** For endpoints operating on a student (e.g. `/api/v1/students/:id/homework`), a middleware/use-case check enforces that the student's `coach_id` matches the authenticated coach's ID.

---

## 3. Scheduled Notifications Design

The scheduling mechanism operates without Redis.
* Schedules are stored in `notification_schedules`.
* Jobs that are ready to run are processed by a recurring Go ticker/worker.
* To support multiple backend instances safely, we use PostgreSQL row locks:
  ```sql
  -- Claims due jobs atomically
  UPDATE notification_schedules
  SET locked_by = $1, locked_until = $2
  WHERE id IN (
      SELECT id FROM notification_schedules
      WHERE next_run_at <= NOW() AND is_active = true AND (locked_until IS NULL OR locked_until < NOW())
      LIMIT 10
      FOR UPDATE SKIP LOCKED
  )
  RETURNING *;
  ```
* Before execution, the worker checks if the coach account is still active and has scheduled notifications permission.
* After sending, the worker updates the `next_run_at` based on the schedule definition (e.g. daily, weekly weekdays) or marks it inactive if single-use, and writes to `notification_executions` for history and idempotency protection.
