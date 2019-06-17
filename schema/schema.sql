CREATE EXTENSION pgcrypto; -- imports gen_random_uuid()

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  bio TEXT,
  -- OWASP Top 10 2017 #3: Sensitive Data Exposure
  -- We store a plaintext password in this field.  An attacker could read
  -- plaintext passwords via SQL injection (or another flaw), allowing them to
  -- access our other users' accounts. Due to password reuse, an attacker could
  -- potentially access our users' accounts in different apps as well.
  --
  -- We should be storing salted and hashed passwords, not plaintext passwords.
  password TEXT NOT NULL, 
  clicks BIGINT, -- BIGINT for the power-users 😉
  last_click TIMESTAMPTZ,
  is_admin BOOLEAN
);