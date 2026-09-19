package utils

import "github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"

func HasPermission(roleID domain.RoleID, perm domain.Permission) bool {
	perms, ok := domain.RolePermissions[roleID]
	if !ok {
		return false
	}
	for _, p := range perms {
		if p == perm {
			return true
		}
	}
	return false
}
