package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// AlertBroadcaster 告警广播接口
type AlertBroadcaster interface {
	BroadcastToAdmins(msg *model.WSMessage)
}

// AlertManager 告警管理器
type AlertManager struct {
	alertRepo   *repository.AlertRepo
	broadcaster AlertBroadcaster
	logger      *zap.Logger
}

// NewAlertManager 创建告警管理器
func NewAlertManager(
	alertRepo *repository.AlertRepo,
	broadcaster AlertBroadcaster,
	logger *zap.Logger,
) *AlertManager {
	return &AlertManager{
		alertRepo:   alertRepo,
		broadcaster: broadcaster,
		logger:      logger.With(zap.String("service", "alert_manager")),
	}
}

// createAlert 创建告警
func (m *AlertManager) createAlert(ctx context.Context, alertType model.AlertType, severity model.AlertSeverity, title string, details *model.AlertDetails) (*model.Alert, error) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	alert := &model.Alert{
		AlertType: alertType,
		Severity:  severity,
		Title:     title,
		Details:   string(detailsJSON),
		Status:    model.AlertStatusActive,
	}

	if err := m.alertRepo.Create(ctx, alert); err != nil {
		m.logger.Error("Failed to create alert", zap.Error(err))
		return nil, err
	}

	// 广播给管理员
	m.broadcastAlert(alert)

	m.logger.Info("Alert created",
		zap.Int64("alert_id", alert.ID),
		zap.String("type", string(alertType)),
		zap.String("severity", string(severity)))

	return alert, nil
}

// broadcastAlert 广播告警给管理员
func (m *AlertManager) broadcastAlert(alert *model.Alert) {
	if m.broadcaster == nil {
		return
	}

	m.broadcaster.BroadcastToAdmins(&model.WSMessage{
		Type: model.WSTypeAlert,
		Payload: &model.WSAlert{
			ID:        alert.ID,
			AlertType: alert.AlertType,
			Severity:  alert.Severity,
			Title:     alert.Title,
			Details:   alert.Details,
			CreatedAt: alert.CreatedAt.UnixMilli(),
		},
	})
}

// TriggerNegativeBalanceAlert 触发负余额告警
func (m *AlertManager) TriggerNegativeBalanceAlert(ctx context.Context, userID int64, balance decimal.Decimal) {
	details := &model.AlertDetails{
		UserID:  &userID,
		Balance: balance,
	}
	title := fmt.Sprintf("用户 %d 余额为负: %s", userID, balance.String())
	m.createAlert(ctx, model.AlertTypeNegativeBalance, model.AlertSeverityCritical, title, details)
}

// TriggerNegativeCustodyAlert 触发负托管额度告警
func (m *AlertManager) TriggerNegativeCustodyAlert(ctx context.Context, ownerID int64, balance decimal.Decimal) {
	details := &model.AlertDetails{
		UserID:  &ownerID,
		Balance: balance,
	}
	title := fmt.Sprintf("房主 %d 托管额度为负: %s", ownerID, balance.String())
	m.createAlert(ctx, model.AlertTypeNegativeCustody, model.AlertSeverityCritical, title, details)
}

// TriggerLargeTransactionAlert 触发大额交易告警
func (m *AlertManager) TriggerLargeTransactionAlert(ctx context.Context, userID int64, amount decimal.Decimal) {
	details := &model.AlertDetails{
		UserID: &userID,
		Amount: amount,
	}
	title := fmt.Sprintf("用户 %d 大额交易: %s", userID, amount.String())
	m.createAlert(ctx, model.AlertTypeLargeTransaction, model.AlertSeverityWarning, title, details)
}

// TriggerDailyVolumeAlert 触发日交易量超限告警
func (m *AlertManager) TriggerDailyVolumeAlert(ctx context.Context, userID int64, volume decimal.Decimal) {
	details := &model.AlertDetails{
		UserID: &userID,
		Amount: volume,
	}
	title := fmt.Sprintf("用户 %d 日交易量超限: %s", userID, volume.String())
	m.createAlert(ctx, model.AlertTypeDailyVolumeExceed, model.AlertSeverityWarning, title, details)
}


// TriggerSettlementFailedAlert 触发结算失败告警
func (m *AlertManager) TriggerSettlementFailedAlert(ctx context.Context, roomID int64, failureCount int, reason string) {
	details := &model.AlertDetails{
		RoomID:       &roomID,
		FailureCount: failureCount,
		AdditionalInfo: reason,
	}
	title := fmt.Sprintf("房间 %d 结算连续失败 %d 次", roomID, failureCount)
	m.createAlert(ctx, model.AlertTypeSettlementFailed, model.AlertSeverityCritical, title, details)
}

// TriggerConservationFailedAlert 触发资金守恒检查失败告警
func (m *AlertManager) TriggerConservationFailedAlert(ctx context.Context, difference decimal.Decimal) {
	details := &model.AlertDetails{
		Difference: difference,
	}
	title := fmt.Sprintf("资金守恒检查失败，差额: %s", difference.String())
	m.createAlert(ctx, model.AlertTypeConservationFailed, model.AlertSeverityCritical, title, details)
}

// TriggerRiskFlagAlert 触发风控标记告警
func (m *AlertManager) TriggerRiskFlagAlert(ctx context.Context, flag *model.RiskFlag) {
	details := &model.AlertDetails{
		UserID:       &flag.UserID,
		RiskFlagID:   &flag.ID,
		RiskFlagType: string(flag.FlagType),
	}
	title := fmt.Sprintf("用户 %d 触发风控: %s", flag.UserID, flag.FlagType)
	m.createAlert(ctx, model.AlertTypeRiskFlagCreated, model.AlertSeverityWarning, title, details)
}

// AcknowledgeAlert 确认告警
func (m *AlertManager) AcknowledgeAlert(ctx context.Context, alertID int64, acknowledgedBy int64) error {
	return m.alertRepo.Acknowledge(ctx, alertID, acknowledgedBy)
}

// ListAlerts 列表告警
func (m *AlertManager) ListAlerts(ctx context.Context, query *model.AlertListQuery) ([]*model.Alert, int64, error) {
	return m.alertRepo.List(ctx, query)
}

// GetAlert 获取告警
func (m *AlertManager) GetAlert(ctx context.Context, id int64) (*model.Alert, error) {
	return m.alertRepo.GetByID(ctx, id)
}

// GetActiveAlertCount 获取活跃告警数量
func (m *AlertManager) GetActiveAlertCount(ctx context.Context) (int64, error) {
	return m.alertRepo.GetActiveCount(ctx)
}

// GetAlertSummary 获取告警摘要
func (m *AlertManager) GetAlertSummary(ctx context.Context) (map[model.AlertSeverity]int64, error) {
	return m.alertRepo.GetActiveBySeverity(ctx)
}
