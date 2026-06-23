# API Specification - Evinizin Koçu

All endpoints are versioned under `/api/v1` and speak JSON.

## 1. Authentication & User Management

### `POST /api/v1/auth/login`
* **Description:** Logs in a user (superadmin, coach, student).
* **Request:**
  ```json
  {
    "identifier": "user@email.com or username",
    "password": "mypassword"
  }
  ```
* **Response (200 OK):**
  ```json
  {
    "access_token": "jwt_access_token_here",
    "refresh_token": "jwt_refresh_token_here",
    "user": {
      "id": "uuid",
      "email": "user@email.com",
      "username": "username",
      "first_name": "John",
      "last_name": "Doe",
      "role": "superadmin|coach|student",
      "must_change_password": false
    }
  }
  ```

### `POST /api/v1/auth/refresh`
* **Description:** Rotates JWT token.
* **Request:**
  ```json
  {
    "refresh_token": "jwt_refresh_token_here"
  }
  ```
* **Response (200 OK):**
  ```json
  {
    "access_token": "jwt_access_token_here",
    "refresh_token": "jwt_refresh_token_here"
  }
  ```

### `POST /api/v1/auth/logout`
* **Description:** Revokes current refresh token and clears device token.
* **Request:**
  ```json
  {
    "refresh_token": "jwt_refresh_token_here",
    "device_token": "fcm_token_optional"
  }
  ```
* **Response (204 No Content)**

### `POST /api/v1/auth/change-password`
* **Description:** Forces first-login temporary password change or regular profile password changes.
* **Headers:** `Authorization: Bearer <token>`
* **Request:**
  ```json
  {
    "current_password": "temppassword",
    "new_password": "newstrongpassword"
  }
  ```
* **Response (200 OK)**

### `POST /api/v1/auth/forgot-password`
* **Description:** Requests numeric verification reset code via email.
* **Request:**
  ```json
  {
    "email": "user@email.com"
  }
  ```
* **Response (200 OK)**

### `POST /api/v1/auth/reset-password`
* **Description:** Resets password using verification code.
* **Request:**
  ```json
  {
    "email": "user@email.com",
    "code": "123456",
    "new_password": "newstrongpassword"
  }
  ```
* **Response (200 OK)**

---

## 2. Coach Profile Management

### `GET /api/v1/coach/profile`
* **Role:** Coach
* **Response:** Profile details.

### `PUT /api/v1/coach/profile`
* **Role:** Coach
* **Request:** `first_name`, `last_name`, `phone`, `city`, `biography`, `specialization`, `social_links`.

---

## 3. Student Management (Coach & Superadmin)

### `GET /api/v1/students`
* **Role:** Coach (gets their students only), Superadmin (gets all).
* **Response:** Page/cursor list of students.

### `POST /api/v1/students`
* **Role:** Coach (automatically assigned to them), Superadmin (explicitly selects coach).
* **Request:**
  ```json
  {
    "first_name": "First",
    "last_name": "Last",
    "email": "student@email.com",
    "phone": "905001234567",
    "class_level": "12",
    "study_track": "Sayisal",
    "exam_type_id": "uuid",
    "coach_id": "uuid" // Optional for coach, mandatory for superadmin
  }
  ```

### `POST /api/v1/students/:id/transfer`
* **Role:** Superadmin
* **Request:** `{"new_coach_id": "uuid"}`

---

## 4. Academic Modules

### `GET /api/v1/catalog`
* **Description:** Fetch structured catalog data (exams, tracks, subjects, topics).

### `POST /api/v1/trial-exams`
* **Role:** Student (their own), Coach (their students).
* **Request:**
  ```json
  {
    "student_id": "uuid",
    "exam_name": "Deneme A",
    "exam_date": "2026-06-20",
    "exam_type_id": "uuid",
    "score": 420.50,
    "ranking": 12,
    "coach_comment": "Excellent!", // Only editable by coach
    "results": [
      {
        "subject_id": "uuid",
        "correct": 18,
        "incorrect": 2,
        "blank": 0
      }
    ]
  }
  ```

### `POST /api/v1/question-solving`
* **Role:** Student (their own), Coach (their students).
* **Request:** `student_id`, `date`, `subject_id`, `topic_id`, `correct`, `incorrect`, `blank`, `note`.

---

## 5. Homework & Resources

### `POST /api/v1/homework`
* **Role:** Coach
* **Request:** `student_id`, `title`, `description`, `subject_id`, `topic_id`, `source`, `page_range`, `url`, `start_date`, `due_date`.

### `PUT /api/v1/homework/:id/status`
* **Request:** `{"status": "started|awaiting_approval|completed"}`
* **Rules:**
  * Student can change to `started` or `awaiting_approval`.
  * Coach can change to `completed` or reject back to `started` with an optional `coach_explanation`.

---

## 6. Scheduled Notifications & Tokens

### `POST /api/v1/device-tokens`
* **Request:** `{"token": "fcm_token", "platform": "android|ios|web"}`

### `POST /api/v1/notifications/schedule`
* **Role:** Coach (requires scheduled permission)
* **Request:**
  ```json
  {
    "title": "Title",
    "body": "Body",
    "target_selection": "one|selected|all",
    "target_student_ids": ["uuid"],
    "schedule_type": "one_time|daily|weekly",
    "selected_weekdays": [1, 3, 5],
    "schedule_time": "14:30:00",
    "start_date": "2026-06-20",
    "end_date": "2026-08-20"
  }
  ```

---

## 7. Public CMS & Coach Applications

### `POST /api/v1/public/apply`
* **Description:** Public coach application.
* **Request:** `first_name`, `last_name`, `phone`, `email`, `city`, `specialization`, `explanation`.

### `GET /api/v1/public/settings`
* **Description:** Public CMS settings, FAQs, legal files.
