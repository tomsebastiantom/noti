package postgres

import (
    "context"
    "encoding/json"
    "getnoti.com/internal/tenants/domain"
    "getnoti.com/internal/tenants/repos"
    "getnoti.com/pkg/db"
    "time"
)

type postgresUserRepository struct {
    db db.Database
}

func NewPostgresUserRepository(db db.Database) repository.UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user domain.User) error {
    now := time.Now()
    consents, err := json.Marshal(user.Consents)
    if err != nil {
        return err
    }

    query := `INSERT INTO users (id, tenant_id, email, phone_number, device_id, web_push_token, consents, preferred_mode, created_at, updated_at) 
              VALUES (\$1, \$2, \$3, \$4, \$5, \$6, \$7, \$8, \$9, \$10)`
    _, err = r.db.Exec(ctx, query, user.ID, user.TenantID, user.Email, user.PhoneNumber, user.DeviceID, user.WebPushToken, consents, user.PreferredMode, now, now)
    return err
}

func (r *postgresUserRepository) GetUserByID(ctx context.Context, userid string) (domain.User, error) {
    query := `SELECT id, tenant_id, email, phone_number, device_id, web_push_token, consents, preferred_mode, created_at, updated_at FROM users WHERE id = \$1`
    row := r.db.QueryRow(ctx, query, userid)
    var user domain.User
    var consents []byte
	err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.PhoneNumber, &user.DeviceID, &user.WebPushToken, &consents, &user.PreferredMode)
    //err := row.Scan(&user.ID, &user.TenantID, &user.Email, &user.PhoneNumber, &user.DeviceID, &user.WebPushToken, &consents, &user.PreferredMode, &user.CreatedAt, &user.UpdatedAt)
    if err != nil {
        return domain.User{}, err
    }

    err = json.Unmarshal(consents, &user.Consents)
    if err != nil {
        return domain.User{}, err
    }

    return user, nil
}

func (r *postgresUserRepository) UpdateUser(ctx context.Context, user domain.User) error {
    now := time.Now()
    consents, err := json.Marshal(user.Consents)
    if err != nil {
        return err
    }

    query := `UPDATE users SET tenant_id = \$1, email = \$2, phone_number = \$3, device_id = \$4, web_push_token = \$5, consents = \$6, preferred_mode = \$7, updated_at = \$8 WHERE id = \$9`
    _, err = r.db.Exec(ctx, query, user.TenantID, user.Email, user.PhoneNumber, user.DeviceID, user.WebPushToken, consents, user.PreferredMode, now, user.ID)
    return err
}

func (r *postgresUserRepository) GetUsersByTenantID(ctx context.Context, tenantid string) ([]domain.User, error) {
    query := `SELECT id, tenant_id, email, phone_number, device_id, web_push_token, consents, preferred_mode, created_at, updated_at FROM users WHERE tenant_id = \$1`
    rows, err := r.db.Query(ctx, query, tenantid)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []domain.User
    for rows.Next() {
        var user domain.User
        var consents []byte
		err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PhoneNumber, &user.DeviceID, &user.WebPushToken, &consents, &user.PreferredMode)
        //err := rows.Scan(&user.ID, &user.TenantID, &user.Email, &user.PhoneNumber, &user.DeviceID, &user.WebPushToken, &consents, &user.PreferredMode, &user.CreatedAt, &user.UpdatedAt)
        if err != nil {
            return nil, err
        }

        err = json.Unmarshal(consents, &user.Consents)
        if err != nil {
            return nil, err
        }

        users = append(users, user)
    }
    return users, nil
}
