# Phone Authentication for ChitChat

This document explains how to set up and use phone authentication in your ChitChat application.

## Overview

Phone authentication allows users to sign up and sign in using their phone number instead of email/password. The system uses SMS verification codes to authenticate users.

## Features

- ✅ SMS verification code generation and validation
- ✅ Phone-based user registration and authentication
- ✅ JWT token generation for authenticated sessions
- ✅ Phone number update functionality
- ✅ Flexible SMS service integration (console/Twilio)
- ✅ Verification code expiration (10 minutes)
- ✅ One-time use verification codes

## Database Changes

The phone authentication feature adds the following database changes:

### New Migration: `000003_add_phone_auth.up.sql`

1. **Users table updates:**

   - `phone_number` (TEXT, UNIQUE) - User's phone number
   - `phone_verified` (BOOLEAN) - Whether phone is verified

2. **New table: `phone_verification_codes`**
   - Stores verification codes with expiration times
   - Tracks used/unused codes
   - Indexed for performance

## Setup Instructions

### 1. Run Database Migrations

```bash
# Apply the phone authentication migration
# This will be done automatically when you start the server
```

### 2. Environment Variables (Optional)

For production SMS sending, set these environment variables:

```bash
export TWILIO_ACCOUNT_SID="your_twilio_account_sid"
export TWILIO_AUTH_TOKEN="your_twilio_auth_token"
export TWILIO_PHONE_NUMBER="your_twilio_phone_number"
```

### 3. Start the Server

```bash
cd server
go run cmd/server/main.go
```

## API Endpoints

### Public Endpoints

| Method | Endpoint                  | Description              |
| ------ | ------------------------- | ------------------------ |
| POST   | `/auth/phone/send-code`   | Send verification code   |
| POST   | `/auth/phone/verify-code` | Verify code              |
| POST   | `/auth/phone/signup`      | Register new user        |
| POST   | `/auth/phone/signin`      | Authenticate user        |
| POST   | `/auth/phone/resend-code` | Resend verification code |

### Protected Endpoints

| Method | Endpoint             | Description         |
| ------ | -------------------- | ------------------- |
| PUT    | `/auth/phone/update` | Update phone number |

## Usage Examples

### 1. User Registration Flow

```bash
# Step 1: Send verification code
curl -X POST http://localhost:4000/auth/phone/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'

# Step 2: Check server console for verification code
# Step 3: Verify code
curl -X POST http://localhost:4000/auth/phone/verify-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "code": "123456"}'

# Step 4: Register user
curl -X POST http://localhost:4000/auth/phone/signup \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "name": "John Doe", "code": "123456"}'
```

### 2. User Authentication Flow

```bash
# Step 1: Send verification code
curl -X POST http://localhost:4000/auth/phone/send-code \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890"}'

# Step 2: Sign in with code
curl -X POST http://localhost:4000/auth/phone/signin \
  -H "Content-Type: application/json" \
  -d '{"phone_number": "+1234567890", "code": "123456"}'
```

### 3. Update Phone Number

```bash
# Requires JWT token
curl -X PUT http://localhost:4000/auth/phone/update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <jwt_token>" \
  -d '{"phone_number": "+1987654321"}'
```

## Testing

### Using the Test Script

1. Start the server:

   ```bash
   go run cmd/server/main.go
   ```

2. Run the test script:

   ```bash
   ./test_phone_auth.ps1
   ```

3. Follow the prompts to enter verification codes from the server console.

### Manual Testing

1. Send a verification code
2. Check server console for the code
3. Use the code to verify, signup, or signin
4. Test protected endpoints with the returned JWT token

## SMS Service Integration

### Development Mode (Default)

In development, SMS messages are printed to the console:

```
SMS to +1234567890: Your ChitChat verification code is: 123456
```

### Production Mode (Twilio)

To enable actual SMS sending:

1. **Add Twilio SDK to go.mod:**

   ```go
   require (
       github.com/twilio/twilio-go v1.x.x
   )
   ```

2. **Set environment variables:**

   ```bash
   export TWILIO_ACCOUNT_SID="your_account_sid"
   export TWILIO_AUTH_TOKEN="your_auth_token"
   export TWILIO_PHONE_NUMBER="your_twilio_number"
   ```

3. **Uncomment Twilio implementation in `sms_service.go`**

## Security Considerations

1. **Rate Limiting**: Consider implementing rate limiting for SMS sending
2. **Phone Number Validation**: Add proper phone number format validation
3. **Code Expiration**: Codes expire after 10 minutes
4. **One-time Use**: Codes can only be used once
5. **JWT Security**: Use strong JWT secrets in production

## Troubleshooting

### Common Issues

1. **"User with this phone number already exists"**

   - The phone number is already registered
   - Use signin instead of signup

2. **"Invalid verification code"**

   - Code is expired (10 minutes)
   - Code was already used
   - Code is incorrect

3. **"User not found"**

   - Phone number is not registered
   - Use signup instead of signin

4. **SMS not sending**
   - Check Twilio credentials (if using Twilio)
   - Check server console for development mode

### Debug Mode

Enable debug logging by checking server console output for:

- SMS messages (development mode)
- Database queries
- Error messages

## Future Enhancements

1. **Rate Limiting**: Implement SMS sending rate limits
2. **Phone Validation**: Add proper phone number format validation
3. **Multiple SMS Providers**: Support AWS SNS, Vonage, etc.
4. **Voice Calls**: Add voice call verification option
5. **Two-Factor Authentication**: Combine with email/password auth
