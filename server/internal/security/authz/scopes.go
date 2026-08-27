package authz

// Scope limits the capabilities exposed by a credential.
type Scope string

const (
	ScopeOrganizationRead Scope = "organization:read"
	ScopeMembersRead      Scope = "members:read"
	ScopeMembersWrite     Scope = "members:write"
	ScopeCredentialsRead  Scope = "credentials:read"
)

func (s Scope) IsValid() bool {
	switch s {
	case ScopeOrganizationRead,
		ScopeMembersRead,
		ScopeMembersWrite,
		ScopeCredentialsRead:
		return true
	default:
		return false
	}
}
