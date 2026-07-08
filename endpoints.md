# E-Dossier API Endpoints Documentation

**Base URL:** `http://localhost:8080/api/v1`

**Authentication:** All endpoints marked with 🔐 require a valid JWT token in the `Authorization` header.

---

## Table of Contents

1. [Academic Sessions](#academic-sessions)
2. [Levels](#levels)
3. [Subjects](#subjects)
4. [Enrollments](#enrollments)
5. [Personnel](#personnel)
6. [Sub-Levels (Class Arms)](#sub-levels)
7. [Results & Scores](#results--scores)
8. [Report Cards](#report-cards)
9. [Score & Grade Configuration](#score--grade-configuration)
10. [Avatars](#avatars)

---

## Academic Sessions

### List All Sessions
- **URL:** `GET /sessions`
- **Auth:** 🔐 Required (`sessions:read`)
- **Description:** Retrieve all academic sessions
- **Query Parameters:**
  - `page` (optional): Page number for pagination
  - `limit` (optional): Records per page
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "2023/2024",
        "start_date": "2023-09-01",
        "end_date": "2024-06-30",
        "is_active": true,
        "created_at": "2023-08-01T00:00:00Z"
      }
    ],
    "pagination": {
      "total": 10,
      "page": 1,
      "limit": 10
    }
  }
  ```

### Get Active Session
- **URL:** `GET /sessions/active`
- **Auth:** 🔐 Required (`sessions:read`)
- **Description:** Get the currently active academic session
- **Response:**
  ```json
  {
    "id": "uuid",
    "name": "2023/2024",
    "start_date": "2023-09-01",
    "end_date": "2024-06-30",
    "is_active": true
  }
  ```

### Get Session by ID
- **URL:** `GET /sessions/{id}`
- **Auth:** 🔐 Required (`sessions:read`)
- **Description:** Retrieve a specific session with its terms
- **Response:**
  ```json
  {
    "id": "uuid",
    "name": "2023/2024",
    "start_date": "2023-09-01",
    "end_date": "2024-06-30",
    "is_active": true,
    "terms": [
      {
        "id": "uuid",
        "name": "First Term",
        "start_date": "2023-09-01",
        "end_date": "2023-12-22",
        "is_active": true
      }
    ]
  }
  ```

### Create Session
- **URL:** `POST /sessions`
- **Auth:** 🔐 Required (`sessions:create`)
- **Description:** Create a new academic session
- **Request Body:**
  ```json
  {
    "name": "2024/2025",
    "start_date": "2024-09-01",
    "end_date": "2025-06-30"
  }
  ```

### Update Session
- **URL:** `PUT /sessions/{id}`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Update session details
- **Request Body:**
  ```json
  {
    "name": "2024/2025",
    "start_date": "2024-09-01",
    "end_date": "2025-06-30"
  }
  ```

### Activate Session
- **URL:** `POST /sessions/{id}/activate`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Set a session as active
- **Request Body:** `{}`

### Delete Session
- **URL:** `DELETE /sessions/{id}`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Delete an academic session

---

## Session Terms

### List Terms in Session
- **URL:** `GET /sessions/{sessionId}/terms`
- **Auth:** 🔐 Required (`sessions:read`)
- **Description:** Get all terms for a specific session
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "session_id": "uuid",
        "name": "First Term",
        "start_date": "2023-09-01",
        "end_date": "2023-12-22",
        "is_active": true
      }
    ]
  }
  ```

### Create Term
- **URL:** `POST /sessions/{sessionId}/terms`
- **Auth:** 🔐 Required (`sessions:create`)
- **Description:** Create a new term within a session
- **Request Body:**
  ```json
  {
    "name": "First Term",
    "start_date": "2023-09-01",
    "end_date": "2023-12-22"
  }
  ```

### Update Term
- **URL:** `PUT /sessions/{sessionId}/terms/{id}`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Update term details

### Activate Term
- **URL:** `POST /sessions/{sessionId}/terms/{id}/activate`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Set a term as active

### Delete Term
- **URL:** `DELETE /sessions/{sessionId}/terms/{id}`
- **Auth:** 🔐 Required (`sessions:update`)
- **Description:** Delete a term

---

## Top-Level Terms

The routes below manage terms independently of a session URL, for admin/global
views. Creation still requires the owning `session_id` in the request body.

### List All Terms
- **URL:** `GET /terms`
- **Auth:** 🔐 Required (`sessions:read`)
- **Description:** Get all terms across every session

### Create Term (top-level)
- **URL:** `POST /terms`
- **Auth:** 🔐 Required (`sessions:create`)
- **Description:** Create a term; the owning session is supplied in the body
- **Request Body:**
  ```json
  {
    "session_id": "uuid",
    "term_number": 1,
    "name": "First Term",
    "start_date": "2023-09-01",
    "end_date": "2023-12-22"
  }
  ```

### Get Term
- **URL:** `GET /terms/{id}`
- **Auth:** 🔐 Required (`sessions:read`)

### Update Term (top-level)
- **URL:** `PUT /terms/{id}`
- **Auth:** 🔐 Required (`sessions:update`)

### Delete Term (top-level)
- **URL:** `DELETE /terms/{id}`
- **Auth:** 🔐 Required (`sessions:update`)

---

## Levels

### List All Levels
- **URL:** `GET /levels`
- **Auth:** 🔐 Required
- **Description:** Get all class levels (state-wide definitions)
- **Query Parameters:**
  - `page` (optional)
  - `limit` (optional)
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "JSS 1",
        "code": "JSS1",
        "description": "Junior Secondary School Year 1",
        "created_at": "2023-01-01T00:00:00Z"
      }
    ]
  }
  ```

### Get Level by ID
- **URL:** `GET /levels/{id}`
- **Auth:** 🔐 Required
- **Description:** Retrieve a specific level

### Create Level
- **URL:** `POST /levels`
- **Auth:** 🔐 Required (`schools:create`)
- **Request Body:**
  ```json
  {
    "name": "JSS 1",
    "code": "JSS1",
    "description": "Junior Secondary School Year 1"
  }
  ```

### Update Level
- **URL:** `PUT /levels/{id}`
- **Auth:** 🔐 Required (`schools:update`)
- **Request Body:**
  ```json
  {
    "name": "JSS 1",
    "code": "JSS1",
    "description": "Updated description"
  }
  ```

### Delete Level
- **URL:** `DELETE /levels/{id}`
- **Auth:** 🔐 Required (`schools:delete`)

### List Sub-Levels by Level
- **URL:** `GET /levels/{levelId}/sub-levels`
- **Auth:** 🔐 Required
- **Description:** Get all class arms (sub-levels) under a specific level
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "level_id": "uuid",
        "name": "JSS 1A",
        "code": "JSS1A",
        "capacity": 50
      }
    ]
  }
  ```

---

## Subjects

### List All Subjects
- **URL:** `GET /subjects`
- **Auth:** 🔐 Required
- **Description:** Get all subjects (state-wide)
- **Query Parameters:**
  - `page` (optional)
  - `limit` (optional)
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "Mathematics",
        "code": "MAT",
        "description": "Mathematics subject"
      }
    ]
  }
  ```

### Get Subject by ID
- **URL:** `GET /subjects/{id}`
- **Auth:** 🔐 Required

### Create Subject
- **URL:** `POST /subjects`
- **Auth:** 🔐 Required (`schools:create`)
- **Request Body:**
  ```json
  {
    "name": "Mathematics",
    "code": "MAT",
    "description": "Mathematics subject"
  }
  ```

### Update Subject
- **URL:** `PUT /subjects/{id}`
- **Auth:** 🔐 Required (`schools:update`)

### Delete Subject
- **URL:** `DELETE /subjects/{id}`
- **Auth:** 🔐 Required (`schools:delete`)

---

## Sub-Levels (Class Arms)

### Create Sub-Level
- **URL:** `POST /schools/{schoolId}/sub-levels`
- **Auth:** 🔐 Required (`schools:update`)
- **Description:** Create a new class arm for a school
- **Request Body:**
  ```json
  {
    "level_id": "uuid",
    "name": "JSS 1A",
    "code": "JSS1A",
    "capacity": 50
  }
  ```

### Update Sub-Level
- **URL:** `PUT /sub-levels/{id}`
- **Auth:** 🔐 Required (`schools:update`)
- **Request Body:**
  ```json
  {
    "name": "JSS 1A",
    "code": "JSS1A",
    "capacity": 45
  }
  ```

### Delete Sub-Level
- **URL:** `DELETE /sub-levels/{id}`
- **Auth:** 🔐 Required (`schools:update`)

### List Sub-Levels by School
- **URL:** `GET /schools/{schoolId}/sub-levels`
- **Auth:** 🔐 Required
- **Description:** Get all class arms in a school

---

## Personnel

### List All Personnel
- **URL:** `GET /personnel`
- **Auth:** 🔐 Required (`personnel:read`)
- **Description:** Get all staff members
- **Query Parameters:**
  - `page` (optional)
  - `limit` (optional)
  - `school_id` (optional): Filter by school
  - `status` (optional): Filter by status (active, inactive, transferred)
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "first_name": "John",
        "last_name": "Doe",
        "email": "john.doe@school.edu",
        "phone": "+234812345678",
        "employee_id": "EMP001",
        "position": "Teacher",
        "department": "Science",
        "school_id": "uuid",
        "date_of_employment": "2020-01-15",
        "status": "active"
      }
    ]
  }
  ```

### Get Personnel by ID
- **URL:** `GET /personnel/{id}`
- **Auth:** 🔐 Required (`personnel:read`)

### Update Personnel
- **URL:** `PUT /personnel/{id}`
- **Auth:** 🔐 Required (`personnel:update`)
- **Request Body:**
  ```json
  {
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@school.edu",
    "phone": "+234812345678",
    "position": "Senior Teacher",
    "department": "Science"
  }
  ```

### Delete Personnel
- **URL:** `DELETE /personnel/{id}`
- **Auth:** 🔐 Required (`personnel:delete`)

### Transfer Personnel
- **URL:** `POST /personnel/{id}/transfer`
- **Auth:** 🔐 Required (`personnel:update`)
- **Description:** Transfer staff to a different school
- **Request Body:**
  ```json
  {
    "new_school_id": "uuid",
    "transfer_date": "2024-01-01",
    "reason": "Promotion"
  }
  ```

### List Personnel Transfers
- **URL:** `GET /personnel/{id}/transfers`
- **Auth:** 🔐 Required (`personnel:read`)
- **Description:** Get transfer history for a staff member

---

## Enrollments

### List All Enrollments
- **URL:** `GET /enrollments`
- **Auth:** 🔐 Required (`enrollments:read`)
- **Description:** Get all student enrollments
- **Query Parameters:**
  - `page` (optional)
  - `limit` (optional)
  - `student_id` (optional): Filter by student
  - `sub_level_id` (optional): Filter by class
  - `status` (optional): active, inactive
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "student_id": "uuid",
        "sub_level_id": "uuid",
        "session_id": "uuid",
        "enrollment_date": "2023-09-01",
        "status": "active"
      }
    ]
  }
  ```

### Enroll Student
- **URL:** `POST /enrollments`
- **Auth:** 🔐 Required (`enrollments:create`)
- **Description:** Enroll a student in a class
- **Request Body:**
  ```json
  {
    "student_id": "uuid",
    "sub_level_id": "uuid",
    "session_id": "uuid",
    "enrollment_date": "2023-09-01"
  }
  ```

### Update Enrollment
- **URL:** `PUT /enrollments/{enrollmentId}`
- **Auth:** 🔐 Required (`enrollments:update`)
- **Request Body:**
  ```json
  {
    "sub_level_id": "uuid",
    "status": "active"
  }
  ```

---

## Results & Scores

### Upsert Score
- **URL:** `POST /results/scores`
- **Auth:** 🔐 Required (`results:create`)
- **Description:** Enter or update a student's score for a subject
- **Request Body:**
  ```json
  {
    "student_id": "uuid",
    "subject_id": "uuid",
    "session_id": "uuid",
    "term_id": "uuid",
    "score": 75,
    "ca_score": 20,
    "exam_score": 55
  }
  ```

### Bulk Upsert Scores
- **URL:** `POST /results/scores/bulk`
- **Auth:** 🔐 Required (`results:create`)
- **Description:** Enter scores for multiple students
- **Request Body:**
  ```json
  {
    "scores": [
      {
        "student_id": "uuid",
        "subject_id": "uuid",
        "session_id": "uuid",
        "term_id": "uuid",
        "score": 75,
        "ca_score": 20,
        "exam_score": 55
      }
    ]
  }
  ```

### Compute Positions
- **URL:** `POST /results/scores/compute-positions`
- **Auth:** 🔐 Required (`results:update`)
- **Description:** Calculate student positions in a class/subject
- **Request Body:**
  ```json
  {
    "sub_level_id": "uuid",
    "session_id": "uuid",
    "term_id": "uuid"
  }
  ```

### Get Student Scores
- **URL:** `GET /students/{studentId}/scores`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get all scores for a specific student
- **Query Parameters:**
  - `session_id` (optional): Filter by session
  - `term_id` (optional): Filter by term
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "student_id": "uuid",
        "subject_id": "uuid",
        "subject_name": "Mathematics",
        "session_id": "uuid",
        "term_id": "uuid",
        "score": 75,
        "ca_score": 20,
        "exam_score": 55,
        "position": 5,
        "grade": "A"
      }
    ]
  }
  ```

