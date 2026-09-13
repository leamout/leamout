package worker

type sharedJobs struct {
	calls                   *callJobs
	recordings              *recordingJobs
}

func newSharedJobs(deps *dependencies) (*sharedJobs, error) {
	calls, err := newCallJobs(deps)
	if err != nil {
		return nil, err
	}
	recordings, err := newRecordingJobs(deps)
	if err != nil {
		return nil, err
	}
	return &sharedJobs{calls: calls, recordings: recordings}, nil
}
