package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// Service defines the interface for Kubernetes cluster operations.
type Service interface {
	ListClusters(ctx context.Context) ([]models.Cluster, error)
	ListNamespaces(ctx context.Context, cluster string) ([]string, error)
	ListPods(ctx context.Context, cluster, namespace, labelSelector string) ([]models.Pod, error)
	ListDeployments(ctx context.Context, cluster, namespace string) ([]models.Deployment, error)
	ListServices(ctx context.Context, cluster, namespace string) ([]models.Service, error)
	ScaleDeployment(ctx context.Context, cluster, namespace, name string, replicas int) error
	RestartPod(ctx context.Context, cluster, namespace, name string) error
	DeleteDeployment(ctx context.Context, cluster, namespace, name string) error
	DeleteService(ctx context.Context, cluster, namespace, name string) error

	// Cluster management
	ListKubeContexts(kubeconfigPath string) ([]models.KubeContext, error)
	AddCluster(ctx context.Context, cfg config.ClusterConfig) error
	RemoveCluster(name string) error
}

type clusterClient struct {
	config config.ClusterConfig
	client k8s.Interface
}

type k8sService struct {
	clusters map[string]*clusterClient
	mu       sync.RWMutex
}

// NewService creates a new Kubernetes service with the given cluster configurations.
func NewService(clusters []config.ClusterConfig) (Service, error) {
	svc := &k8sService{
		clusters: make(map[string]*clusterClient),
	}

	for _, cc := range clusters {
		client, err := buildClient(cc)
		if err != nil {
			slog.Warn("failed to create k8s client, marking as disconnected",
				"cluster", cc.Name, "error", err)
			svc.clusters[cc.Name] = &clusterClient{config: cc, client: nil}
			continue
		}
		svc.clusters[cc.Name] = &clusterClient{config: cc, client: client}
	}

	return svc, nil
}

func buildClient(cc config.ClusterConfig) (k8s.Interface, error) {
	kubeconfigPath := cc.Kubeconfig
	if strings.HasPrefix(kubeconfigPath, "~") {
		home, _ := os.UserHomeDir()
		kubeconfigPath = home + kubeconfigPath[1:]
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	overrides := &clientcmd.ConfigOverrides{}
	if cc.Context != "" {
		overrides.CurrentContext = cc.Context
	}

	restCfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("building config: %w", err)
	}

	restCfg.Timeout = 10 * time.Second

	clientset, err := k8s.NewForConfig(restCfg)
	if err != nil {
		return nil, fmt.Errorf("creating clientset: %w", err)
	}

	return clientset, nil
}

func (s *k8sService) getClient(cluster string) (*clusterClient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cc, ok := s.clusters[cluster]
	if !ok {
		return nil, fmt.Errorf("cluster %q not found", cluster)
	}
	if cc.client == nil {
		return nil, fmt.Errorf("cluster %q is disconnected", cluster)
	}
	return cc, nil
}

func (s *k8sService) ListClusters(ctx context.Context) ([]models.Cluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Cluster, 0, len(s.clusters))
	for _, cc := range s.clusters {
		cluster := models.Cluster{
			Name:    cc.config.Name,
			Type:    cc.config.Type,
			Context: cc.config.Context,
			Status:  "disconnected",
		}

		if cc.client != nil {
			ver, err := cc.client.Discovery().ServerVersion()
			if err == nil {
				cluster.Status = "connected"
				cluster.Info.Version = ver.GitVersion

				nodes, err := cc.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
				if err == nil {
					cluster.Info.Nodes = len(nodes.Items)
				}
				pods, err := cc.client.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
				if err == nil {
					cluster.Info.Pods = len(pods.Items)
				}
				nsList, err := cc.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
				if err == nil {
					cluster.Info.Namespaces = len(nsList.Items)
				}
			}
		}

		result = append(result, cluster)
	}
	return result, nil
}

