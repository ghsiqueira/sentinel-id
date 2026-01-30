package repositories

import (
	"context"
	"sentinel-id/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	DB *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (r *SessionRepository) Create(session models.Session) error {
	query := `INSERT INTO sessions (id, user_id, token, refresh_token, device_info, ip_address, expires_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.Exec(context.Background(), query, session.ID, session.UserID, session.Token, session.RefreshToken, session.DeviceInfo, session.IPAddress, session.ExpiresAt)
	return err
}

func (r *SessionRepository) Revoke(refreshToken string) error {
	query := `DELETE FROM sessions WHERE refresh_token = $1`
	_, err := r.DB.Exec(context.Background(), query, refreshToken)
	return err
}

func (r *SessionRepository) RevokeAll(userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.DB.Exec(context.Background(), query, userID)
	return err
}

func (r *SessionRepository) ListByUser(userID uuid.UUID) ([]models.Session, error) {
	query := `SELECT id, user_id, token, refresh_token, device_info, ip_address, created_at, expires_at FROM sessions WHERE user_id = $1`
	rows, err := r.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		if err := rows.Scan(&s.ID, &s.UserID, &s.Token, &s.RefreshToken, &s.DeviceInfo, &s.IPAddress, &s.CreatedAt, &s.ExpiresAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
