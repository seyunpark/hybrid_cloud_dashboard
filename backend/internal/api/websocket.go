package api

import (
	"bufio"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
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

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	// Monitor client disconnect
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			containers, err := s.docker.ListContainers(ctx, false)
			if err != nil {
				slog.Debug("failed to list containers for stats", "error", err)
				continue
			}

			statsData := make([]gin.H, 0, len(containers))
			for _, container := range containers {
				entry := gin.H{
					"container_id": container.ID,
					"name":         container.Name,
					"state":        container.State,
				}
				if container.Stats != nil {
					entry["stats"] = container.Stats
				}
				statsData = append(statsData, entry)
			}

			if err := conn.WriteJSON(gin.H{
				"type":       "docker_stats",
				"timestamp":  time.Now().Format(time.RFC3339),
				"containers": statsData,
			}); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleK8sMetricsWS(c *gin.Context) {
	cluster := c.Param("cluster")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pods, err := s.kubernetes.ListPods(ctx, cluster, "", "")
			if err != nil {
				_ = conn.WriteJSON(gin.H{
					"type":    "k8s_metrics",
					"cluster": cluster,
					"error":   err.Error(),
				})
				continue
			}

			deployments, _ := s.kubernetes.ListDeployments(ctx, cluster, "")

			if err := conn.WriteJSON(gin.H{
				"type":        "k8s_metrics",
				"cluster":     cluster,
				"timestamp":   time.Now().Format(time.RFC3339),
				"total_pods":  len(pods),
				"deployments": len(deployments),
				"pods":        pods,
			}); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleDockerLogsWS(c *gin.Context) {
	containerID := c.Param("container_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	// Create a Docker client directly for log streaming
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "failed to create docker client"})
		return
	}
	defer cli.Close()

	logReader, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "100",
		Timestamps: true,
	})
	if err != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": err.Error()})
		return
	}
	defer logReader.Close()

	scanner := bufio.NewScanner(logReader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			// Docker log lines have 8-byte header prefix in multiplexed mode
			if len(line) > 8 {
				line = line[8:]
			}
			if err := conn.WriteJSON(gin.H{
				"type":         "log",
				"container_id": containerID,
				"message":      line,
				"timestamp":    time.Now().Format(time.RFC3339),
			}); err != nil {
				return
			}
		}
	}
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

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	// We need to get the K8s client for log streaming
	// For now, send a message that this requires direct K8s client access
	_ = conn.WriteJSON(gin.H{
		"type":      "log",
		"cluster":   cluster,
		"namespace": namespace,
		"pod":       pod,
		"message":   "Connected to K8s pod log stream",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Keep connection alive and periodically send status
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteJSON(gin.H{
				"type":      "log",
				"cluster":   cluster,
				"namespace": namespace,
				"pod":       pod,
				"message":   "heartbeat",
				"timestamp": time.Now().Format(time.RFC3339),
			}); err != nil {
				return
			}
		}
	}
}

func (s *Server) handleDeployStatusWS(c *gin.Context) {
	deployID := c.Param("deploy_id")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	var lastJSON []byte
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			state, exists := s.deployStates[deployID]
			s.mu.RUnlock()

			if !exists {
				_ = conn.WriteJSON(gin.H{
					"type":      "deploy_status",
					"deploy_id": deployID,
					"status":    "not_found",
				})
				return
			}

			currentJSON, _ := json.Marshal(state.Status)
			if string(currentJSON) != string(lastJSON) {
				if err := conn.WriteJSON(gin.H{
					"type":      "deploy_status",
					"deploy_id": deployID,
					"data":      state.Status,
				}); err != nil {
					return
				}
				lastJSON = currentJSON
			}

			// If deployment is completed or failed, send final update and close
			if state.Status.Status == "completed" || state.Status.Status == "failed" || state.Status.Status == "cancelled" {
				return
			}
		}
	}
}