---

## Report Cards

### List Report Cards
- **URL:** `GET /results/report-cards`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get all generated report cards
- **Query Parameters:**
  - `page` (optional)
  - `limit` (optional)
  - `session_id` (optional)
  - `term_id` (optional)
  - `sub_level_id` (optional)
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "student_id": "uuid",
        "session_id": "uuid",
        "term_id": "uuid",
        "sub_level_id": "uuid",
        "is_published": false,
        "remarks": "Good performance",
        "generated_at": "2024-01-15T10:00:00Z"
      }
    ]
  }
  ```

### Generate Report Cards
- **URL:** `POST /results/report-cards/generate`
- **Auth:** 🔐 Required (`results:create`)
- **Description:** Generate report cards for a class/term
- **Request Body:**
  ```json
  {
    "sub_level_id": "uuid",
    "session_id": "uuid",
    "term_id": "uuid"
  }
  ```

### Get Report Card by ID
- **URL:** `GET /results/report-cards/{id}`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get full report card details including scores
- **Response:**
  ```json
  {
    "id": "uuid",
    "student_id": "uuid",
    "student_name": "Jane Doe",
    "session_id": "uuid",
    "term_id": "uuid",
    "sub_level_id": "uuid",
    "is_published": false,
    "remarks": "Good performance",
    "generated_at": "2024-01-15T10:00:00Z",
    "scores": [
      {
        "subject_id": "uuid",
        "subject_name": "Mathematics",
        "ca_score": 20,
        "exam_score": 55,
        "total_score": 75,
        "grade": "A",
        "position": 5
      }
    ]
  }
  ```

### Get Student All Report Cards
- **URL:** `GET /students/{studentId}/report-cards`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get all report cards for a student across all sessions/terms
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "session": "2023/2024",
        "term": "First Term",
        "average_score": 72.5,
        "position": 5,
        "is_published": true
      }
    ]
  }
  ```

