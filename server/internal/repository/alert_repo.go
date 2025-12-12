package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/fiveseconds/server/internal/model"
	"github.com/jackc/pgx/v5"
)

// AlertRepo 告警仓库
type AlertRepo struct{}

// NewAlertRepo 创建告警仓库
func NewAlertRepo() *AlertRepo {
	return &AlertRepo{}
}

// Create 创建告警
func (r *AlertRepo) Create(ctx context.Context, alert *model.Alert) error {
	sql := `INSERT INTO alerts (alert_type, severity, title, details, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return DB.QueryRow(ctx, sql,
		alert.AlertType, alert.Severity, alert.Title, alert.Details, alert.Status,
	).Scan(&alert.ID, &alert.CreatedAt)
}

// GetByID 根据ID获取告警
func (r *AlertRepo) GetByID(ctx context.Context, id int64) (*model.Alert, error) {
	sql := `SELECT id, alert_type, severity, title, details, status, acknowledged_by, acknowledged_at, created_at
		FROM alerts WHERE id = $1`
	alert := &model.Alert{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&alert.ID, &alert.AlertType, &alert.Severity, &alert.Title, &alert.Details,
		&alert.Status, &alert.AcknowledgedBy, &alert.AcknowledgedAt, &alert.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return alert, err
}

// List 列表告警
func (r *AlertRepo) List(ctx context.Context, query *model.AlertListQuery) ([]*model.Alert, int64, error) {
	countSQL := `SELECT COUNT(*) FROM alerts WHERE 1=1`
	listSQL := `SELECT id, alert_type, severity, title, details, status, acknowledged_by, acknowledged_at, created_at
		FROM alerts WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.AlertType != nil {
		countSQL += fmt.Sprintf(` AND alert_type = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND alert_type = $%d`, argIdx)
		args = append(args, *query.AlertType)
		argIdx++
	}
	if query.Severity != nil {
		countSQL += fmt.Sprintf(` AND severity = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND severity = $%d`, argIdx)
		args = append(args, *query.Severity)
		argIdx++
	}
	if query.Status != nil {
		countSQL += fmt.Sprintf(` AND status = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, *query.Status)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []*model.Alert
	for rows.Next() {
		alert := &model.Alert{}
		if err := rows.Scan(
			&alert.ID, &alert.AlertType, &alert.Severity, &alert.Title, &alert.Details,
			&alert.Status, &alert.AcknowledgedBy, &alert.AcknowledgedAt, &alert.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, total, nil
}

// Acknowledge 确认告警
func (r *AlertRepo) Acknowledge(ctx context.Context, id int64, acknowledgedBy int64) error {
	sql := `UPDATE alerts SET status = 'acknowledged', acknowledged_by = $1, acknowledged_at = NOW()
		WHERE id = $2 AND status = 'active'`
	tag, err := DB.Exec(ctx, sql, acknowledgedBy, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("alert not found or already acknowledged")
	}
	return nil
}

// GetActiveCount 获取活跃告警数量
func (r *AlertRepo) GetActiveCount(ctx context.Context) (int64, error) {
	sql := `SELECT COUNT(*) FROM alerts WHERE status = 'active'`
	var count int64
	err := DB.QueryRow(ctx, sql).Scan(&count)
	return count, err
}

// GetActiveBySeverity 按严重程度获取活跃告警数量
func (r *AlertRepo) GetActiveBySeverity(ctx context.Context) (map[model.AlertSeverity]int64, error) {
	sql := `SELECT severity, COUNT(*) FROM alerts WHERE status = 'active' GROUP BY severity`
	rows, err := DB.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[model.AlertSeverity]int64)
	for rows.Next() {
		var severity model.AlertSeverity
		var count int64
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, err
		}
		result[severity] = count
	}
	return result, nil
}