func (s *k8sService) ListNamespaces(ctx context.Context, cluster string) ([]string, error) {
	cc, err := s.getClient(cluster)
	if err != nil {
		return nil, err
	}

	nsList, err := cc.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing namespaces: %w", err)
	}

	result := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		result = append(result, ns.Name)
	}
	return result, nil
}

func (s *k8sService) ListPods(ctx context.Context, cluster, namespace, labelSelector string) ([]models.Pod, error) {
	cc, err := s.getClient(cluster)
	if err != nil {
		return nil, err
	}

	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}

	pods, err := cc.client.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("listing pods: %w", err)
	}

	result := make([]models.Pod, 0, len(pods.Items))
	for _, p := range pods.Items {
		containers := make([]models.PodContainer, 0, len(p.Status.ContainerStatuses))
		for _, cs := range p.Status.ContainerStatuses {
			state := "waiting"
			if cs.State.Running != nil {
				state = "running"
			} else if cs.State.Terminated != nil {
				state = "terminated"
			}
			containers = append(containers, models.PodContainer{
				Name:         cs.Name,
				Image:        cs.Image,
				Ready:        cs.Ready,
				RestartCount: int(cs.RestartCount),
				State:        state,
			})
		}

		var resources models.PodResources
		if len(p.Spec.Containers) > 0 {
			c := p.Spec.Containers[0]
			resources = models.PodResources{
				CPURequest:    c.Resources.Requests.Cpu().String(),
				CPULimit:      c.Resources.Limits.Cpu().String(),
				MemoryRequest: c.Resources.Requests.Memory().String(),
				MemoryLimit:   c.Resources.Limits.Memory().String(),
			}
		}

		conditions := make([]models.Condition, 0, len(p.Status.Conditions))
		for _, cond := range p.Status.Conditions {
			conditions = append(conditions, models.Condition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			})
		}

		result = append(result, models.Pod{
			Name:       p.Name,
			Namespace:  p.Namespace,
			Status:     string(p.Status.Phase),
			Phase:      string(p.Status.Phase),
			Node:       p.Spec.NodeName,
			IP:         p.Status.PodIP,
			CreatedAt:  p.CreationTimestamp.Time,
			Containers: containers,
			Resources:  resources,
			Conditions: conditions,
		})
	}
	return result, nil
}

func (s *k8sService) ListDeployments(ctx context.Context, cluster, namespace string) ([]models.Deployment, error) {
	cc, err := s.getClient(cluster)
	if err != nil {
		return nil, err
	}

	deployments, err := cc.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing deployments: %w", err)
	}

	result := make([]models.Deployment, 0, len(deployments.Items))
	for _, d := range deployments.Items {
		image := ""
		if len(d.Spec.Template.Spec.Containers) > 0 {
			image = d.Spec.Template.Spec.Containers[0].Image
		}

		conditions := make([]models.Condition, 0, len(d.Status.Conditions))
		for _, cond := range d.Status.Conditions {
			conditions = append(conditions, models.Condition{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Reason:  cond.Reason,
				Message: cond.Message,
			})
		}

		selector := make(map[string]string)
		if d.Spec.Selector != nil {
			selector = d.Spec.Selector.MatchLabels
		}

		result = append(result, models.Deployment{
			Name:              d.Name,
			Namespace:         d.Namespace,
			Replicas:          int(*d.Spec.Replicas),
			ReadyReplicas:     int(d.Status.ReadyReplicas),
			AvailableReplicas: int(d.Status.AvailableReplicas),
			UpdatedReplicas:   int(d.Status.UpdatedReplicas),
			Image:             image,
			CreatedAt:         d.CreationTimestamp.Time,
			Conditions:        conditions,
			Selector:          selector,
		})
	}
	return result, nil
}

