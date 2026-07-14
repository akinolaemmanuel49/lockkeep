package domain

type Permission string

const (
	PermSecretRead   Permission = "secret:read"
	PermSecretWrite  Permission = "secret:write"
	PermSecretDelete Permission = "secret:delete"
	PermSecretShare  Permission = "secret:share"

	PermOrgRead   Permission = "org:read"
	PermOrgManage Permission = "org:manage"
	PermOrgDelete Permission = "org:delete"

	PermTeamRead   Permission = "team:read"
	PermTeamManage Permission = "team:manage"

	PermMemberInvite Permission = "member:invite"
	PermMemberRemove Permission = "member:remove"
	PermMemberRole   Permission = "member:role"
)
