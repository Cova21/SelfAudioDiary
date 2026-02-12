package migrations

// CreateEntriesTable migration
var CreateEntriesTable = `
CREATE TABLE IF NOT EXISTS diary_entries (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    audio_file_id VARCHAR(255) NOT NULL,
    duration_seconds INTEGER NOT NULL,
    transcription TEXT,
    segments JSONB,
    language VARCHAR(50),
    sentiment VARCHAR(50),
    emotions TEXT[],
    keywords TEXT[],
    tags TEXT[],
    summary TEXT,
    status INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_entries_user_id ON diary_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_entries_created_at ON diary_entries(created_at);
CREATE INDEX IF NOT EXISTS idx_entries_status ON diary_entries(status);
`

// GetMigrations returns all migrations
func GetMigrations() []string {
	return []string{
		CreateEntriesTable,
	}
}
