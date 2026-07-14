package middleware

// import (
// 	"net/http"

// 	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
// 	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
// 	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
// 	"github.com/gin-gonic/gin"
// )

// const (
// 	CtxKeyOrgID      = "orgID"
// 	CtxKeyMembership = "membership"
// 	CtxKeyTeamID     = "teamID"
// )

// // RequireOrgAccess verifies the user is a member of the requested organization
// func RequireOrgAccess(membershipRepo ports.MembershipRepository) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userID := c.GetString(CtxKeyUserID)
// 		orgSlug := c.Param("orgSlug")

// 		// Need to resolve slug → org ID first (via org repo lookup)
// 		// Simplified: assume orgID is available or resolved earlier
// 		// Full implementation needs OrganizationRepository injected

// 		c.Next()
// 	}
// }

// // RequirePermission checks if the user has a specific permission
// func RequirePermission(perm domain.Permission) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		claims := c.MustGet(CtxKeyClaims).(jwt.Claims)

// 		// System admin bypass
// 		if claims.SystemRole == string(domain.RoleSystemAdmin) {
// 			c.Next()
// 			return
// 		}

// 		// Check org-level permission
// 		orgID := c.GetString(CtxKeyOrgID)
// 		if orgID != "" {
// 			for _, org := range claims.Orgs {
// 				if org.OrgID == orgID {
// 					if hasPermission(domain.RoleID(org.Role), perm) {
// 						c.Next()
// 						return
// 					}

// 					// Check team-level if team context
// 					teamID := c.GetString(CtxKeyTeamID)
// 					if teamID != "" {
// 						if teamRole, ok := org.TeamRoles[teamID]; ok {
// 							if hasPermission(domain.RoleID(teamRole), perm) {
// 								c.Next()
// 								return
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}

// 		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
// 	}
// }

// func hasPermission(roleID domain.RoleID, perm domain.Permission) bool {
// 	perms, ok := domain.RolePermissions[roleID]
// 	if !ok {
// 		return false
// 	}
// 	for _, p := range perms {
// 		if p == perm {
// 			return true
// 		}
// 	}
// 	return false
// }
