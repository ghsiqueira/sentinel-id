package repositories

import (
	"context"
	"sentinel-id/internal/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type QRRepository struct {
	DB *pgxpool.Pool
}

func NewQRRepository(db *pgxpool.Pool) *QRRepository {
	return &QRRepository{DB: db}
}

func (r *QRRepository) Create(req models.LoginRequest) error {
	query := `INSERT INTO login_requests (id, status, expires_at, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.DB.Exec(context.Background(), query, req.ID, req.Status, req.ExpiresAt, req.CreatedAt)
	return err
}

func (r *QRRepository) FindByID(id uuid.UUID) (*models.LoginRequest, error) {
	query := `SELECT id, status, user_id, expires_at, created_at FROM login_requests WHERE id = $1`

	var req models.LoginRequest
	err := r.DB.QueryRow(context.Background(), query, id).Scan(&req.ID, &req.Status, &req.UserID, &req.ExpiresAt, &req.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *QRRepository) Approve(id uuid.UUID, userID uuid.UUID) error {
	query := `UPDATE login_requests SET status = 'APPROVED', user_id = $1 WHERE id = $2 AND status = 'PENDING'`
	tag, err := r.DB.Exec(context.Background(), query, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return context.DeadlineExceeded
	}
	return nil
}

func (r *QRRepository) MarkAsUsed(id uuid.UUID) error {
	query := `UPDATE login_requests SET status = 'USED' WHERE id = $1`
	_, err := r.DB.Exec(context.Background(), query, id)
	return err
}

func (r *QRRepository) CreatePromptRequest(reqID string, userID string, deviceInfo string, ip string, expiresAt time.Time) error {
	query := `INSERT INTO login_requests (id, status, user_id, device_info, ip_address, expires_at) VALUES ($1, 'PENDING', $2, $3, $4, $5)`
	_, err := r.DB.Exec(context.Background(), query, reqID, userID, deviceInfo, ip, expiresAt)
	return err
}

func (r *QRRepository) GetPendingPrompt(userID string) (string, string, string, error) {
	var reqID, devInfo, ipAddress string
	query := `SELECT id, COALESCE(device_info, 'Desconhecido'), COALESCE(ip_address, 'IP Oculto') FROM login_requests WHERE user_id = $1 AND status = 'PENDING' AND expires_at > NOW() LIMIT 1`
	err := r.DB.QueryRow(context.Background(), query, userID).Scan(&reqID, &devInfo, &ipAddress)
	return reqID, devInfo, ipAddress, err
}
