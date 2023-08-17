CREATE TABLE IF NOT EXISTS vaults (
    id TEXT PRIMARY KEY NOT NULL,
	name TEXT,
	type INTEGER NOT NULL,
	value BLOB,
	server_updated_at INTEGER,
	client_updated_at INTEGER NOT NULL);