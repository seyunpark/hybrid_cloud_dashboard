package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Restrict origins in production
		return true
	},
}

func (s *Server) handleDockerStatsWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// TODO: Stream Docker container stats
	// 1. Start a goroutine to collect stats periodically
	// 2. Broadcast stats to connected clients
	// 3. Handle client disconnection

	_ = conn.WriteJSON(gin.H{
		"type":    "docker_stats",
		"message": "Docker stats streaming not yet implemented",
	})
}

func (s *Server) handleK8sMetricsWS(c *gin.Context) {
	cluster := c.Param("cluster")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// TODO: Stream K8s metrics for the specified cluster

	_ = conn.WriteJSON(gin.H{
		"type":    "k8s_metrics",
		"cluster": cluster,
		"message": "K8s metrics streaming not yet implemented",
	})
}

func (s *Server) handleDockerLogsWS(c *gin.Context) {
	containerID := c.Param("container_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// TODO: Stream Docker container logs

	_ = conn.WriteJSON(gin.H{
		"type":         "log",
		"container_id": containerID,
		"message":      "Docker log streaming not yet implemented",
	})
}

func (s *Server) handleK8sLogsWS(c *gin.Context) {
	cluster := c.Param("cluster")
	namespace := c.Param("namespace")
	pod := c.Param("pod")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// TODO: Stream K8s pod logs

	_ = conn.WriteJSON(gin.H{
		"type":      "log",
		"cluster":   cluster,
		"namespace": namespace,
		"pod":       pod,
		"message":   "K8s log streaming not yet implemented",
	})
}

func (s *Server) handleDeployStatusWS(c *gin.Context) {
	deployID := c.Param("deploy_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	// TODO: Stream deployment status updates

	_ = conn.WriteJSON(gin.H{
		"type":      "deploy_status",
		"deploy_id": deployID,
		"message":   "Deploy status streaming not yet implemented",
	})
}
