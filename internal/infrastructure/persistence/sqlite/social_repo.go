package sqlite

import (
	"context"
	"database/sql"
	"fmt"
)

type SocialRepository struct {
	db *sql.DB
}

func NewSocialRepository(db *sql.DB) *SocialRepository {
	return &SocialRepository{db: db}
}

func (r *SocialRepository) ClearTables(ctx context.Context) error {
	tables := []string{"social_posts", "social_likes", "social_connections", "social_comments", "social_messages"}
	for _, table := range tables {
		if _, err := r.db.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}
	return nil
}

func (r *SocialRepository) InsertPost(ctx context.Context, platform, postType, caption, location, timestamp, path, mediaURL, metadata string) error {
	query := `INSERT INTO social_posts (platform, post_type, caption, location, timestamp, path, media_url, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, platform, postType, caption, location, timestamp, path, mediaURL, metadata)
	return err
}

func (r *SocialRepository) InsertLike(ctx context.Context, platform, timestamp, targetUsername, targetType, reaction string) error {
	query := `INSERT INTO social_likes (platform, timestamp, target_username, target_type, reaction) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, platform, timestamp, targetUsername, targetType, reaction)
	return err
}

func (r *SocialRepository) InsertConnection(ctx context.Context, platform, connectionType, username, timestamp string) error {
	query := `INSERT INTO social_connections (platform, connection_type, username, timestamp) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, platform, connectionType, username, timestamp)
	return err
}

func (r *SocialRepository) InsertComment(ctx context.Context, platform, timestamp, username, text string) error {
	query := `INSERT INTO social_comments (platform, timestamp, username, text) VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, platform, timestamp, username, text)
	return err
}

func (r *SocialRepository) InsertMessage(ctx context.Context, platform, timestamp, sender, receiver, text, mediaURL, storyShare, metadata string) error {
	query := `INSERT INTO social_messages (platform, timestamp, sender, receiver, text, media_url, story_share, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, platform, timestamp, sender, receiver, text, mediaURL, storyShare, metadata)
	return err
}
