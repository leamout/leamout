package users

const pageSize = 25

type Filter struct {
	Search         string
	Status         string
	Admin          string
	Verified       string
	OrganizationID string
	Page           int32
}

type User struct {
	ID                string
	Name              string
	Email             string
	EmailVerified     bool
	IsPlatformAdmin   bool
	Status            string
	OrganizationCount int64
	CreatedAt         string
}

type List struct {
	Users   []User
	Filter  Filter
	HasNext bool
}

type Detail struct {
	User
	ActiveSessionCount int64
	LastSeenAt         string
	UpdatedAt          string
	DisabledAt         string
	Memberships        []Membership
	Sessions           []Session
	MemberPage         int32
	HasNextMembers     bool
}

type Membership struct {
	OrganizationID     string
	Name               string
	OrganizationStatus string
	Role               string
	Status             string
	JoinedAt           string
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
