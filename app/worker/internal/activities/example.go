package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.temporal.io/sdk/activity"
)

type ValidateOrderInput struct {
	OrderID string
	UserID  string
}

type ValidateOrderResult struct {
	Valid   bool
	Message string
}

func (a *Activities) ValidateOrder(ctx context.Context, input ValidateOrderInput) (*ValidateOrderResult, error) {
	logger := log.NewHelper(a.logger)
	logger.Infof("ValidateOrder activity started: OrderID=%s, UserID=%s", input.OrderID, input.UserID)

	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			logger.Warn("ValidateOrder activity cancelled")
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
			activity.RecordHeartbeat(ctx, i+1)
		}
	}

	if input.OrderID == "" {
		return &ValidateOrderResult{
			Valid:   false,
			Message: "order ID is required",
		}, nil
	}

	return &ValidateOrderResult{
		Valid:   true,
		Message: "order is valid",
	}, nil
}

type ReserveInventoryInput struct {
	OrderID   string
	ProductID string
	Quantity  int32
}

type ReserveInventoryResult struct {
	Reserved      bool
	ReservationID string
	Message       string
}

func (a *Activities) ReserveInventory(ctx context.Context, input ReserveInventoryInput) (*ReserveInventoryResult, error) {
	logger := log.NewHelper(a.logger)
	logger.Infof("ReserveInventory activity started: OrderID=%s, ProductID=%s, Quantity=%d",
		input.OrderID, input.ProductID, input.Quantity)

	for i := 0; i < 3; i++ {
		select {
		case <-ctx.Done():
			logger.Warn("ReserveInventory activity cancelled")
			return nil, ctx.Err()
		case <-time.After(300 * time.Millisecond):
			activity.RecordHeartbeat(ctx, i+1)
		}
	}

	if input.Quantity <= 0 {
		return &ReserveInventoryResult{
			Reserved: false,
			Message:  "invalid quantity",
		}, nil
	}

	reservationID := fmt.Sprintf("RES-%s-%d", input.OrderID, time.Now().Unix())

	return &ReserveInventoryResult{
		Reserved:      true,
		ReservationID: reservationID,
		Message:       "inventory reserved successfully",
	}, nil
}

type ProcessPaymentInput struct {
	OrderID string
	Amount  float64
	UserID  string
}

type ProcessPaymentResult struct {
	Success       bool
	TransactionID string
	Message       string
}

func (a *Activities) ProcessPayment(ctx context.Context, input ProcessPaymentInput) (*ProcessPaymentResult, error) {
	logger := log.NewHelper(a.logger)
	logger.Infof("ProcessPayment activity started: OrderID=%s, Amount=%.2f, UserID=%s",
		input.OrderID, input.Amount, input.UserID)

	for i := 0; i < 4; i++ {
		select {
		case <-ctx.Done():
			logger.Warn("ProcessPayment activity cancelled")
			return nil, ctx.Err()
		case <-time.After(250 * time.Millisecond):
			activity.RecordHeartbeat(ctx, i+1)
		}
	}

	if input.Amount <= 0 {
		return &ProcessPaymentResult{
			Success: false,
			Message: "invalid payment amount",
		}, nil
	}

	transactionID := fmt.Sprintf("TXN-%s-%d", input.OrderID, time.Now().Unix())

	return &ProcessPaymentResult{
		Success:       true,
		TransactionID: transactionID,
		Message:       "payment processed successfully",
	}, nil
}

type SendNotificationInput struct {
	UserID  string
	OrderID string
	Message string
}

func (a *Activities) SendNotification(ctx context.Context, input SendNotificationInput) error {
	logger := log.NewHelper(a.logger)
	logger.Infof("SendNotification activity started: UserID=%s, OrderID=%s, Message=%s",
		input.UserID, input.OrderID, input.Message)

	select {
	case <-ctx.Done():
		logger.Warn("SendNotification activity cancelled")
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
		activity.RecordHeartbeat(ctx, "notification sent")
	}

	logger.Infof("Notification sent successfully to user %s", input.UserID)
	return nil
}