func (s *k8sService) ListServices(ctx context.Context, cluster, namespace string) ([]models.Service, error) {
	cc, err := s.getClient(cluster)
	if err != nil {
		return nil, err
	}

	services, err := cc.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("listing services: %w", err)
	}

	result := make([]models.Service, 0, len(services.Items))
	for _, svc := range services.Items {
		ports := make([]models.ServicePort, 0, len(svc.Spec.Ports))
		for _, p := range svc.Spec.Ports {
			ports = append(ports, models.ServicePort{
				Name:       p.Name,
				Port:       int(p.Port),
				TargetPort: p.TargetPort.IntValue(),
				Protocol:   string(p.Protocol),
			})
		}

		result = append(result, models.Service{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Selector:  svc.Spec.Selector,
		})
	}
	return result, nil
}

func (s *k8sService) ScaleDeployment(ctx context.Context, cluster, namespace, name string, replicas int) error {
	cc, err := s.getClient(cluster)
	if err != nil {
		return err
	}

	scale, err := cc.client.AppsV1().Deployments(namespace).GetScale(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting deployment scale: %w", err)
	}

	scale.Spec.Replicas = int32(replicas)
	_, err = cc.client.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating deployment scale: %w", err)
	}
	return nil
}

func (s *k8sService) RestartPod(ctx context.Context, cluster, namespace, name string) error {
	cc, err := s.getClient(cluster)
	if err != nil {
		return err
	}

	err = cc.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting pod: %w", err)
	}
	return nil
}

func (s *k8sService) DeleteDeployment(ctx context.Context, cluster, namespace, name string) error {
	cc, err := s.getClient(cluster)
	if err != nil {
		return err
	}
	err = cc.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return nil
}

func (s *k8sService) DeleteService(ctx context.Context, cluster, namespace, name string) error {
	cc, err := s.getClient(cluster)
	if err != nil {
		return err
	}
	err = cc.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	return nil
}

// ListKubeContexts parses the given kubeconfig file and returns all available contexts.
func (s *k8sService) ListKubeContexts(kubeconfigPath string) ([]models.KubeContext, error) {
	if kubeconfigPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home dir: %w", err)
		}
		kubeconfigPath = home + "/.kube/config"
	} else if strings.HasPrefix(kubeconfigPath, "~") {
		home, _ := os.UserHomeDir()
		kubeconfigPath = home + kubeconfigPath[1:]
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath}
	cfgObj, err := loadingRules.Load()
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig: %w", err)
	}

	// Check which contexts are already registered
	s.mu.RLock()
	registered := make(map[string]bool)
	for _, cc := range s.clusters {
		registered[cc.config.Context] = true
	}
	s.mu.RUnlock()

	result := make([]models.KubeContext, 0, len(cfgObj.Contexts))
	for name, ctx := range cfgObj.Contexts {
		result = append(result, models.KubeContext{
			Name:      name,
			Cluster:   ctx.Cluster,
			User:      ctx.AuthInfo,
			Namespace: ctx.Namespace,
			IsActive:  registered[name],
		})
	}
	return result, nil
}

// AddCluster dynamically creates a client and adds it to the cluster map.
func (s *k8sService) AddCluster(ctx context.Context, cfg config.ClusterConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[cfg.Name]; exists {
		return fmt.Errorf("cluster %q already registered", cfg.Name)
	}

	client, err := buildClient(cfg)
	if err != nil {
		slog.Warn("failed to create k8s client for new cluster, marking as disconnected",
			"cluster", cfg.Name, "error", err)
		s.clusters[cfg.Name] = &clusterClient{config: cfg, client: nil}
		return nil
	}

	s.clusters[cfg.Name] = &clusterClient{config: cfg, client: client}
	slog.Info("cluster registered", "name", cfg.Name, "context", cfg.Context)
	return nil
}

// RemoveCluster removes a cluster from the service by name.
func (s *k8sService) RemoveCluster(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.clusters[name]; !exists {
		return fmt.Errorf("cluster %q not found", name)
	}

	delete(s.clusters, name)
	slog.Info("cluster deregistered", "name", name)
	return nil
}
