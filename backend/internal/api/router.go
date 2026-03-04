package api

import (
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/ai"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/data"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/docker"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/kubernetes"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/metrics"
	"github.com/seyunpark/hybrid_cloud_dashboard/internal/registry"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// deployState holds in-memory state for an active deployment.
type deployState struct {
	Status    *models.DeployStatus
	Response  *models.DeployResponse
	Request   *models.DeployRequest
	Manifests *models.ManifestResult
}

// Server holds all dependencies for the HTTP server.
type Server struct {
	cfg        *config.Config
	router     *gin.Engine
	docker     docker.Service
	kubernetes kubernetes.Service
	ai         ai.Service
	data       data.Store
	registry   registry.Service
	metrics    *metrics.Collector

	deployStates      map[string]*deployState
	stackDeployStates map[string]*stackDeployState
	mu                sync.RWMutex
}

// NewServer creates and configures a new API server with all routes registered.
func NewServer(
	cfg *config.Config,
	dockerSvc docker.Service,
	k8sSvc kubernetes.Service,
	aiSvc ai.Service,
	dataStore data.Store,
	registrySvc registry.Service,
	metricsColl *metrics.Collector,
) *Server {
	s := &Server{
		cfg:          cfg,
		docker:       dockerSvc,
		kubernetes:   k8sSvc,
		ai:           aiSvc,
		data:         dataStore,
		registry:     registrySvc,
		metrics:      metricsColl,
		deployStates:      make(map[string]*deployState),
		stackDeployStates: make(map[string]*stackDeployState),
	}

	s.setupRouter()
	return s
}

func (s *Server) setupRouter() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(RequestLogger())
	r.Use(ErrorHandler())

	// CORS
	if s.cfg.Security.CORS.Enabled {
		corsConfig := cors.Config{
			AllowMethods:    s.cfg.Security.CORS.AllowedMethods,
			AllowHeaders:    s.cfg.Security.CORS.AllowedHeaders,
			AllowWebSockets: true,
		}
		// "*" in allowed_origins means allow all
		if len(s.cfg.Security.CORS.AllowedOrigins) == 1 && s.cfg.Security.CORS.AllowedOrigins[0] == "*" {
			corsConfig.AllowAllOrigins = true
		} else {
			corsConfig.AllowOrigins = s.cfg.Security.CORS.AllowedOrigins
			corsConfig.AllowCredentials = true
		}
		r.Use(cors.New(corsConfig))
	}

	// Health checks
	r.GET("/health", s.handleHealth)
	r.GET("/ready", s.handleReady)

	// REST API
	api := r.Group("/api")
	{
		// Docker
		dockerGroup := api.Group("/docker")
		{
			dockerGroup.GET("/containers", s.handleListContainers)
			dockerGroup.GET("/containers/:id", s.handleGetContainer)
			dockerGroup.POST("/containers/:id/restart", s.handleRestartContainer)
			dockerGroup.POST("/containers/:id/stop", s.handleStopContainer)
			dockerGroup.DELETE("/containers/:id", s.handleDeleteContainer)
		}

		// Kubernetes
		k8sGroup := api.Group("/k8s")
		{
			k8sGroup.GET("/clusters", s.handleListClusters)
			k8sGroup.GET("/:cluster/namespaces", s.handleListNamespaces)
			k8sGroup.GET("/:cluster/pods", s.handleListPods)
			k8sGroup.GET("/:cluster/deployments", s.handleListDeployments)
			k8sGroup.GET("/:cluster/services", s.handleListServices)
			k8sGroup.POST("/:cluster/deployments/:ns/:name/scale", s.handleScaleDeployment)
			k8sGroup.POST("/:cluster/pods/:ns/:name/restart", s.handleRestartPod)
		}

		// Deploy
		deployGroup := api.Group("/deploy")
		{
			deployGroup.POST("/docker-to-k8s", s.handleDeployDockerToK8s)
			deployGroup.POST("/:deploy_id/execute", s.handleExecuteDeploy)
			deployGroup.POST("/:deploy_id/refine", s.handleRefineDeploy)
			deployGroup.POST("/:deploy_id/undeploy", s.handleUndeployFromK8s)
			deployGroup.POST("/:deploy_id/redeploy", s.handleRedeployToK8s)
			deployGroup.DELETE("/:deploy_id", s.handleDeleteDeployRecord)
			deployGroup.GET("/:deploy_id/status", s.handleGetDeployStatus)
			deployGroup.GET("/history", s.handleGetDeployHistory)

			// Stack Deploy
			deployGroup.GET("/stack", s.handleListActiveStackDeploys)
			deployGroup.GET("/stack/:deploy_id", s.handleGetStackDeployDetail)
			deployGroup.GET("/stack/:deploy_id/status", s.handleGetStackDeployStatus)
			deployGroup.POST("/stack", s.handleDeployStack)
			deployGroup.POST("/stack/:deploy_id/refine", s.handleRefineStackDeploy)
			deployGroup.POST("/stack/:deploy_id/regenerate", s.handleRegenerateStackDeploy)
			deployGroup.POST("/stack/:deploy_id/reopen", s.handleReopenStackDeploy)
			deployGroup.POST("/stack/:deploy_id/execute", s.handleExecuteStackDeploy)
			deployGroup.POST("/stack/:deploy_id/undeploy", s.handleUndeployStack)
			deployGroup.POST("/stack/:deploy_id/redeploy", s.handleRedeployStack)
			deployGroup.DELETE("/stack/:deploy_id", s.handleDeleteStackDeploy)
		}

		// Config
		configGroup := api.Group("/config")
		{
			configGroup.GET("/clusters", s.handleGetClustersConfig)
			configGroup.GET("/ai", s.handleGetAIConfig)
			configGroup.PUT("/ai", s.handleUpdateAIConfig)
			configGroup.GET("/ai/models", s.handleListAIModels)
			configGroup.GET("/kubecontexts", s.handleListKubeContexts)
			configGroup.POST("/clusters", s.handleRegisterCluster)
			configGroup.DELETE("/clusters/:name", s.handleUnregisterCluster)
		}
	}

	// WebSocket
	ws := r.Group("/ws")
	{
		ws.GET("/docker/stats", s.handleDockerStatsWS)
		ws.GET("/k8s/:cluster/metrics", s.handleK8sMetricsWS)
		ws.GET("/docker/:container_id/logs", s.handleDockerLogsWS)
		ws.GET("/k8s/:cluster/:namespace/:pod/logs", s.handleK8sLogsWS)
		ws.GET("/deploy/:deploy_id/status", s.handleDeployStatusWS)
	}

	s.router = r
}

// Router returns the underlying gin.Engine for the server.
func (s *Server) Router() *gin.Engine {
	return s.router
}
