CREATE TABLE IF NOT EXISTS vaults (
	email TEXT NOT NULL,
	name TEXT NOT NULL,
	value BLOB,
	updated_at DATETIME,
	PRIMARY KEY(email, name));