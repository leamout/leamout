package providers

type Provider struct {
	ID                string
	Slug              string
	Name              string
	Adapter           string
	Status            string
	Connections       int64
	PendingOperations int64
	FailedOperations  int64
}
