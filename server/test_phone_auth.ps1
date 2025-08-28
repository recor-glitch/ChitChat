# Test script for Phone Authentication API
# Make sure the server is running on http://localhost:4000

$baseUrl = "http://localhost:4000"
$phoneNumber = "+1234567890"

Write-Host "Testing Phone Authentication API..." -ForegroundColor Green

# Test 1: Send verification code
Write-Host "`n1. Sending verification code..." -ForegroundColor Yellow
$sendCodeResponse = Invoke-RestMethod -Uri "$baseUrl/auth/phone/send-code" -Method POST -ContentType "application/json" -Body '{"phone_number": "' + $phoneNumber + '"}'
Write-Host "Response: $($sendCodeResponse | ConvertTo-Json)" -ForegroundColor Cyan

# Note: In development, the verification code is printed to the console
# You'll need to check the server console output to get the actual code
Write-Host "`nCheck the server console for the verification code!" -ForegroundColor Red
$code = Read-Host "Enter the verification code from console"

# Test 2: Verify code
Write-Host "`n2. Verifying code..." -ForegroundColor Yellow
$verifyCodeResponse = Invoke-RestMethod -Uri "$baseUrl/auth/phone/verify-code" -Method POST -ContentType "application/json" -Body '{"phone_number": "' + $phoneNumber + '", "code": "' + $code + '"}'
Write-Host "Response: $($verifyCodeResponse | ConvertTo-Json)" -ForegroundColor Cyan

# Test 3: Phone signup
Write-Host "`n3. Testing phone signup..." -ForegroundColor Yellow
$signupResponse = Invoke-RestMethod -Uri "$baseUrl/auth/phone/signup" -Method POST -ContentType "application/json" -Body '{"phone_number": "' + $phoneNumber + '", "name": "Test User", "code": "' + $code + '"}'
Write-Host "Response: $($signupResponse | ConvertTo-Json)" -ForegroundColor Cyan

# Test 4: Phone signin (after signup)
Write-Host "`n4. Testing phone signin..." -ForegroundColor Yellow
$signinResponse = Invoke-RestMethod -Uri "$baseUrl/auth/phone/signin" -Method POST -ContentType "application/json" -Body '{"phone_number": "' + $phoneNumber + '", "code": "' + $code + '"}'
Write-Host "Response: $($signinResponse | ConvertTo-Json)" -ForegroundColor Cyan

Write-Host "`nPhone authentication tests completed!" -ForegroundColor Green 