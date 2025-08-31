# Phone Authentication API Documentation

This document describes the phone authentication endpoints for the ChitChat application.

## Base URL

```
http://localhost:4000
```

## Authentication Flow

1. **Send Verification Code**: User provides phone number to receive a 6-digit verification code
2. **Verify Code**: User enters the received code to verify their phone number
3. **Sign Up/Sign In**: User completes registration or authentication with verified phone number

## Endpoints

### 1. Send Verification Code

**POST** `/auth/phone/send-code`

Sends a 6-digit verification code to the provided phone number.

**Request Body:**

```json
{
  "phone_number": "+1234567890"
}
```

**Response:**

```json
{
  "message": "Verification code sent successfully",
  "phone_number": "+1234567890"
}
```

**Error Response:**

```json
{
  "error": "Invalid phone number format"
}
```

### 2. Verify Code

**POST** `/auth/phone/verify-code`

Verifies the provided code for a phone number.

**Request Body:**

```json
{
  "phone_number": "+1234567890",
  "code": "123456"
}
```

**Response:**

```json
{
  "message": "Code verified successfully",
  "phone_number": "+1234567890"
}
```

**Error Response:**

```json
{
  "error": "Invalid verification code"
}
```

### 3. Phone Sign Up

**POST** `/auth/phone/signup`

Creates a new user account with phone number authentication.

**Request Body:**

```json
{
  "phone_number": "+1234567890",
  "name": "John Doe",
  "code": "123456"
}
```

**Response:**

```json
{
  "message": "User created successfully",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "tenant_id": "uuid",
    "phone_number": "+1234567890",
    "name": "John Doe",
    "role": "user",
    "phone_verified": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Error Response:**

```json
{
  "error": "User with this phone number already exists"
}
```

### 4. Phone Sign In

**POST** `/auth/phone/signin`

Authenticates an existing user with phone number.

**Request Body:**

```json
{
  "phone_number": "+1234567890",
  "code": "123456"
}
```

**Response:**

```json
{
  "message": "Authentication successful",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Error Response:**

```json
{
  "error": "User not found"
}
```

### 5. Resend Verification Code

**POST** `/auth/phone/resend-code`

Resends a verification code to the phone number.

**Request Body:**

```json
{
  "phone_number": "+1234567890"
}
```

**Response:**

```json
{
  "message": "Verification code resent successfully",
  "phone_number": "+1234567890"
}
```

### 6. Update Phone Number (Protected)

**PUT** `/auth/phone/update`

Updates the authenticated user's phone number. Requires JWT token.

**Headers:**

```
Authorization: Bearer <jwt_token>
```

**Request Body:**

```json
{
  "phone_number": "+1987654321"
}
```

**Response:**

```json
{
  "message": "Phone number updated successfully",
  "phone_number": "+1987654321"
}
```

## Error Codes

| Status Code | Description                                      |
| ----------- | ------------------------------------------------ |
| 400         | Bad Request - Invalid request body or parameters |
| 401         | Unauthorized - Invalid or missing authentication |
| 500         | Internal Server Error - Server-side error        |

## Common Error Messages

- `"Invalid phone number format"` - Phone number doesn't meet minimum requirements
- `"Invalid verification code"` - Code is incorrect, expired, or already used
- `"User not found"` - No user exists with the provided phone number
- `"User with this phone number already exists"` - Phone number is already registered
- `"Phone number not verified"` - Phone number hasn't been verified yet

## Development Notes

### SMS Service Integration

The application includes a flexible SMS service interface that supports:

1. **Console SMS Service** (Development): Prints SMS messages to console
2. **Twilio SMS Service** (Production): Sends actual SMS via Twilio

To enable Twilio SMS:

1. Set environment variables:

   ```bash
   export TWILIO_ACCOUNT_SID="your_account_sid"
   export TWILIO_AUTH_TOKEN="your_auth_token"
   export TWILIO_PHONE_NUMBER="your_twilio_number"
   ```

2. Add Twilio Go SDK to go.mod (when ready for production)

### Verification Code Expiration

- Codes expire after 10 minutes
- Codes can only be used once
- Multiple codes can be requested for the same phone number

### Phone Number Format

- Basic validation requires at least 10 characters
- For production, consider using a proper phone number validation library
- International format recommended (e.g., +1234567890)

## Testing

### Using curl

1. Send verification code:

```bash
curl -X POST http://localhost:4000/auth/phone/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'
```

2. Verify code:

```bash
curl -X POST http://localhost:4000/auth/phone/verify-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "code": "123456"}'
```

3. Sign up:

```bash
curl -X POST http://localhost:4000/auth/phone/signup \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "name": "John Doe", "code": "123456"}'
```

4. Sign in:

```bash
curl -X POST http://localhost:4000/auth/phone/signin \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "code": "123456"}'
```
