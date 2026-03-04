package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

func TestNewService(t *testing.T) {
	svc, err := NewService(config.AIConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		APIKey:      "test-key",
		Temperature: 0.3,
		MaxTokens:   2000,
	})
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	prompt := buildSystemPrompt()
	if prompt == "" {
		t.Fatal("expected non-empty system prompt")
	}

	// Verify key elements are present
	keywords := []string{"Kubernetes", "deployment", "JSON", "resource", "SecurityContext"}
	for _, kw := range keywords {
		if !containsStr(prompt, kw) {
			t.Errorf("system prompt should contain %q", kw)
		}
	}
}

func TestBuildUserPrompt_Basic(t *testing.T) {
	info := ContainerInfo{
		Name:     "my-app",
		Image:    "myorg/myapp",
		ImageTag: "v1.0",
		Ports:    []int{8080, 443},
		EnvVars:  map[string]string{"NODE_ENV": "production"},
	}

	prompt := buildUserPrompt(info, nil)
	if prompt == "" {
		t.Fatal("expected non-empty user prompt")
	}

	if !containsStr(prompt, "my-app") {
		t.Error("user prompt should contain container name")
	}
	if !containsStr(prompt, "myorg/myapp:v1.0") {
		t.Error("user prompt should contain image:tag")
	}
	if !containsStr(prompt, "8080") {
		t.Error("user prompt should contain port")
	}
}

func TestBuildUserPrompt_WithHistory(t *testing.T) {
	info := ContainerInfo{
		Name:  "app",
		Image: "app",
	}

	history := []models.DeploymentHistory{
		{ServiceName: "old-app", ImageName: "app", ImageTag: "v0.9", CPURequest: "200m", CPULimit: "1", MemoryRequest: "256Mi", MemoryLimit: "1Gi", Replicas: 3, Success: true},
	}

	prompt := buildUserPrompt(info, history)

	if !containsStr(prompt, "Similar Deployment History") {
		t.Error("user prompt should include history section")
	}
	if !containsStr(prompt, "old-app") {
		t.Error("user prompt should include historical service name")
	}
}

func TestBuildUserPrompt_WithResourceUsage(t *testing.T) {
	info := ContainerInfo{
		Name:        "app",
		Image:       "app",
		CPUUsage:    "25.3%",
		MemoryUsage: "128Mi",
		Volumes:     []string{"/data", "/logs"},
		Command:     []string{"node", "server.js"},
	}

	prompt := buildUserPrompt(info, nil)

	if !containsStr(prompt, "25.3%") {
		t.Error("prompt should contain CPU usage")
	}
	if !containsStr(prompt, "128Mi") {
		t.Error("prompt should contain memory usage")
	}
	if !containsStr(prompt, "/data") {
		t.Error("prompt should contain volumes")
	}
	if !containsStr(prompt, "node server.js") {
		t.Error("prompt should contain command")
	}
}

func TestParseManifestResponse_ValidJSON(t *testing.T) {
	response := `{
		"deployment": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test",
		"service": "apiVersion: v1\nkind: Service\nmetadata:\n  name: test",
		"hpa": "",
		"configmap": "",
		"reasoning": "Based on the container info...",
		"confidence": 0.85
	}`

	result, err := parseManifestResponse(response)
	if err != nil {
		t.Fatalf("parseManifestResponse failed: %v", err)
	}

	if result.Deployment == "" {
		t.Error("expected non-empty deployment")
	}
	if result.Service == "" {
		t.Error("expected non-empty service")
	}
	if result.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", result.Confidence)
	}
	if result.Reasoning == "" {
		t.Error("expected non-empty reasoning")
	}
}

func TestParseManifestResponse_JSONInCodeBlock(t *testing.T) {
	response := "Here's the manifest:\n```json\n" + `{
		"deployment": "apiVersion: apps/v1\nkind: Deployment",
		"service": "apiVersion: v1\nkind: Service",
		"reasoning": "test",
		"confidence": 0.9
	}` + "\n```"

	result, err := parseManifestResponse(response)
	if err != nil {
		t.Fatalf("parseManifestResponse with code block failed: %v", err)
	}

	if result.Deployment == "" {
		t.Error("expected non-empty deployment from code block")
	}
}

func TestParseManifestResponse_YAMLBlocks(t *testing.T) {
	response := "```yaml\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test\n```\n\n```yaml\napiVersion: v1\nkind: Service\nmetadata:\n  name: test\n```"

	result, err := parseManifestResponse(response)
	if err != nil {
		t.Fatalf("parseManifestResponse with YAML blocks failed: %v", err)
	}

	if result.Deployment == "" {
		t.Error("expected non-empty deployment from YAML block")
	}
	if result.Service == "" {
		t.Error("expected non-empty service from YAML block")
	}
	if result.Confidence != 0.7 {
		t.Errorf("expected default confidence 0.7, got %f", result.Confidence)
	}
}

func TestParseManifestResponse_InvalidResponse(t *testing.T) {
	response := "I can't generate manifests for this container."

	_, err := parseManifestResponse(response)
	if err == nil {
		t.Fatal("expected error for invalid response")
	}
}

func TestValidateYAML_Valid(t *testing.T) {
	validYAML := "apiVersion: v1\nkind: Service\nmetadata:\n  name: test"
	if err := validateYAML(validYAML); err != nil {
		t.Fatalf("expected valid YAML: %v", err)
	}
}

