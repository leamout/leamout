package authz

// Policy is a small, dependency-free authorization policy.
type Policy struct{}

func (Policy) Allows(role Role, permission Permission) bool {
	return Allows(role, permission)
}
