package authz

// Principal identifies the caller for an authorization decision. Scopes are
// optional and constrain principals authenticated with a scoped credential.
type Principal struct {
	Role   Role
	Scopes []Scope
}

// Access evaluates permissions for a principal.
type Access struct{}

func (Access) Allows(principal Principal, permission Permission) bool {
	return Allows(principal.Role, permission)
}

func (Access) AllowsScoped(principal Principal, permission Permission, scope Scope) bool {
	if !Allows(principal.Role, permission) {
		return false
	}
	if principal.Scopes == nil {
		return true
	}
	return HasScope(principal.Scopes, scope)
}
