# Project Rules - Evinizin Koçu

This document outlines the non-negotiable engineering rules, working methods, and design guidelines for **Evinizin Koçu**.

## 1. General Principles

* **Production-Oriented Code:** Write clean, maintainable, readable, and production-ready code. Avoid unnecessary abstractions, but enforce clear boundaries between layers.
* **SOLID Principles:** Apply SOLID principles throughout the codebase.
* **Secrets Management:** Never commit secrets, passwords, API keys, or private keys to version control. Use `.env` files for local development and provide `.env.example` templates in both frontend and backend directories.
* **Date & Time:**
  * All dates and times must be stored and processed as **UTC** in the database and backend.
  * All dates and times must be displayed to the user in the timezone **`Europe/Istanbul`** (UTC+3).
* **Primary Keys:** Use UUID v4 for all primary keys in the database.
* **Database Operations:**
  * Use database transactions (`BEGIN`, `COMMIT`, `ROLLBACK`) for any multi-step business operation.
  * Enforce data integrity using database constraints (foreign keys, unique constraints, check constraints, non-nullability) in addition to API-level validation.
* **Tenant Isolation:**
  * Prevent cross-coach data access at the repository/query and authorization layers.
  * A coach must never access another coach's student records, payments, plans, exams, notifications, or analytics.
  * Derive ownership and identity from the authenticated principal/token. Never trust `coach_id` parameters sent from the client.
* **API Error Envelopes:** Return consistent, machine-readable API error envelopes (e.g., standard code and description structure).
* **UI States:** Ensure all data lists and forms support pagination, filtering, sorting, loading, empty, validation, and retry/error states.

---

## 2. Backend Engineering Rules (Go)

* **Architecture:** Clean Architecture with:
  * **Transport Layer:** HTTP handlers, routing, middleware, request/response validation.
  * **Application/Use-case Layer:** Core business workflows, transaction boundaries.
  * **Domain Layer:** Entities, business rule validators, repository interfaces.
  * **Infrastructure Layer:** Database access (pgx), email sender (SMTP), FCM client, scheduled worker.
* **Database Access:** Use `pgx` directly or with a lightweight builder/helper for raw SQL. Do not use heavy ORMs. Enforce strict connection pooling and query timeouts.
* **SQL Migrations:** Use a standard migration tool (e.g., `golang-migrate/migrate` or `dbmate`) to manage migrations.
* **Routing:** Use a lightweight, fast, and production-ready HTTP router (e.g., `chi`).
* **Security:**
  * Use Argon2id or bcrypt (with secure parameters: bcrypt cost >= 12, Argon2id with 1 memory passes/parallelism parameters according to RFC 9106) for password hashing.
  * Implement rotating JWT access and refresh tokens. Revoke refresh tokens on logout or refresh rotation detection.
  * Normalize email, username, and phone numbers before checking for uniqueness.
  * Implement rate limiting (e.g., token bucket middleware) on critical endpoints (login, register, reset password, coach application).
* **Scheduled Workers:**
  * No Redis is allowed for scheduled notifications. Use PostgreSQL as the job queue.
  * Claim due jobs using `SELECT ... FOR UPDATE SKIP LOCKED` inside a transaction.
  * Enforce idempotency using execution records.
  * Re-verify coach active status and notification permission immediately before sending.

---

## 3. Frontend Engineering Rules (Flutter)

* **Architecture:** Full Clean Architecture:
  * `presentation`: Widgets, UI logic, Riverpod controllers/notifiers.
  * `domain`: Core domain models/entities, use-cases/repository definitions.
  * `data`: API DTOs, data sources, repository implementations.
* **State Management & DI:** Riverpod (with code generation where appropriate). Do not store business logic inside widgets.
* **Routing:** `go_router` for route guards, deep links, role-based redirects, and responsive web navigation.
* **Networking:** Use `Dio` with custom interceptors for attaching JWTs, token refresh, timeout handling (connect/send/receive timeouts), and normalized error parsing.
* **Storage:** Use `flutter_secure_storage` on Android/iOS. Use a secure storage strategy on Web (e.g., in-memory JWT storage, combined with HTTP-Only secure cookies from backend or local session security).
* **Design Language:**
  * **Turkish UI language only** for version one.
  * **Light theme only** for version one.
  * Education-focused professional blue palette, white surfaces, restrained neutral grays, and accessible status colors.
  * Avoid emojis as UI icons. Use a professional package such as Material Symbols or Lucide icons.
  * Breakpoints: Enforce support for typical viewport widths (360px, 390px, 768px, 1024px, 1440px).
* **Reusable Widgets:**
  * Build reusable, configurable widgets for forms, fields, pickers, dialogs, sheets, buttons, tables, and pagination. Do not duplicate these files across features.

---

## 4. Coding & Linting Guidelines

* Run `go fmt` and `go vet` before checking in Go code. Enforce strict lint rules (e.g., `golangci-lint` if configured).
* Run `flutter format` and `flutter analyze` on Dart code. Maintain a clean `analysis_options.yaml`.
* Write tests to cover critical rules. Run tests using `go test ./...` and `flutter test`.