### Get Student Current Report Card
- **URL:** `GET /students/{studentId}/report-card`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get report card for active session/term

### Update Report Card Remarks
- **URL:** `PUT /results/report-cards/{id}/remarks`
- **Auth:** 🔐 Required (`results:update`)
- **Description:** Add or update remarks on a report card
- **Request Body:**
  ```json
  {
    "remarks": "Excellent performance this term. Keep it up!"
  }
  ```

### Publish Report Card
- **URL:** `POST /results/report-cards/{id}/publish`
- **Auth:** 🔐 Required (`results:publish`)
- **Description:** Mark a report card as published
- **Request Body:** `{}`

---

## Score & Grade Configuration

### Upsert Score Configuration
- **URL:** `POST /results/score-config`
- **Auth:** 🔐 Required (`results:update`)
- **Description:** Configure score settings (max score, pass mark, etc.)
- **Request Body:**
  ```json
  {
    "school_id": "uuid",
    "max_score": 100,
    "pass_mark": 40,
    "ca_weight": 0.3,
    "exam_weight": 0.7
  }
  ```

### Upsert Grade Configuration
- **URL:** `POST /results/grade-config`
- **Auth:** 🔐 Required (`results:update`)
- **Description:** Configure grade ranges (A: 80-100, B: 70-79, etc.)
- **Request Body:**
  ```json
  {
    "school_id": "uuid",
    "grades": [
      {
        "grade": "A",
        "min_score": 80,
        "max_score": 100,
        "description": "Excellent"
      },
      {
        "grade": "B",
        "min_score": 70,
        "max_score": 79,
        "description": "Good"
      },
      {
        "grade": "C",
        "min_score": 60,
        "max_score": 69,
        "description": "Average"
      },
      {
        "grade": "D",
        "min_score": 40,
        "max_score": 59,
        "description": "Pass"
      },
      {
        "grade": "F",
        "min_score": 0,
        "max_score": 39,
        "description": "Fail"
      }
    ]
  }
  ```

