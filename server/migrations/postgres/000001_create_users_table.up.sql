CREATE TABLE IF NOT EXISTS users (
	email TEXT PRIMARY KEY,
	is_email_verified  BOOLEAN,
	auth_key bytea,
	created_at timestamp);