package authz

// Guard evaluates a role against a required permission.
func Guard(role Role, permission Permission) bool {
	return Allows(role, permission)
}

// HasScope reports whether a credential contains a scope.
func HasScope(scopes []Scope, required Scope) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}
