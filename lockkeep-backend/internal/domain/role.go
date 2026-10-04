package domain

import "time"

type RoleID string

// System roles
const (
	RoleSystemAdmin RoleID = "system:admin"
	RoleSystemUser  RoleID = "system:user"
)

// Organization roles
const (
	RoleOrgOwner  RoleID = "org:owner"
	RoleOrgAdmin  RoleID = "org:admin"
	RoleOrgMember RoleID = "org:member"
)

// Team roles
const (
	RoleTeamAdmin RoleID = "team:admin"
	RoleTeamUser  RoleID = "team:user"
	RoleTeamRead  RoleID = "team:read"
)

type RoleScope string

const (
	RoleScopeSystem       RoleScope = "system"
	RoleScopeOrganization RoleScope = "organization"
	RoleScopeTeam         RoleScope = "team"
)

type Role struct {
	ID          RoleID       `bson:"_id" json:"id"`
	Name        string       `bson:"name" json:"name"`
	Description string       `bson:"description" json:"description"`
	Scope       RoleScope    `bson:"scope" json:"scope"`
	Permissions []Permission `bson:"permissions" json:"permissions"`
	CreatedAt   time.Time    `bson:"created_at" json:"createdAt"`
}

// Role assignment maps
var RolePermissions = map[RoleID][]Permission{
	RoleSystemAdmin: {
		PermSecretRead, PermSecretWrite, PermSecretDelete, PermSecretShare,
		PermOrgRead, PermOrgManage, PermOrgDelete,
		PermTeamRead, PermTeamManage,
		PermMemberInvite, PermMemberRemove, PermMemberRole,
	},
	RoleSystemUser: {
		PermSecretRead, PermSecretWrite, PermSecretDelete,
	},
	RoleOrgOwner: {
		PermSecretRead, PermSecretWrite, PermSecretDelete, PermSecretShare,
		PermOrgRead, PermOrgManage, PermOrgDelete,
		PermTeamRead, PermTeamManage,
		PermMemberInvite, PermMemberRemove, PermMemberRole,
	},
	RoleOrgAdmin: {
		PermSecretRead, PermSecretWrite, PermSecretDelete, PermSecretShare,
		PermOrgRead, PermOrgManage,
		PermTeamRead, PermTeamManage,
		PermMemberInvite, PermMemberRemove, PermMemberRole,
	},
	RoleOrgMember: {
		PermSecretRead,
	},
	RoleTeamAdmin: {
		PermSecretRead, PermSecretWrite, PermSecretDelete, PermSecretShare,
		PermTeamRead, PermTeamManage,
		PermMemberInvite, PermMemberRemove, PermMemberRole,
	},
	RoleTeamUser: {
		PermSecretRead, PermSecretWrite,
	},
	RoleTeamRead: {
		PermSecretRead,
	},
}
