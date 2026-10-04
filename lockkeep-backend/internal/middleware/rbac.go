package middleware

import (
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/domain"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/utils"
	"github.com/akinolaemmanuel49/lockkeep-backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CtxKeyOrgID      = "orgID"
	CtxKeyTeamID     = "teamID"
	CtxKeyMembership = "membership"
)

// RequireOrgAccess verifies the user is a member of the requested organization
func RequireOrgAccess(membershipRepo ports.MembershipRepository, orgRepo ports.OrganizationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(CtxKeyUserID)
		orgSlug := c.Param("orgSlug")

		// Resolve slug to org
		org, err := orgRepo.FindBySlug(c.Request.Context(), orgSlug)
		if err != nil || org == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "organization not found"})
			return
		}

		objUserID, err := bson.ObjectIDFromHex(userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
			return
		}

		membership, err := membershipRepo.FindByUserAndOrg(c.Request.Context(), objUserID, org.ID)
		if err != nil || membership == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not a member of this organization"})
			return
		}

		c.Set(CtxKeyOrgID, org.ID.Hex())
		c.Set(CtxKeyMembership, membership)
		c.Next()
	}
}

// RequirePermission checks if the user has a specific permission.
//
// It prefers the authoritative DB membership loaded by RequireOrgAccess (see
// CtxKeyMembership) so permissions reflect the current organization state
// rather than the JWT snapshot taken at login — a freshly created workspace is
// usable immediately without re-issuing the access token. It falls back to the
// org claims embedded in the token only when no membership is in context.
func RequirePermission(perm domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet(CtxKeyClaims).(*jwt.Claims)

		// System admin bypass
		if claims.SystemRole == string(domain.RoleSystemAdmin) {
			c.Next()
			return
		}

		if membership, ok := c.Get(CtxKeyMembership); ok {
			m, ok := membership.(*domain.Membership)
			if ok && m != nil {
				if utils.HasPermission(m.RoleID, perm) {
					c.Next()
					return
				}

				teamID := c.GetString(CtxKeyTeamID)
				for _, tr := range m.TeamRoles {
					if teamID != "" && tr.TeamID.Hex() == teamID {
						if utils.HasPermission(tr.RoleID, perm) {
							c.Next()
							return
						}
					}
				}

				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
		}

		orgID := c.GetString(CtxKeyOrgID)
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "organization context required"})
			return
		}

		for _, org := range claims.Orgs {
			if org.OrgID == orgID {
				if utils.HasPermission(domain.RoleID(org.Role), perm) {
					c.Next()
					return
				}

				teamID := c.GetString(CtxKeyTeamID)
				if teamID != "" {
					if teamRole, ok := org.TeamRoles[teamID]; ok {
						if utils.HasPermission(domain.RoleID(teamRole), perm) {
							c.Next()
							return
						}
					}
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
