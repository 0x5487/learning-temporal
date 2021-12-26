package signal

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nite-coder/blackbear/pkg/log"
	"go.temporal.io/sdk/workflow"
)

type CreateWithdrawOrderRequest struct {
	Amount int
}

type WithdrawOrder struct {
	OrderID string
	Amount  int
	State   int
}

type MySignal struct {
	State int
}

type Order struct {
	OrderID       string
	TXHash        string
	CallbackCount int
}

func WithdrawWorkflow(ctx workflow.Context, request CreateWithdrawOrderRequest) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	wfInfo := workflow.GetInfo(ctx)

	logger := log.FromContext(context.Background()).Logger()
	logger.Infof("withdraw workflow started, RUN_ID: %s", wfInfo.WorkflowExecution.RunID)

	order := WithdrawOrder{}
	err := workflow.ExecuteActivity(ctx, "CreateWithdrawOrderActivity", request).Get(ctx, &order)
	if err != nil {
		return err
	}

	ctx, cancel := workflow.WithCancel(ctx)
	signalChan := workflow.GetSignalChannel(ctx, "withdraw_signal")

	var signalVal MySignal

	s := workflow.NewSelector(ctx)
	s.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &signalVal)
		logger.Infof("Received message: %d", signalVal.State)
		workflow.GetLogger(ctx).Info("Received message!", signalVal.State)
	})
	s.AddFuture(workflow.NewTimer(ctx, 10*time.Second), func(f workflow.Future) {
		logger.Infof("timeout, order_id: %s", order.OrderID)
		cancel()
	})
	s.Select(ctx)

	switch signalVal.State {
	case 2:
		err = workflow.ExecuteActivity(ctx, "ApproveActivity", &order).Get(ctx, nil)
		if err != nil {
			return err
		}
	case 3:
		err = workflow.ExecuteActivity(ctx, "RejectActivity", &order).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	logger.Infof("withdraw workflow completed. order_id: %s", order.OrderID)

	return nil
}

func CreateWithdrawOrderActivity(ctx context.Context, request CreateWithdrawOrderRequest) (*WithdrawOrder, error) {
	log.Info(" === begin CreateWithdrawOrderActivity ===")

	resp := WithdrawOrder{
		OrderID: uuid.NewString(),
		Amount:  request.Amount,
		State:   1,
	}

	return &resp, nil
}

func ApproveActivity(ctx context.Context, request *Order) (*WithdrawOrder, error) {
	log.Info(" === begin ApproveActivity ===")

	return nil, nil
}

func RejectActivity(ctx context.Context, request *Order) error {
	log.Info(" === begin RejectActivity ===")

	return nil
}
