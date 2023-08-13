CREATE TABLE IF NOT EXISTS users (
	email TEXT PRIMARY KEY,
	is_email_verified INTEGER,
	auth_key BLOB,
	created_at DATETIME);