### List Grade Configurations
- **URL:** `GET /results/grade-config`
- **Auth:** 🔐 Required (`results:read`)
- **Description:** Get all grade configurations
- **Response:**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "school_id": "uuid",
        "grades": [
          {
            "grade": "A",
            "min_score": 80,
            "max_score": 100,
            "description": "Excellent"
          }
        ]
      }
    ]
  }
  ```

---

## Avatars

### Upload Personnel Avatar
- **URL:** `PUT /avatar/personnel/{id}`
- **Auth:** 🔐 Required (`avatar:update`)
- **Description:** Upload profile picture for staff member
- **Content-Type:** `multipart/form-data`
- **Form Data:**
  - `file`: Image file (JPG, PNG, WebP)
- **Response:**
  ```json
  {
    "id": "uuid",
    "avatar_url": "https://cdn.example.com/avatars/personnel/uuid",
    "uploaded_at": "2024-01-15T10:00:00Z"
  }
  ```

### Upload Student Avatar
- **URL:** `PUT /avatar/students/{id}`
- **Auth:** 🔐 Required (`avatar:update`)
- **Description:** Upload profile picture for student
- **Content-Type:** `multipart/form-data`
- **Form Data:**
  - `file`: Image file (JPG, PNG, WebP)
- **Response:**
  ```json
  {
    "id": "uuid",
    "avatar_url": "https://cdn.example.com/avatars/students/uuid",
    "uploaded_at": "2024-01-15T10:00:00Z"
  }
  ```

---

## Error Responses

All endpoints return standardized error responses:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ]
  }
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| `INVALID_REQUEST` | 400 | Invalid request parameters |
| `UNAUTHORIZED` | 401 | Missing or invalid authentication token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Resource already exists or conflict |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Authentication

### Get Auth Token
- **URL:** `POST /auth/login`
- **Auth:** None
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "password"
  }
  ```
