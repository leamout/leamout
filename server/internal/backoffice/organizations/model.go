package organizations

type Organization struct {
	ID        string
	Name      string
	Plan      string
	Members   int64
	Status    string
	CreatedAt string
}

type Detail struct {
	ID                     string
	Name                   string
	Status                 string
	Members                int64
	Plan                   string
	SubscriptionStatus     string
	BillingProvider        string
	ProviderSubscriptionID string
	RenewsAt               string
	EndsAt                 string
	CreatedAt              string
	UpdatedAt              string
	MembersList            []Member
}

type Member struct {
	UserID           string
	Name             string
	Email            string
	Role             string
	MembershipStatus string
	UserStatus       string
	EmailVerified    bool
	PlatformAdmin    bool
	JoinedAt         string
}
