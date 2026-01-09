package logic

import (
	"context"
	"os"
	"testing"
	"time"

	"SLGaming/back/services/order/internal/config"
	"SLGaming/back/services/order/internal/model"
	"SLGaming/back/services/order/internal/svc"
	"SLGaming/back/services/order/order"
	orderioc "SLGaming/back/services/order/internal/ioc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// setupTestDB 初始化测试数据库连接
func setupTestDB(t *testing.T) *gorm.DB {
	// 从环境变量获取数据库连接信息，如果没有则使用默认值
	dsn := os.Getenv("TEST_MYSQL_DSN")
	if dsn == "" {
		// 默认测试数据库配置，可以根据实际情况修改
		dsn = "root:123456@tcp(localhost:3306)/slgaming_test?charset=utf8mb4&parseTime=true&loc=Local"
	}

	cfg := config.MysqlConf{
		DSN:             dsn,
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: 300 * time.Second,
		ConnMaxIdleTime: 60 * time.Second,
	}

	db, err := orderioc.InitMysql(cfg)
	require.NoError(t, err, "初始化测试数据库失败")

	// 自动迁移表结构
	err = db.AutoMigrate(&model.Order{}, &model.OrderEventOutbox{})
	require.NoError(t, err, "数据库表迁移失败")

	return db
}

// cleanupTestData 清理测试数据
func cleanupTestData(t *testing.T, db *gorm.DB, orderIDs []uint64) {
	if len(orderIDs) > 0 {
		db.Where("id IN ?", orderIDs).Delete(&model.Order{})
	}
	// 清理 outbox 表中的测试数据
	db.Where("event_type = ?", "ORDER_CANCELLED").Delete(&model.OrderEventOutbox{})
}

// createTestOrder 创建测试订单
func createTestOrder(t *testing.T, db *gorm.DB, status int32, bossID, companionID uint64) *model.Order {
	now := time.Now()
	order := &model.Order{
		OrderNo:        "TEST" + time.Now().Format("20060102150405"),
		BossID:         bossID,
		CompanionID:    companionID,
		GameName:       "测试游戏",
		GameMode:       "排位赛",
		DurationMinutes: 60,
		PricePerHour:   100,
		TotalAmount:    100,
		Status:         status,
		CreatedAt:      now,
	}

	switch status {
	case model.OrderStatusPaid:
		order.PaidAt = &now
	case model.OrderStatusAccepted:
		order.PaidAt = &now
		acceptedAt := now.Add(5 * time.Minute)
		order.AcceptedAt = &acceptedAt
	case model.OrderStatusCompleted:
		order.PaidAt = &now
		acceptedAt := now.Add(5 * time.Minute)
		order.AcceptedAt = &acceptedAt
		startAt := now.Add(10 * time.Minute)
		order.StartAt = &startAt
		completedAt := now.Add(70 * time.Minute)
		order.CompletedAt = &completedAt
	case model.OrderStatusCancelled:
		cancelledAt := now.Add(5 * time.Minute)
		order.CancelledAt = &cancelledAt
		order.CancelReason = "测试取消"
	}

	err := db.Create(order).Error
	require.NoError(t, err, "创建测试订单失败")

	return order
}

// TestCancelOrderLogic_CancelOrder_CreatedStatus_BossCanCancel 测试已创建状态下老板可以取消订单
func TestCancelOrderLogic_CancelOrder_CreatedStatus_BossCanCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	// 创建测试订单
	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusCreated, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 老板取消订单
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: bossID,
		Reason:     "不想玩了",
	}

	resp, err := logic.CancelOrder(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Order)
	assert.Equal(t, model.OrderStatusCancelled, resp.Order.Status)
	assert.Equal(t, "不想玩了", resp.Order.CancelReason)
	assert.NotZero(t, resp.Order.CancelledAt)

	// 验证数据库中的状态
	var dbOrder model.Order
	err = db.Where("id = ?", order.ID).First(&dbOrder).Error
	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusCancelled, dbOrder.Status)
	assert.Equal(t, "不想玩了", dbOrder.CancelReason)

	// 验证没有创建退款事件（因为未支付）
	var outboxCount int64
	db.Model(&model.OrderEventOutbox{}).Where("event_type = ? AND payload LIKE ?", "ORDER_CANCELLED", "%"+order.OrderNo+"%").Count(&outboxCount)
	assert.Equal(t, int64(0), outboxCount, "未支付订单取消不应该创建退款事件")
}

// TestCancelOrderLogic_CancelOrder_CreatedStatus_CompanionCannotCancel 测试已创建状态下陪玩不能取消订单
func TestCancelOrderLogic_CancelOrder_CreatedStatus_CompanionCannotCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusCreated, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 陪玩尝试取消订单（应该失败）
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: companionID,
		Reason:     "不想接单",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "only the boss can cancel")

	// 验证订单状态未改变
	var dbOrder model.Order
	err = db.Where("id = ?", order.ID).First(&dbOrder).Error
	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusCreated, dbOrder.Status)
}

