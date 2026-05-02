package repository

import (
	"context"
	"time"
	"todo_api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateRefreshToken persists a new refresh token for the given user.
func CreateRefreshToken(pool *pgxpool.Pool, rt *models.RefreshToken) (*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, created_at, revoked
	`

	err := pool.QueryRow(ctx, query, rt.UserID, rt.Token, rt.ExpiresAt).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.Revoked,
	)
	if err != nil {
		return nil, err
	}

	return rt, nil
}

// GetRefreshToken looks up a token string. Returns the record even if revoked
// so callers can distinguish "not found" from "revoked".
func GetRefreshToken(pool *pgxpool.Pool, token string) (*models.RefreshToken, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked
		FROM refresh_tokens
		WHERE token = $1
	`

	var rt models.RefreshToken
	err := pool.QueryRow(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&rt.Revoked,
	)
	if err != nil {
		return nil, err
	}

	return &rt, nil
}

// RevokeRefreshToken marks a single token as revoked.
func RevokeRefreshToken(pool *pgxpool.Pool, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE refresh_tokens SET revoked = true WHERE token = $1`
	_, err := pool.Exec(ctx, query, token)
	return err
}

// RevokeAllUserTokens revokes every token belonging to a user (full logout).
func RevokeAllUserTokens(pool *pgxpool.Pool, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := pool.Exec(ctx, query, userID)
	return err
}

// DeleteExpiredTokens is a housekeeping helper — call it from a cron or
// startup routine to prevent the table from growing indefinitely.
func DeleteExpiredTokens(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = true`
	_, err := pool.Exec(ctx, query)
	return err
}