- **Response:**
  ```json
  {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_in": 3600
  }
  ```

### Refresh Token
- **URL:** `POST /auth/refresh`
- **Auth:** None
- **Request Body:**
  ```json
  {
    "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
  }
  ```

### Get Current User
- **URL:** `GET /auth/me`
- **Auth:** 🔐 Required
- **Response:**
  ```json
  {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "roles": ["admin", "teacher"]
  }
  ```

### Change Password
- **URL:** `POST /auth/change-password`
- **Auth:** 🔐 Required
- **Request Body:**
  ```json
  {
    "old_password": "oldpassword",
    "new_password": "newpassword"
  }
  ```

### Logout
- **URL:** `POST /auth/logout`
- **Auth:** 🔐 Required
- **Request Body:** `{}`

---

## Pagination

Endpoints that support pagination use query parameters:

- `page`: Page number (default: 1)
- `limit`: Records per page (default: 10, max: 100)

Response includes:
```json
{
  "data": [...],
  "pagination": {
    "total": 50,
    "page": 1,
    "limit": 10,
    "total_pages": 5
  }
}
```

---

## Rate Limiting

- Rate limit: 100 requests per minute per user
- Headers returned:
  - `X-RateLimit-Limit`: 100
  - `X-RateLimit-Remaining`: 99
  - `X-RateLimit-Reset`: Unix timestamp

---

## SDK Usage Examples

### React Hook Example

```typescript
// useAcademicSessions.ts
import { useState, useEffect } from 'react';

interface Session {
  id: string;
  name: string;
  start_date: string;
  end_date: string;
  is_active: boolean;
}

export function useAcademicSessions() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchSessions = async () => {
      try {
        const token = localStorage.getItem('access_token');
        const response = await fetch('/api/v1/sessions', {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        });
        
        if (!response.ok) throw new Error('Failed to fetch sessions');
        const data = await response.json();
        setSessions(data.data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchSessions();
  }, []);

  return { sessions, loading, error };
}
```

### API Client Example

```typescript
// apiClient.ts
const API_BASE = 'http://localhost:8080/api/v1';

export class ApiClient {
  private token: string;

  constructor(token: string) {
    this.token = token;
  }

  async request(endpoint: string, options: RequestInit = {}) {
    const response = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.statusText}`);
    }

    return response.json();
  }

  // Sessions
  getSessions(page = 1, limit = 10) {
    return this.request(`/sessions?page=${page}&limit=${limit}`);
  }

  getActiveSession() {
    return this.request('/sessions/active');
  }

  createSession(data: any) {
    return this.request('/sessions', { 
      method: 'POST', 
      body: JSON.stringify(data) 
    });
  }

  // Personnel
  getPersonnel(page = 1, limit = 10) {
    return this.request(`/personnel?page=${page}&limit=${limit}`);
  }

  // Results
  getStudentScores(studentId: string) {
    return this.request(`/students/${studentId}/scores`);
  }

  uploadAvatar(type: 'personnel' | 'students', id: string, file: File) {
    const formData = new FormData();
    formData.append('file', file);
    
    return this.request(`/avatar/${type}/${id}`, {
      method: 'PUT',
      headers: {}, // Let browser set Content-Type with boundary
      body: formData,
    });
  }
}
```

---

## Version History

- **v1** (Current)
  - Initial API release
  - Academic Sessions management
  - Personnel management
  - Student enrollment and results
  - Report card generation
  - Score and grade configuration

---

## Support

For issues or questions, please contact the development team at `dev@e-dossier.edu`