// TestCancelOrderLogic_CancelOrder_PaidStatus_BossCanCancel 测试已支付状态下老板可以取消订单并触发退款
func TestCancelOrderLogic_CancelOrder_PaidStatus_BossCanCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusPaid, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 老板取消订单
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: bossID,
		Reason:     "改变主意了",
	}

	resp, err := logic.CancelOrder(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, model.OrderStatusCancelRefunding, resp.Order.Status)

	// 验证数据库中的状态
	var dbOrder model.Order
	err = db.Where("id = ?", order.ID).First(&dbOrder).Error
	require.NoError(t, err)
	assert.Equal(t, model.OrderStatusCancelRefunding, dbOrder.Status)

	// 验证创建了退款事件
	var outbox model.OrderEventOutbox
	err = db.Where("event_type = ? AND payload LIKE ?", "ORDER_CANCELLED", "%"+order.OrderNo+"%").First(&outbox).Error
	require.NoError(t, err)
	assert.Equal(t, "ORDER_CANCELLED", outbox.EventType)
	assert.Equal(t, "PENDING", outbox.Status)
}

// TestCancelOrderLogic_CancelOrder_AcceptedStatus_CompanionCanCancel 测试已接单状态下只有陪玩可以取消订单
func TestCancelOrderLogic_CancelOrder_AcceptedStatus_CompanionCanCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusAccepted, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 陪玩取消订单
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: companionID,
		Reason:     "临时有事",
	}

	resp, err := logic.CancelOrder(req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, model.OrderStatusCancelRefunding, resp.Order.Status)

	// 验证创建了退款事件
	var outboxCount int64
	db.Model(&model.OrderEventOutbox{}).Where("event_type = ? AND payload LIKE ?", "ORDER_CANCELLED", "%"+order.OrderNo+"%").Count(&outboxCount)
	assert.Equal(t, int64(1), outboxCount)
}

// TestCancelOrderLogic_CancelOrder_AcceptedStatus_BossCannotCancel 测试已接单状态下老板不能取消订单
func TestCancelOrderLogic_CancelOrder_AcceptedStatus_BossCannotCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusAccepted, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 老板尝试取消订单（应该失败）
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: bossID,
		Reason:     "不想玩了",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, st.Code())
	assert.Contains(t, st.Message(), "only the companion can cancel")
}

// TestCancelOrderLogic_CancelOrder_CompletedStatus_CannotCancel 测试已完成状态下不能取消订单
func TestCancelOrderLogic_CancelOrder_CompletedStatus_CannotCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusCompleted, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	// 尝试取消已完成订单（应该失败）
	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: bossID,
		Reason:     "测试",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Contains(t, st.Message(), "already finished or cancelled")
}

// TestCancelOrderLogic_CancelOrder_CancelledStatus_CannotCancel 测试已取消状态下不能再次取消
func TestCancelOrderLogic_CancelOrder_CancelledStatus_CannotCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusCancelled, bossID, companionID)
	defer cleanupTestData(t, db, []uint64{order.ID})

	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: bossID,
		Reason:     "再次取消",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
}

// TestCancelOrderLogic_CancelOrder_OrderNotFound 测试订单不存在的情况
func TestCancelOrderLogic_CancelOrder_OrderNotFound(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	req := &order.CancelOrderRequest{
		OrderId:    999999999,
		OperatorId: 1001,
		Reason:     "测试",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "not found")
}

// TestCancelOrderLogic_CancelOrder_InvalidOrderId 测试无效订单ID
func TestCancelOrderLogic_CancelOrder_InvalidOrderId(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	req := &order.CancelOrderRequest{
		OrderId:    0, // 无效ID
		OperatorId: 1001,
		Reason:     "测试",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "required")
}

// TestCancelOrderLogic_CancelOrder_InService_CannotCancel 测试服务中状态下不能取消
func TestCancelOrderLogic_CancelOrder_InService_CannotCancel(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestData(t, db, nil)

	ctx := context.Background()
	svcCtx := &svc.ServiceContext{
		DB: db,
	}

	logic := NewCancelOrderLogic(ctx, svcCtx)

	bossID := uint64(1001)
	companionID := uint64(2001)
	order := createTestOrder(t, db, model.OrderStatusAccepted, bossID, companionID)
	
	// 更新为服务中状态
	startAt := time.Now()
	order.Status = model.OrderStatusInService
	order.StartAt = &startAt
	db.Save(order)
	defer cleanupTestData(t, db, []uint64{order.ID})

	req := &order.CancelOrderRequest{
		OrderId:    order.ID,
		OperatorId: companionID,
		Reason:     "测试",
	}

	resp, err := logic.CancelOrder(req)
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.FailedPrecondition, st.Code())
	assert.Contains(t, st.Message(), "in service")
}

func init() {
	// 初始化日志，避免测试时出现日志错误
	logx.DisableStat()
}

