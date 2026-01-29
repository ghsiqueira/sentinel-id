package repositories

import (
	"context"
	"errors"
	"sentinel-id/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	DB *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (full_name, email, cpf, password_hash, mfa_enabled, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.DB.QueryRow(context.Background(), query,
		user.FullName,
		user.Email,
		user.CPF,
		user.PasswordHash,
		user.MfaEnabled,
		user.CreatedAt,
	).Scan(&user.ID)

	return err
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `SELECT id, full_name, email, cpf, password_hash, mfa_enabled FROM users WHERE email = $1`

	var user models.User
	err := r.DB.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.CPF,
		&user.PasswordHash,
		&user.MfaEnabled,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func (r *UserRepository) GetByCPF(cpf string) (*models.User, error) {
	query := `SELECT id, full_name, email, cpf FROM users WHERE cpf = $1`

	var user models.User
	err := r.DB.QueryRow(context.Background(), query, cpf).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.CPF,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	query := `SELECT id, full_name, email, cpf, password_hash FROM users WHERE email = $1`

	err := r.DB.QueryRow(context.Background(), query, email).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.CPF,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
