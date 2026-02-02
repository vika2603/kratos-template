package worker

import (
	"context"
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	"kratos-template/app/worker/internal/activities"
	"kratos-template/app/worker/internal/conf"
	"kratos-template/app/worker/internal/workflows"
)

type WorkerParams struct {
	fx.In
	Config     *conf.Bootstrap
	Logger     log.Logger
	Activities *activities.Activities
}

type WorkerResult struct {
	fx.Out
	Worker worker.Worker
	Client client.Client
}

func NewTemporalWorker(lc fx.Lifecycle, params WorkerParams) (WorkerResult, error) {
	logger := log.NewHelper(params.Logger)

	hostPort := os.Getenv("TEMPORAL_ADDR")
	if hostPort == "" {
		hostPort = params.Config.Temporal.HostPort
	}
	clientOptions := client.Options{
		HostPort:  hostPort,
		Namespace: params.Config.Temporal.Namespace,
	}

	temporalClient, err := client.Dial(clientOptions)
	if err != nil {
		logger.Errorf("Failed to create Temporal client: %v", err)
		return WorkerResult{}, err
	}

	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:     int(params.Config.Temporal.MaxConcurrentActivities),
		MaxConcurrentWorkflowTaskExecutionSize: int(params.Config.Temporal.MaxConcurrentWorkflows),
	}

	temporalWorker := worker.New(temporalClient, params.Config.Temporal.TaskQueue, workerOptions)

	temporalWorker.RegisterWorkflow(workflows.ProcessOrderWorkflow)

	temporalWorker.RegisterActivity(params.Activities.ValidateOrder)
	temporalWorker.RegisterActivity(params.Activities.ReserveInventory)
	temporalWorker.RegisterActivity(params.Activities.ProcessPayment)
	temporalWorker.RegisterActivity(params.Activities.SendNotification)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting Temporal worker")
			go func() {
				if err := temporalWorker.Run(worker.InterruptCh()); err != nil {
					logger.Errorf("Temporal worker run error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping Temporal worker")
			temporalWorker.Stop()
			temporalClient.Close()
			return nil
		},
	})

	return WorkerResult{
		Worker: temporalWorker,
		Client: temporalClient,
	}, nil
}
