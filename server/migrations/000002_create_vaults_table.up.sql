CREATE TABLE IF NOT EXISTS vaults (
	id TEXT NOT NULL,
	email TEXT NOT NULL REFERENCES users (email),
	name TEXT NOT NULL,
	type INTEGER,
	value BLOB,
	updated_at DATETIME,
	PRIMARY KEY(id,email));