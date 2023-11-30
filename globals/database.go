package globals

import "database/sql"

var Db *sql.DB

func SetupDb(db *sql.DB) error {
	if db == nil {
		return nil
	}
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id BIGINT NOT NULL,
	avatar VARCHAR(128) DEFAULT NULL,
	password VARCHAR(64),
	token BIGINT NOT NULL,
	username VARCHAR(32) NOT NULL,
	flags BIGINT NOT NULL DEFAULT 0,
	bio VARCHAR(500) DEFAULT NULL,
	PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS posts (
	id BIGINT NOT NULL,
	author BIGINT NOT NULL,
	tags INTEGER NOT NULL DEFAULT 0,
	title VARCHAR(100) NOT NULL,
	edited_timestamp BIGINT DEFAULT NULL,
	body VARCHAR(32000) NOT NULL,
	PRIMARY KEY(id)
);
	
CREATE TABLE IF NOT EXISTS comments (
	id BIGINT NOT NULL,
	post_id BIGINT NOT NULL,
	author_id BIGINT NOT NULL,
	body VARCHAR(6000) NOT NULL,
	PRIMARY KEY(id)
);
		
CREATE TABLE IF NOT EXISTS ratings (
	resource_id BIGINT NOT NULL,
	author_id BIGINT NOT NULL,
	unit INTEGER NOT NULL,
	PRIMARY KEY(resource_id)
);`
	_, err := db.Exec(schema)
	return err
}
