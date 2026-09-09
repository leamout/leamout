package trunks

type Trunk struct {
	ID           string
	Organization string
	Name         string
	Mode         string
	Provider     string
	Direction    string
	Status       string
	Endpoints    int64
}
