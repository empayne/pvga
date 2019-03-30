CREATE EXTENSION pgcrypto; -- imports gen_random_uuid()

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT UNIQUE NOT NULL,
  email TEXT UNIQUE NOT NULL,
  bio TEXT,
  -- The 'password' column is going to store a plaintext password.
  -- THIS IS A BAD IDEA. DO NOT USE THIS FOR ANYTHING BUT EDUCATIONAL PURPOSES.
  password TEXT NOT NULL, 
  clicks BIGINT, -- BIGINT for the power-users 😉
  last_click TIMESTAMPTZ,
  is_admin BOOLEAN
);