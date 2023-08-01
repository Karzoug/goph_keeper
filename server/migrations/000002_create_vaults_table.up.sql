CREATE TABLE IF NOT EXISTS vaults (
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL REFERENCES users (email),
	name TEXT NOT NULL,
	type INTEGER,
	value BLOB,
	updated_at DATETIME);
CREATE INDEX vaults_email_index ON vaults (email);