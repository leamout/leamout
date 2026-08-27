package authz

// Allows reports whether a role grants a permission.
func Allows(role Role, permission Permission) bool {
	switch role {
	case RoleOwner:
		return true
	case RoleAdmin:
		switch permission {
		case PermissionOrganizationRead, PermissionOrganizationUpdate,
			PermissionMembersRead, PermissionMembersWrite,
			PermissionCredentialsRead, PermissionCredentialsWrite:
			return true
		}
	case RoleMember:
		switch permission {
		case PermissionOrganizationRead, PermissionMembersRead:
			return true
		}
	}
	return false
}
