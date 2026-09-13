package worker

import "github.com/leamout/leamout/internal/telecom/recordings"

type recordingJobs struct {
	consumer       *recordings.Consumer
	reconciliation *recordings.ReconciliationJob
}

func newRecordingJobs(deps *dependencies) (*recordingJobs, error) {
	repository := recordings.NewRepository(deps.db)
	service := recordings.NewService(repository, nil)
	reconciliation, err := recordings.NewReconciliationJob(
		repository,
		recordings.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		return nil, wrapWorkerError("initialize recording reconciliation job", err)
	}
	return &recordingJobs{
		consumer:       recordings.NewConsumer(service),
		reconciliation: reconciliation,
	}, nil
}
