-- Drop phone verification codes table
DROP TABLE IF EXISTS phone_verification_codes;

-- Remove phone fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS phone_verified;
ALTER TABLE users DROP COLUMN IF EXISTS phone_number; 