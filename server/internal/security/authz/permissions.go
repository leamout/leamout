package authz

// Permission identifies an action that a principal may perform.
type Permission string

const (
	PermissionOrganizationRead   Permission = "organization:read"
	PermissionOrganizationUpdate Permission = "organization:update"
	PermissionOrganizationDelete Permission = "organization:delete"
	PermissionMembersRead        Permission = "members:read"
	PermissionMembersWrite       Permission = "members:write"
	PermissionCredentialsRead    Permission = "credentials:read"
	PermissionCredentialsWrite   Permission = "credentials:write"
)
