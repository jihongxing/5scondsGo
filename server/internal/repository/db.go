package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DB 全局数据库连接池
var DB *pgxpool.Pool

// InitDB 初始化数据库连接
func InitDB(dsn string, logger *zap.Logger) error {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("parse db config: %w", err)
	}

	config.MaxConns = 50
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	DB = pool
	logger.Info("Database connected", zap.String("host", config.ConnConfig.Host))
	return nil
}

// CloseDB 关闭数据库连接
func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

// Tx 事务辅助函数
func Tx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// TxExecutor 事务执行器接口，用于支持事务和非事务操作
type TxExecutor interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

// GetExecutor 获取执行器，如果tx不为nil则使用事务，否则使用连接池
func GetExecutor(tx pgx.Tx) TxExecutor {
	if tx != nil {
		return tx
	}
	return DB
}
