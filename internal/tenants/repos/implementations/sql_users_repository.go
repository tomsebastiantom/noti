package repos

import (
	"context"
	"fmt"
	"time"

	"getnoti.com/internal/tenants/domain"
	repository "getnoti.com/internal/tenants/repos"
	"getnoti.com/pkg/db"
)

type sqlUserRepository struct {
	db db.Database
}

func NewUserRepository(db db.Database) repository.UserRepository {
	return &sqlUserRepository{db: db}
}

func (r *sqlUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	now := time.Now()
	query := `INSERT INTO users (id, email, phone_number, device_id, created_at, updated_at) 
              VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.PhoneNumber, user.DeviceID, now, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *sqlUserRepository) GetUserByID(ctx context.Context, userid string) (domain.User, error) {
	query := `SELECT id, email, phone_number, device_id FROM users WHERE id = ?`
	row := r.db.QueryRow(ctx, query, userid)
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.DeviceID)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *sqlUserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	now := time.Now()
	query := `UPDATE users SET email = ?, phone_number = ?, device_id = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(ctx, query, user.Email, user.PhoneNumber, user.DeviceID, now, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *sqlUserRepository) GetUsers(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, phone_number, device_id FROM users`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}
	return users, nil
}
