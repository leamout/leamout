package conversations

// TurnRecovery carries the minimal durable turn data necessary to resume
// monotonic conversation sequencing after a worker restart.
type TurnRecovery struct {
	Sequence int32
}
