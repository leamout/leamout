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

type Operation struct {
	ID, Organization, Number, Type, State, ProviderOperationID, LastError, CreatedAt, UpdatedAt string
	Attempts                                                                                    int32
}
type Detail struct {
	Provider
	PhoneNumbers         int64
	CreatedAt, UpdatedAt string
	Operations           []Operation
}
