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

-- Seed data, we didn't implement a 'Create User' page.
INSERT INTO users 
(username, email, bio, password, clicks, last_click, is_admin)
VALUES
('admin', 'admin@clickonthiscat.com', 'admin user', 'hunter2', 0, '2019-04-01 23:50:44.52908+00', true),
('user0', 'user0@clickonthiscat.com', 'zeroth user', '8M-55a1~Ew4OkwO', 5000, '2019-04-01 23:50:44.52908+00', false),
('user1', 'user1@clickonthiscat.com', 'first user', '%Tlz)5 ?>7O&4&C', 4000, '2019-04-01 23:50:44.52908+00', false),
('user2', 'user2@clickonthiscat.com', 'second user', 'B7</PV{ro298b2l', 300, '2019-04-01 23:50:44.52908+00', false),
('user3', 'user3@clickonthiscat.com', 'third user', '22392(6^8=;b/yC', 200, '2019-04-01 23:50:44.52908+00', false),
('user4', 'user4@clickonthiscat.com', 'fourth user', '&]"<%R|7+-28t8b', 50, '2019-04-01 23:50:44.52908+00', false),
('hacker', 'hacker@clickonthiscat.com', 'admin user', 'hackerpassword', 0, '2019-04-01 23:50:44.52908+00', false);
