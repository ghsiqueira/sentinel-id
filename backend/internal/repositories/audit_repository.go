package repositories

import (
	"context"
	"sentinel-id/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	DB *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{DB: db}
}

func (r *AuditRepository) Create(log models.AuditLog) error {
	query := `INSERT INTO audit_logs (id, user_id, action, ip_address, user_agent, details, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.DB.Exec(context.Background(), query, log.ID, log.UserID, log.Action, log.IPAddress, log.UserAgent, log.Details, log.CreatedAt)
	return err
}

func (r *AuditRepository) ListByUser(userID string) ([]models.AuditLog, error) {
	query := `SELECT id, user_id, action, ip_address, user_agent, details, created_at FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50`

	rows, err := r.DB.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.IPAddress, &l.UserAgent, &l.Details, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}
