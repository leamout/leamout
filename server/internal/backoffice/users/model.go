package users

type User struct {
	ID                string
	Name              string
	Email             string
	EmailVerified     bool
	PlatformAdmin     bool
	Status            string
	OrganizationCount int64
	CreatedAt         string
}

type OrganizationMembership struct {
	OrganizationID     string
	OrganizationName   string
	OrganizationStatus string
	DetailAvailable    bool
	Role               string
	MembershipStatus   string
	JoinedAt            string
}

type Session struct {
	ID         string
	Assurance  string
	IPAddress  string
	UserAgent  string
	Status     string
	CreatedAt  string
	LastSeenAt string
	ExpiresAt  string
	RevokedAt  string
}

type Detail struct {
	User
	ActiveSessions int64
	LastSeenAt     string
	UpdatedAt      string
	DisabledAt     string
	Organizations  []OrganizationMembership
	Sessions       []Session
}
