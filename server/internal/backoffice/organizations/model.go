package organizations

type Organization struct {
	ID         string
	Name       string
	Members    int64
	Wallets    int64
	Currencies string
	Status     string
	CreatedAt  string
}

type Detail struct {
	ID           string
	Name         string
	Status       string
	Members      int64
	Wallets      int64
	Currencies   string
	BillingModel string
	CreatedAt    string
	UpdatedAt    string
	MembersList  []Member
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
