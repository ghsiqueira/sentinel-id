package repositories

import (
	"context"
	"sentinel-id/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	DB *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (r *SessionRepository) CreateSession(session *models.Session) error {
	query := `
		INSERT INTO sessions (user_id, refresh_token, device_name, ip_address, expires_at, is_revoked)
		VALUES ($1, $2, $3, $4, $5, false)
		RETURNING id
	`

	err := r.DB.QueryRow(context.Background(), query,
		session.UserID,
		session.RefreshToken,
		session.DeviceName,
		session.IPAddress,
		session.ExpiresAt,
	).Scan(&session.ID)

	return err
}

func (r *SessionRepository) RevokeSession(refreshToken string) error {
	query := `UPDATE sessions SET is_revoked = true WHERE refresh_token = $1`
	_, err := r.DB.Exec(context.Background(), query, refreshToken)
	return err
}

func (r *SessionRepository) GetSessionByToken(refreshToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, is_revoked, expires_at 
		FROM sessions 
		WHERE refresh_token = $1
	`

	var session models.Session
	err := r.DB.QueryRow(context.Background(), query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.IsRevoked,
		&session.ExpiresAt,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

func (r *SessionRepository) RevokeAllUserSessions(userID string) error {
	query := `UPDATE sessions SET is_revoked = true WHERE user_id = $1`
	_, err := r.DB.Exec(context.Background(), query, userID)
	return err
}