func TestValidateYAML_Empty(t *testing.T) {
	if err := validateYAML(""); err == nil {
		t.Fatal("expected error for empty YAML")
	}
}

func TestValidateYAML_Invalid(t *testing.T) {
	invalidYAML := "{{invalid yaml"
	if err := validateYAML(invalidYAML); err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestGenerateFallbackManifest(t *testing.T) {
	svc := &aiService{}

	tests := []struct {
		name     string
		info     ContainerInfo
		wantName string
		wantPort int
	}{
		{
			name:     "with name and ports",
			info:     ContainerInfo{Name: "my-app", Image: "myorg/myapp", ImageTag: "v1", Ports: []int{3000}},
			wantName: "my-app",
			wantPort: 3000,
		},
		{
			name:     "without name - derives from image",
			info:     ContainerInfo{Image: "myorg/nginx", ImageTag: "latest"},
			wantName: "nginx",
			wantPort: 80,
		},
		{
			name:     "no ports - defaults to 80",
			info:     ContainerInfo{Name: "app", Image: "app"},
			wantPort: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.generateFallbackManifest(tt.info)

			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if result.Deployment == "" {
				t.Error("expected non-empty deployment")
			}
			if result.Service == "" {
				t.Error("expected non-empty service")
			}
			if result.Confidence != 0.5 {
				t.Errorf("expected fallback confidence 0.5, got %f", result.Confidence)
			}
			if !containsStr(result.Deployment, tt.wantName) {
				t.Errorf("deployment should contain name %q", tt.wantName)
			}
			// generateFallbackManifest sets empty reasoning; caller adds context
			if result.Reasoning != "" {
				t.Errorf("expected empty reasoning from fallback generator, got %q", result.Reasoning)
			}
		})
	}
}

func TestGenerateManifest_FallbackWhenNoAPIKey(t *testing.T) {
	svc, err := NewService(config.AIConfig{
		Provider: "openai",
		APIKey:   "",
		Model:    "gpt-4",
	})
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	info := ContainerInfo{
		Name:  "test-app",
		Image: "test-app",
		Ports: []int{8080},
	}

	result, err := svc.GenerateManifest(context.Background(), info, nil)
	if err != nil {
		t.Fatalf("GenerateManifest should not fail with fallback: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil fallback result")
	}
	if result.Deployment == "" {
		t.Error("expected non-empty fallback deployment")
	}
}

func TestGenerateManifest_FallbackWhenPlaceholderKey(t *testing.T) {
	svc, err := NewService(config.AIConfig{
		Provider: "openai",
		APIKey:   "your-api-key-here",
		Model:    "gpt-4",
	})
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	result, err := svc.GenerateManifest(context.Background(), ContainerInfo{Name: "app", Image: "app"}, nil)
	if err != nil {
		t.Fatalf("GenerateManifest should use fallback: %v", err)
	}
	if result == nil {
		t.Fatal("expected fallback result")
	}
}

func TestCallOpenAIAPI_Success(t *testing.T) {
	// Mock OpenAI server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("expected Authorization header")
		}

		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"deployment":"apiVersion: apps/v1\nkind: Deployment","service":"apiVersion: v1\nkind: Service","reasoning":"test","confidence":0.9}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := &aiService{
		provider:    "openai",
		model:       "gpt-4",
		apiKey:      "test-key",
		temperature: 0.3,
		maxTokens:   2000,
		endpoint:    server.URL,
		httpClient:  server.Client(),
	}

	response, err := svc.callOpenAIAPI(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("callOpenAIAPI failed: %v", err)
	}
	if response == "" {
		t.Error("expected non-empty response")
	}
}

func TestCallOpenAIAPI_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	svc := &aiService{
		provider:   "openai",
		apiKey:     "test-key",
		endpoint:   server.URL,
		httpClient: server.Client(),
	}

	_, err := svc.callOpenAIAPI(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestCallClaudeAPI_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Error("expected x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Error("expected anthropic-version header")
		}

		resp := claudeResponse{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: `{"deployment":"apiVersion: apps/v1","service":"apiVersion: v1","reasoning":"test","confidence":0.8}`},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := &aiService{
		provider:   "claude",
		model:      "claude-3-sonnet",
		apiKey:     "test-key",
		endpoint:   server.URL,
		httpClient: server.Client(),
	}

	response, err := svc.callClaudeAPI(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("callClaudeAPI failed: %v", err)
	}
	if response == "" {
		t.Error("expected non-empty response")
	}
}

func TestGenerateManifest_WithMockOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `{"deployment":"apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: test","service":"apiVersion: v1\nkind: Service\nmetadata:\n  name: test","reasoning":"Generated for test","confidence":0.92}`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	svc := &aiService{
		provider:    "openai",
		model:       "gpt-4",
		apiKey:      "real-key",
		temperature: 0.3,
		maxTokens:   2000,
		endpoint:    server.URL,
		httpClient:  server.Client(),
	}

	info := ContainerInfo{
		Name:  "web-app",
		Image: "nginx",
		Ports: []int{80},
	}

	result, err := svc.GenerateManifest(context.Background(), info, nil)
	if err != nil {
		t.Fatalf("GenerateManifest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Confidence != 0.92 {
		t.Errorf("expected confidence 0.92, got %f", result.Confidence)
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
