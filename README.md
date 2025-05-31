# Auth Service – GridWhiz Candidate Assignment

This is a gRPC-based authentication microservice built with Go and MongoDB.  
It supports user registration, JWT-based login, profile management, password reset, and role-based access control.

---

## 📌 Features

- ✅ Register / Login with email & password (bcrypt hashed)
- ✅ JWT token generation and validation
- ✅ Role-based access control (`admin`, `user`)
- ✅ User profile management (view, update, delete)
- ✅ Rate limiting for login attempts
- ✅ Password reset flow with token validation
- ✅ Soft delete via `is_deleted` flag

---

## 🚀 Tech Stack

- Go (Golang)
- gRPC
- MongoDB (NoSQL)
- Protocol Buffers (`.proto`)
- JWT (JSON Web Token)
- bcrypt for secure password hashing

---

## 🛠️ Setup Instructions

### 1. Clone Repository

```bash
git clone https://github.com/bekbek22/auth_service.git
cd auth_service
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Start MongoDB

Ensure MongoDB is running locally at:

```
mongodb://localhost:27017
```

Or update the config in `config/config.go` if needed.

### 4. Generate gRPC Code

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/auth.proto
```

### 5. Run Server

```bash
go run cmd/main.go
```

---

## 🧪 Testing with Postman or grpcurl

Use **Postman v10+ (gRPC tab)** or `grpcurl` to test.

1. Call `Login` to receive a JWT token
2. For protected routes, add metadata:

```
authorization: Bearer <access_token>
```

---

## 📄 API Documentation

### 🔐 Register

```proto
rpc Register (RegisterRequest) returns (RegisterResponse);
```

**Request**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "secure123"
}
```

**Response**
```json
{ "message": "Registration successful" }
```

---

### 🔐 Login

```proto
rpc Login (LoginRequest) returns (LoginResponse);
```

**Request**
```json
{
  "email": "john@example.com",
  "password": "secure123"
}
```

**Response**
```json
{ "access_token": "<JWT token>" }
```

---

### 🔓 Logout

```proto
rpc Logout (LogoutRequest) returns (LogoutResponse);
```

**Request**
```json
{ "access_token": "<JWT token>" }
```

**Response**
```json
{ "message": "Logged out successfully" }
```

---

### 📩 RequestPasswordReset

```proto
rpc RequestPasswordReset(RequestPasswordResetRequest) returns (RequestPasswordResetResponse);
```

**Request**
```json
{ "email": "john@example.com" }
```

**Response**
```json
{ "reset_token": "uuid-token-here" }
```

---

### 🔁 ResetPassword

```proto
rpc ResetPassword(ResetPasswordRequest) returns (ResetPasswordResponse);
```

**Request**
```json
{
  "reset_token": "uuid-token-here",
  "new_password": "newSecurePassword123"
}
```

**Response**
```json
{ "message": "Password reset successfully" }
```

---

## 👤 User Management

### 📋 ListUsers (admin only)

```proto
rpc ListUsers (ListUsersRequest) returns (ListUsersResponse);
```

**Metadata**
```
authorization: Bearer <admin_token>
```

**Request**
```json
{
  "name": "",
  "email": "",
  "page": 1,
  "limit": 10
}
```

**Response**
```json
{
  "users": [
    {
      "id": "64f...",
      "name": "John Doe",
      "email": "john@example.com",
      "role": "user"
    }
  ],
  "total": 1
}
```

---

### ✅ GetProfile

```proto
rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
```

**Metadata**
```
authorization: Bearer <user_token>
```

**Request**
```json
{ "user_id": "64f..." }
```

**Response**
```json
{
  "id": "64f...",
  "name": "John Doe",
  "email": "john@example.com",
  "role": "user"
}
```

---

### ✏️ UpdateProfile

```proto
rpc UpdateProfile(UpdateProfileRequest) returns (UpdateProfileResponse);
```

**Metadata**
```
authorization: Bearer <user_token>
```

**Request**
```json
{
  "name": "John Updated",
  "email": "updated@example.com"
}
```

**Response**
```json
{ "message": "Profile updated successfully" }
```

---

### ❌ DeleteProfile

```proto
rpc DeleteProfile(DeleteProfileRequest) returns (DeleteProfileResponse);
```

Soft delete: sets `is_deleted: true`.

**Metadata**
```
authorization: Bearer <user_token>
```

**Request**
```json
{}
```

**Response**
```json
{ "message": "Profile deleted (soft delete)" }
```

---

## 🏗️ Architecture Overview

This project follows Clean Architecture principles:

| Layer       | Description                |
|-------------|----------------------------|
| `handler`   | gRPC service handlers      |
| `service`   | Business logic layer       |
| `repository`| MongoDB interaction layer  |

Other design decisions:

- JWT-based authentication (`user_id`, `role` in claims)
- Passwords securely hashed with bcrypt
- In-memory rate limiting (login attempts)
- Password reset via token with expiry (15 min)
- Soft deletion via `is_deleted: true`

---

## 📜 License

For GridWhiz assessment purposes only.
