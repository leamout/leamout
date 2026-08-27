package authz

// Scope limits the capabilities exposed by a credential.
type Scope string

const (
	ScopeOrganizationRead Scope = "organization:read"
	ScopeMembersRead      Scope = "members:read"
	ScopeMembersWrite     Scope = "members:write"
	ScopeCredentialsRead  Scope = "credentials:read"
	ScopeCredentialsWrite Scope = "credentials:write"
)
