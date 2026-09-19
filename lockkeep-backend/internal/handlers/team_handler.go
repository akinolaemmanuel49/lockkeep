package handlers

import (
	"errors"
	"net/http"

	"github.com/akinolaemmanuel49/lockkeep-backend/internal/middleware"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/ports"
	"github.com/akinolaemmanuel49/lockkeep-backend/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TeamHandler struct {
	teamService ports.TeamService
}

func NewTeamHandler(teamService ports.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

type CreateTeamRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

func (h *TeamHandler) Create(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	team, err := h.teamService.CreateTeam(c.Request.Context(), objUserID, objOrgID, req.Name, req.Slug)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrTeamSlugTaken):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"team":    team,
		"message": "team created successfully",
	})
}

func (h *TeamHandler) List(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	teams, err := h.teamService.ListTeamsByOrganization(c.Request.Context(), objOrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"teams": teams})
}

func (h *TeamHandler) Get(c *gin.Context) {
	orgID := c.GetString(middleware.CtxKeyOrgID)
	slug := c.Param("teamSlug")

	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}

	team, err := h.teamService.GetTeamBySlug(c.Request.Context(), objOrgID, slug)
	if err != nil {
		if errors.Is(err, services.ErrTeamNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"team": team})
}

func (h *TeamHandler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.CtxKeyUserID)
	orgID := c.GetString(middleware.CtxKeyOrgID)
	teamID := c.Param("teamID")

	objUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	objOrgID, err := bson.ObjectIDFromHex(orgID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization"})
		return
	}
	objTeamID, err := bson.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team id"})
		return
	}

	if err := h.teamService.DeleteTeams(c.Request.Context(), objUserID, objOrgID, []bson.ObjectID{objTeamID}); err != nil {
		switch {
		case errors.Is(err, services.ErrEntityHasNoRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrEntityHasInvalidRoles):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, services.ErrTeamNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "team deleted successfully"})
}
