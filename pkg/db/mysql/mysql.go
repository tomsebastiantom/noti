package mysql

// import (
// 	"context"
// 	"database/sql"
// 	"fmt"

// 	"getnoti.com/config"
// 	_ "github.com/go-sql-driver/mysql"
// )

// type MySQLDB struct {
// 	DB *sql.DB
// }

// func New(cfg *config.Config) (*MySQLDB, error) {
// 	db, err := sql.Open("mysql", cfg.Database.MySQL.DSN)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &MySQLDB{DB: db}, nil
// }

// func (db *MySQLDB) Close() error {
// 	return db.DB.Close()
// }

// func (db *MySQLDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
// 	return db.DB.QueryContext(ctx, query, args...)
// }

// func (db *MySQLDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
// 	return db.DB.ExecContext(ctx, query, args...)
// }

