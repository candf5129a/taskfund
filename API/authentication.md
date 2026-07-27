# API Structure
## Authentication API
### Register
```
POST /api/v1/auth/register
```

### Body
```
{
  "first_name": "",
  "last_name": "",
  "email": "",
  "password": ""
}
```

### Response
```
{
  "success": true,
  "message": "Registration successful."
}
```
### Login
```
POST /api/v1/auth/login
```
### Logout
```
POST /api/v1/auth/logout
```
### Verify Email
```
POST /api/v1/auth/verify-email
```
## User API
```
GET /api/v1/profile

PUT /api/v1/profile

GET /api/v1/profile/reviews

GET /api/v1/profile/ratings
```
## Task API
```
GET /api/v1/tasks

GET /api/v1/tasks/{id}

POST /api/v1/tasks/{id}/accept

POST /api/v1/tasks/{id}/submit

GET /api/v1/tasks/my

GET /api/v1/tasks/completed
```
## Wallet API
```
GET /api/v1/wallet

POST /api/v1/wallet/withdraw

GET /api/v1/wallet/history
```
## Campaign API
```
POST /api/v1/campaigns

GET /api/v1/campaigns

PUT /api/v1/campaigns/{id}

DELETE /api/v1/campaigns/{id}
```
## Notification API
```
GET /api/v1/notifications

PATCH /api/v1/notifications/{id}/read
```
## Admin API
```
GET /api/v1/admin/users

GET /api/v1/admin/reports

PATCH /api/v1/admin/users/{id}/ban

PATCH /api/v1/admin/withdrawals/{id}/approve
```
## API Standards

Document these rules:

All responses use JSON.
- Every endpoint returns a consistent structure.
- Version all APIs (/api/v1/...).
- Use JWT authentication for protected endpoints.
- Return proper HTTP status codes.
- Support pagination for large lists.

Example:
```
{
  "success": true,
  "message": "Tasks retrieved successfully.",
  "data": [],
  "pagination": {}
}
```