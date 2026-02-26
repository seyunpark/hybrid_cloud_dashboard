package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func (s *Server) handleListContainers(c *gin.Context) {
	all := c.DefaultQuery("all", "false") == "true"

	containers, err := s.docker.ListContainers(c.Request.Context(), all)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DOCKER_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"containers": containers})
}

func (s *Server) handleGetContainer(c *gin.Context) {
	id := c.Param("id")

	container, err := s.docker.GetContainer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "RESOURCE_NOT_FOUND", Message: "Container not found"},
		})
		return
	}

	c.JSON(http.StatusOK, container)
}

func (s *Server) handleRestartContainer(c *gin.Context) {
	id := c.Param("id")

	if err := s.docker.RestartContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DOCKER_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Container restarted successfully",
	})
}

func (s *Server) handleStopContainer(c *gin.Context) {
	id := c.Param("id")

	if err := s.docker.StopContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DOCKER_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Container stopped successfully",
	})
}

func (s *Server) handleDeleteContainer(c *gin.Context) {
	id := c.Param("id")
	force := c.DefaultQuery("force", "false") == "true"

	if err := s.docker.DeleteContainer(c.Request.Context(), id, force); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: models.ErrorDetail{Code: "DOCKER_ERROR", Message: err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Success: true,
		Message: "Container deleted successfully",
	})
}
