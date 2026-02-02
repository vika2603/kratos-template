package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"kratos-template/app/worker/internal/activities"
)

type ProcessOrderInput struct {
	OrderID   string
	UserID    string
	ProductID string
	Quantity  int32
	Amount    float64
}

type ProcessOrderResult struct {
	Success       bool
	OrderID       string
	ReservationID string
	TransactionID string
	Message       string
}

func ProcessOrderWorkflow(ctx workflow.Context, input ProcessOrderInput) (*ProcessOrderResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ProcessOrderWorkflow started", "OrderID", input.OrderID, "UserID", input.UserID)

	result := &ProcessOrderResult{
		OrderID: input.OrderID,
	}

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		HeartbeatTimeout:    5 * time.Second,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var validateResult activities.ValidateOrderResult
	err := workflow.ExecuteActivity(ctx, "ValidateOrder", activities.ValidateOrderInput{
		OrderID: input.OrderID,
		UserID:  input.UserID,
	}).Get(ctx, &validateResult)
	if err != nil {
		logger.Error("ValidateOrder activity failed", "error", err)
		result.Success = false
		result.Message = "order validation failed: " + err.Error()
		return result, err
	}

	if !validateResult.Valid {
		logger.Warn("Order validation failed", "message", validateResult.Message)
		result.Success = false
		result.Message = validateResult.Message
		return result, nil
	}

	var reserveResult activities.ReserveInventoryResult
	err = workflow.ExecuteActivity(ctx, "ReserveInventory", activities.ReserveInventoryInput{
		OrderID:   input.OrderID,
		ProductID: input.ProductID,
		Quantity:  input.Quantity,
	}).Get(ctx, &reserveResult)
	if err != nil {
		logger.Error("ReserveInventory activity failed", "error", err)
		result.Success = false
		result.Message = "inventory reservation failed: " + err.Error()
		return result, err
	}

	if !reserveResult.Reserved {
		logger.Warn("Inventory reservation failed", "message", reserveResult.Message)
		result.Success = false
		result.Message = reserveResult.Message
		return result, nil
	}

	result.ReservationID = reserveResult.ReservationID

	var paymentResult activities.ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, "ProcessPayment", activities.ProcessPaymentInput{
		OrderID: input.OrderID,
		Amount:  input.Amount,
		UserID:  input.UserID,
	}).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("ProcessPayment activity failed", "error", err)
		result.Success = false
		result.Message = "payment processing failed: " + err.Error()
		return result, err
	}

	if !paymentResult.Success {
		logger.Warn("Payment processing failed", "message", paymentResult.Message)
		result.Success = false
		result.Message = paymentResult.Message
		return result, nil
	}

	result.TransactionID = paymentResult.TransactionID

	notificationOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		HeartbeatTimeout:    3 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    500 * time.Millisecond,
			BackoffCoefficient: 1.5,
			MaximumInterval:    10 * time.Second,
			MaximumAttempts:    5,
		},
	}
	notificationCtx := workflow.WithActivityOptions(ctx, notificationOptions)

	err = workflow.ExecuteActivity(notificationCtx, "SendNotification", activities.SendNotificationInput{
		UserID:  input.UserID,
		OrderID: input.OrderID,
		Message: "Your order has been processed successfully",
	}).Get(notificationCtx, nil)
	if err != nil {
		logger.Warn("SendNotification activity failed (non-critical)", "error", err)
	}

	result.Success = true
	result.Message = "order processed successfully"

	logger.Info("ProcessOrderWorkflow completed", "OrderID", input.OrderID, "Success", result.Success)
	return result, nil
}
