package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/seyunpark/hybrid_cloud_dashboard/internal/config"
	"github.com/seyunpark/hybrid_cloud_dashboard/pkg/models"
)

// ContainerInfo holds Docker container information used for AI manifest generation.
type ContainerInfo struct {
	Name        string
	Image       string
	ImageTag    string
	EnvVars     map[string]string
	Ports       []int
	Volumes     []string
	Command     []string
	WorkingDir  string
	CPUUsage    string
	MemoryUsage string
	NetworkMode string
}

// StackContainerInfo holds information about multiple containers for stack deployment.
type StackContainerInfo struct {
	StackName  string
	Containers []ContainerInfo
	Namespace  string
}

// StackManifestResult is the AI response for stack manifest generation.
// Manifests maps resource kind (e.g. "Deployment", "Service", "ConfigMap", "Secret")
// to a map of resource name → YAML string.
type StackManifestResult struct {
	Topology   models.StackTopology       `json:"topology"`
	Manifests  map[string]map[string]string `json:"manifests"`
	Reasoning  string                      `json:"reasoning"`
	Confidence float64                     `json:"confidence"`
}

// Service defines the interface for AI-based manifest generation.
type Service interface {
	GenerateManifest(ctx context.Context, info ContainerInfo, history []models.DeploymentHistory) (*models.ManifestResult, error)
	RefineManifest(ctx context.Context, currentManifest *models.ManifestResult, feedback string) (*models.ManifestResult, error)
	GenerateStackManifest(ctx context.Context, info StackContainerInfo, history []models.DeploymentHistory) (*StackManifestResult, error)
	RefineStackManifest(ctx context.Context, current *StackManifestResult, feedback string) (*StackManifestResult, error)
	UpdateConfig(provider, apiKey, model string)
	GetConfig() map[string]interface{}
	ListModels(ctx context.Context, provider, apiKey string) ([]string, error)
}

// NewService creates a new AI service with the given configuration.
func NewService(cfg config.AIConfig) (Service, error) {
	maxTokens := cfg.MaxTokens
	if maxTokens < 4096 {
		maxTokens = 4096
	}
	svc := &aiService{
		provider:    cfg.Provider,
		model:       cfg.Model,
		apiKey:      cfg.APIKey,
		temperature: cfg.Temperature,
		maxTokens:   maxTokens,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
	}

	switch cfg.Provider {
	case "openai":
		svc.endpoint = "https://api.openai.com/v1/chat/completions"
	case "claude":
		svc.endpoint = "https://api.anthropic.com/v1/messages"
	case "gemini":
		svc.endpoint = "https://generativelanguage.googleapis.com/v1beta"
	case "azure-openai":
		svc.endpoint = cfg.APIKey // Azure uses a custom endpoint
	default:
		svc.endpoint = "https://api.openai.com/v1/chat/completions"
	}

	return svc, nil
}

type aiService struct {
	provider    string
	model       string
	apiKey      string
	temperature float64
	maxTokens   int
	endpoint    string
	httpClient  *http.Client
}

// UpdateConfig updates the AI service configuration at runtime.
func (s *aiService) UpdateConfig(provider, apiKey, model string) {
	if provider != "" {
		s.provider = provider
		switch provider {
		case "claude":
			s.endpoint = "https://api.anthropic.com/v1/messages"
		case "openai":
			s.endpoint = "https://api.openai.com/v1/chat/completions"
		case "gemini":
			s.endpoint = "https://generativelanguage.googleapis.com/v1beta"
		}
	}
	if apiKey != "" {
		s.apiKey = apiKey
	}
	if model != "" {
		s.model = model
	}
}

// GetConfig returns the current AI service configuration (api key is masked).
func (s *aiService) GetConfig() map[string]interface{} {
	maskedKey := ""
	if s.apiKey != "" && s.apiKey != "your-api-key-here" {
		if len(s.apiKey) > 8 {
			maskedKey = s.apiKey[:4] + "..." + s.apiKey[len(s.apiKey)-4:]
		} else {
			maskedKey = "****"
		}
	}
	return map[string]interface{}{
		"provider":    s.provider,
		"model":       s.model,
		"api_key":     maskedKey,
		"temperature": s.temperature,
		"configured":  s.apiKey != "" && s.apiKey != "your-api-key-here",
	}
}

func (s *aiService) callProvider(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	switch s.provider {
	case "claude":
		return s.callClaudeAPI(ctx, systemPrompt, userPrompt)
	case "gemini":
		return s.callGeminiAPI(ctx, systemPrompt, userPrompt)
	default:
		return s.callOpenAIAPI(ctx, systemPrompt, userPrompt)
	}
}

// callWithRetry calls the AI provider with one retry on failure.
func (s *aiService) callWithRetry(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	response, err := s.callProvider(ctx, systemPrompt, userPrompt)
	if err == nil {
		return response, nil
	}
	slog.Warn("AI API call failed, retrying once", "provider", s.provider, "error", err)
	time.Sleep(2 * time.Second)
	return s.callProvider(ctx, systemPrompt, userPrompt)
}

func (s *aiService) GenerateManifest(ctx context.Context, info ContainerInfo, history []models.DeploymentHistory) (*models.ManifestResult, error) {
	if s.apiKey == "" || s.apiKey == "your-api-key-here" {
		fb := s.generateFallbackManifest(info)
		fb.Reasoning = "[Fallback] AI API 키가 설정되지 않아 기본 템플릿으로 생성되었습니다. Settings에서 AI 설정을 구성하면 더 정확한 매니페스트를 생성할 수 있습니다."
		return fb, nil
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildUserPrompt(info, history)

	response, err := s.callWithRetry(ctx, systemPrompt, userPrompt)
	if err != nil {
		slog.Error("AI API call failed after retry", "provider", s.provider, "error", err)
		fb := s.generateFallbackManifest(info)
		fb.Reasoning = fmt.Sprintf("[Fallback] AI API 호출 실패: %v. 기본 템플릿으로 생성되었습니다. API 키와 네트워크 연결을 확인해주세요.", err)
		return fb, nil
	}

	result, err := parseManifestResponse(response)
	if err != nil {
		slog.Error("failed to parse AI response", "provider", s.provider, "error", err, "response_len", len(response))
		fb := s.generateFallbackManifest(info)
		fb.Reasoning = fmt.Sprintf("[Fallback] AI 응답 파싱 실패: %v. 기본 템플릿으로 생성되었습니다. 다시 시도해보세요.", err)
		return fb, nil
	}

	return result, nil
}

func (s *aiService) RefineManifest(ctx context.Context, currentManifest *models.ManifestResult, feedback string) (*models.ManifestResult, error) {
	if s.apiKey == "" || s.apiKey == "your-api-key-here" {
		return nil, fmt.Errorf("AI API key not configured. Settings에서 API 키를 설정해주세요.")
	}

	systemPrompt := buildSystemPrompt()
	userPrompt := buildRefinePrompt(currentManifest, feedback)

	response, err := s.callWithRetry(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %w", err)
	}

	result, err := parseManifestResponse(response)
	if err != nil {
		return nil, fmt.Errorf("AI 응답 파싱 실패: %w", err)
	}

	return result, nil
}

func buildRefinePrompt(manifest *models.ManifestResult, feedback string) string {
	var b strings.Builder

	b.WriteString("## 현재 생성된 Manifest\n\n")
	b.WriteString("### Deployment\n```yaml\n")
	b.WriteString(manifest.Deployment)
	b.WriteString("\n```\n\n")

	if manifest.Service != "" {
		b.WriteString("### Service\n```yaml\n")
		b.WriteString(manifest.Service)
		b.WriteString("\n```\n\n")
	}

	if manifest.HPA != "" {
		b.WriteString("### HPA\n```yaml\n")
		b.WriteString(manifest.HPA)
		b.WriteString("\n```\n\n")
	}

	b.WriteString("## 사용자 피드백\n")
	b.WriteString(feedback)
	b.WriteString("\n\n위 피드백을 반영하여 manifest를 수정해주세요. 동일한 JSON 형식으로 응답하세요. reasoning은 한국어로 작성하세요.")

	return b.String()
}

func buildSystemPrompt() string {
	return `당신은 Kubernetes 배포 전문가입니다. Docker 컨테이너 정보와 유사 배포 이력을 바탕으로 최적의 Kubernetes manifest를 생성합니다.

반드시 아래 JSON 구조로만 응답하세요 (앞뒤에 다른 텍스트 없이):
{
  "deployment": "<Deployment YAML 문자열>",
  "service": "<Service YAML 문자열>",
  "hpa": "<HPA YAML 문자열 또는 빈 문자열>",
  "configmap": "<ConfigMap YAML 문자열 또는 빈 문자열>",
  "reasoning": "<선택 이유에 대한 간략한 설명 (한국어로)>",
  "confidence": <0~1 사이의 float>
}

manifest 생성 요구사항:
- 반드시 resource requests와 limits 포함
- 반드시 liveness/readiness probe 포함
- 반드시 SecurityContext 포함 (runAsNonRoot, readOnlyRootFilesystem 등 가능한 경우)
- 서비스 유형은 용도에 맞게 설정 (내부: ClusterIP, 외부: LoadBalancer)
- 서비스 타입에 따라 적절한 replica 수 설정

중요: reasoning 필드는 반드시 한국어로 작성하세요.`
}

func buildUserPrompt(info ContainerInfo, history []models.DeploymentHistory) string {
	var b strings.Builder

	if len(history) > 0 {
		b.WriteString("## Similar Deployment History (for reference)\n")
		for i, h := range history {
			if i >= 3 {
				break
			}
			fmt.Fprintf(&b, "- Service: %s, Image: %s:%s, CPU: %s/%s, Memory: %s/%s, Replicas: %d, Success: %v\n",
				h.ServiceName, h.ImageName, h.ImageTag,
				h.CPURequest, h.CPULimit, h.MemoryRequest, h.MemoryLimit,
				h.Replicas, h.Success)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Current Container Information\n")
	fmt.Fprintf(&b, "- Name: %s\n", info.Name)
	fmt.Fprintf(&b, "- Image: %s:%s\n", info.Image, info.ImageTag)

	if len(info.Ports) > 0 {
		portStrs := make([]string, len(info.Ports))
		for i, p := range info.Ports {
			portStrs[i] = fmt.Sprintf("%d", p)
		}
		fmt.Fprintf(&b, "- Ports: %s\n", strings.Join(portStrs, ", "))
	}

	if len(info.EnvVars) > 0 {
		b.WriteString("- Environment Variables: ")
		envPairs := make([]string, 0, len(info.EnvVars))
		for k := range info.EnvVars {
			envPairs = append(envPairs, k)
		}
		fmt.Fprintf(&b, "%s\n", strings.Join(envPairs, ", "))
	}

	if info.CPUUsage != "" {
		fmt.Fprintf(&b, "- Current CPU Usage: %s\n", info.CPUUsage)
	}
	if info.MemoryUsage != "" {
		fmt.Fprintf(&b, "- Current Memory Usage: %s\n", info.MemoryUsage)
	}

	if len(info.Volumes) > 0 {
		fmt.Fprintf(&b, "- Volumes: %s\n", strings.Join(info.Volumes, ", "))
	}

	if len(info.Command) > 0 {
		fmt.Fprintf(&b, "- Command: %s\n", strings.Join(info.Command, " "))
	}

	b.WriteString("\nGenerate Kubernetes Deployment + Service YAML manifests in the JSON format specified.")
	return b.String()
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *aiService) callOpenAIAPI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	maxTokens := s.maxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	reqBody := openAIRequest{
		Model: s.model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: s.temperature,
		MaxTokens:   maxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

type claudeRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	System      string          `json:"system"`
	Messages    []claudeMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (s *aiService) callClaudeAPI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	maxTokens := s.maxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	reqBody := claudeRequest{
		Model:     s.model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []claudeMessage{
			{Role: "user", Content: userPrompt},
		},
		Temperature: s.temperature,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return claudeResp.Content[0].Text, nil
}

// --- Gemini API ---

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
	GenerationConfig  *geminiGenerationConfig `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	Temperature float64 `json:"temperature"`
	MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason,omitempty"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error,omitempty"`
}

func (s *aiService) callGeminiAPI(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	model := s.model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", s.endpoint, model, s.apiKey)

	maxTokens := s.maxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}
	// Gemini 2.5 "thinking" models need higher token budget since
	// thinking tokens count towards maxOutputTokens
	if strings.Contains(model, "2.5") && maxTokens < 8192 {
		maxTokens = 8192
	}

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{
			Parts: []geminiPart{{Text: systemPrompt}},
		},
		Contents: []geminiContent{
			{
				Role: "user",
				Parts: []geminiPart{{Text: userPrompt}},
			},
		},
		GenerationConfig: &geminiGenerationConfig{
			Temperature: s.temperature,
			MaxOutputTokens: maxTokens,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("calling Gemini API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if geminiResp.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in Gemini response")
	}

	// Concatenate all parts (Gemini may split response across multiple parts)
	var fullText strings.Builder
	for _, part := range geminiResp.Candidates[0].Content.Parts {
		fullText.WriteString(part.Text)
	}

	text := fullText.String()
	slog.Info("Gemini API response received",
		"text_len", len(text),
		"parts_count", len(geminiResp.Candidates[0].Content.Parts),
		"finish_reason", geminiResp.Candidates[0].FinishReason)

	return text, nil
}

// --- List Models ---

func (s *aiService) ListModels(ctx context.Context, provider, apiKey string) ([]string, error) {
	if provider == "" {
		provider = s.provider
	}
	if apiKey == "" {
		apiKey = s.apiKey
	}
	if apiKey == "" || apiKey == "your-api-key-here" {
		return nil, fmt.Errorf("API key is required to list models")
	}

	switch provider {
	case "openai":
		return s.listOpenAIModels(ctx, apiKey)
	case "claude":
		return s.listClaudeModels(ctx, apiKey)
	case "gemini":
		return s.listGeminiModels(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *aiService) listOpenAIModels(ctx context.Context, apiKey string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling OpenAI models API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID      string `json:"id"`
			OwnedBy string `json:"owned_by"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Filter to chat models only
	var models []string
	for _, m := range result.Data {
		if strings.HasPrefix(m.ID, "gpt-") {
			models = append(models, m.ID)
		}
	}
	sort.Strings(models)
	return models, nil
}

func (s *aiService) listClaudeModels(ctx context.Context, apiKey string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Anthropic models API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Anthropic API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Data {
		models = append(models, m.ID)
	}
	sort.Strings(models)
	return models, nil
}

func (s *aiService) listGeminiModels(ctx context.Context, apiKey string) ([]string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Gemini models API: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Models []struct {
			Name                       string   `json:"name"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var models []string
	for _, m := range result.Models {
		// Filter to models that support content generation
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				// "models/gemini-2.0-flash" → "gemini-2.0-flash"
				name := strings.TrimPrefix(m.Name, "models/")
				models = append(models, name)
				break
			}
		}
	}
	sort.Strings(models)
	return models, nil
}

func parseManifestResponse(response string) (*models.ManifestResult, error) {
	// Try parsing as JSON first (clean response)
	var result models.ManifestResult
	trimmed := strings.TrimSpace(response)
	if err := json.Unmarshal([]byte(trimmed), &result); err == nil {
		if err := validateYAML(result.Deployment); err == nil {
			return &result, nil
		} else {
			slog.Debug("direct JSON parsed but YAML invalid", "error", err)
		}
	} else {
		slog.Debug("direct JSON parse failed", "error", err)
	}

	// Try extracting content between code fences (```json ... ``` or ``` ... ```)
	codeBlockRe := regexp.MustCompile("(?s)```(?:json)?\\s*\n(.*?)\n\\s*```")
	codeMatches := codeBlockRe.FindStringSubmatch(response)
	if len(codeMatches) > 1 {
		content := strings.TrimSpace(codeMatches[1])
		if err := json.Unmarshal([]byte(content), &result); err == nil {
			if err := validateYAML(result.Deployment); err == nil {
				return &result, nil
			} else {
				slog.Warn("code block JSON parsed but YAML invalid", "error", err, "deployment_preview", truncate(result.Deployment, 200))
			}
		} else {
			slog.Warn("code block JSON parse failed", "error", err, "content_preview", truncate(content, 200))
		}
	} else {
		slog.Debug("no code block found in response")
	}

	// Try finding JSON by locating first { and last } in the response
	firstBrace := strings.Index(trimmed, "{")
	lastBrace := strings.LastIndex(trimmed, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		jsonCandidate := trimmed[firstBrace : lastBrace+1]
		if err := json.Unmarshal([]byte(jsonCandidate), &result); err == nil {
			if err := validateYAML(result.Deployment); err == nil {
				return &result, nil
			} else {
				slog.Warn("brace extraction JSON parsed but YAML invalid", "error", err, "deployment_preview", truncate(result.Deployment, 200))
			}
		} else {
			slog.Warn("brace extraction JSON parse failed", "error", err)
		}
	}

	// Try extracting YAML blocks
	yamlRe := regexp.MustCompile("(?s)```yaml\\s*\n?(.*?)\n?\\s*```")
	yamlMatches := yamlRe.FindAllStringSubmatch(response, -1)
	if len(yamlMatches) >= 1 {
		result.Deployment = yamlMatches[0][1]
		if len(yamlMatches) >= 2 {
			result.Service = yamlMatches[1][1]
		}
		if len(yamlMatches) >= 3 {
			result.HPA = yamlMatches[2][1]
		}
		result.Reasoning = "Extracted from LLM YAML response"
		result.Confidence = 0.7
		return &result, nil
	}

	slog.Warn("could not parse AI response", "response_len", len(response), "response_preview", truncate(response, 500))
	return nil, fmt.Errorf("could not parse manifest from response")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func validateYAML(content string) error {
	if content == "" {
		return fmt.Errorf("empty YAML content")
	}
	var out interface{}
	return yaml.Unmarshal([]byte(content), &out)
}

func (s *aiService) generateFallbackManifest(info ContainerInfo) *models.ManifestResult {
	name := info.Name
	if name == "" {
		parts := strings.Split(info.Image, "/")
		name = strings.Split(parts[len(parts)-1], ":")[0]
	}

	image := info.Image
	if info.ImageTag != "" && info.ImageTag != "latest" {
		image = info.Image + ":" + info.ImageTag
	}

	containerPort := 80
	if len(info.Ports) > 0 {
		containerPort = info.Ports[0]
	}

	deployment := fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 2
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /
            port: %d
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /
            port: %d
          initialDelaySeconds: 5
          periodSeconds: 10
        securityContext:
          runAsNonRoot: false
          readOnlyRootFilesystem: false`, name, name, name, name, name, image, containerPort, containerPort, containerPort)

	service := fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: %s
  ports:
  - protocol: TCP
    port: %d
    targetPort: %d
  type: ClusterIP`, name, name, containerPort, containerPort)

	return &models.ManifestResult{
		Deployment: deployment,
		Service:    service,
		Reasoning:  "", // caller sets context-specific reasoning
		Confidence: 0.5,
	}
}

// --- Stack Manifest Generation ---

func (s *aiService) GenerateStackManifest(ctx context.Context, info StackContainerInfo, history []models.DeploymentHistory) (*StackManifestResult, error) {
	if s.apiKey == "" || s.apiKey == "your-api-key-here" {
		fb := s.generateFallbackStackManifest(info)
		fb.Reasoning = "[Fallback] AI API 키가 설정되지 않아 기본 템플릿으로 생성되었습니다. Settings에서 AI 설정을 구성하면 서비스 간 연결 자동 감지 등 더 정확한 매니페스트를 생성할 수 있습니다."
		return fb, nil
	}

	// Stack manifests are much larger — temporarily increase token limit.
	// Multiple services with Deployment+Service+ConfigMap+Secret YAML requires a large budget.
	origMaxTokens := s.maxTokens
	if s.maxTokens < 16384 {
		s.maxTokens = 16384
	}
	defer func() { s.maxTokens = origMaxTokens }()

	systemPrompt := buildStackSystemPrompt()
	userPrompt := buildStackUserPrompt(info, history)

	response, err := s.callWithRetry(ctx, systemPrompt, userPrompt)
	if err != nil {
		slog.Error("AI API call failed for stack after retry", "provider", s.provider, "error", err)
		fb := s.generateFallbackStackManifest(info)
		fb.Reasoning = fmt.Sprintf("[Fallback] AI API 호출 실패: %v. 기본 템플릿으로 생성되었습니다. API 키와 네트워크 연결을 확인해주세요.", err)
		return fb, nil
	}

	result, err := parseStackManifestResponse(response)
	if err != nil {
		slog.Error("failed to parse stack AI response", "provider", s.provider, "error", err, "response_len", len(response))
		fb := s.generateFallbackStackManifest(info)
		fb.Reasoning = fmt.Sprintf("[Fallback] AI 응답 파싱 실패: %v. 기본 템플릿으로 생성되었습니다. 다시 시도해보세요.", err)
		return fb, nil
	}

	return result, nil
}

func (s *aiService) RefineStackManifest(ctx context.Context, current *StackManifestResult, feedback string) (*StackManifestResult, error) {
	if s.apiKey == "" || s.apiKey == "your-api-key-here" {
		return nil, fmt.Errorf("AI API key not configured. Settings에서 API 키를 설정해주세요.")
	}

	origMaxTokens := s.maxTokens
	if s.maxTokens < 16384 {
		s.maxTokens = 16384
	}
	defer func() { s.maxTokens = origMaxTokens }()

	systemPrompt := buildStackSystemPrompt()
	userPrompt := buildStackRefinePrompt(current, feedback)

	response, err := s.callWithRetry(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("AI API 호출 실패: %w", err)
	}

	result, err := parseStackManifestResponse(response)
	if err != nil {
		return nil, fmt.Errorf("AI 응답 파싱 실패: %w", err)
	}

	return result, nil
}

func buildStackSystemPrompt() string {
	return `당신은 Kubernetes 멀티 서비스 배포 전문가입니다. 여러 Docker 컨테이너를 분석하여 서로 연결된 Kubernetes 스택 manifest를 생성합니다.

반드시 아래 JSON 구조로만 응답하세요 (앞뒤에 다른 텍스트 없이):
{
  "topology": {
    "services": [{"container_id": "id", "service_name": "name", "service_type": "type", "image": "img"}],
    "connections": [{"from": "svc-a", "to": "svc-b", "port": 5432, "env_var": "DB_HOST"}],
    "deploy_order": ["svc-c", "svc-b", "svc-a"]
  },
  "manifests": {
    "Deployment": {"svc-name": "<Deployment YAML>"},
    "Service": {"svc-name": "<Service YAML>"},
    "ConfigMap": {"svc-name-config": "<ConfigMap YAML>"},
    "Secret": {"svc-name-secret": "<Secret YAML>"}
  },
  "reasoning": "<한국어 설명>",
  "confidence": 0.85
}

manifests 규칙:
- manifests 키는 Kubernetes 리소스 종류(Kind)를 사용 (Deployment, Service, ConfigMap, Secret, HPA, Ingress 등)
- 각 리소스 종류 안에서 리소스 이름을 키로, YAML 문자열을 값으로 매핑
- 필요한 리소스만 포함 (불필요한 빈 객체는 생략)
- 기본적으로 Deployment, Service는 필수
- 컨테이너에 환경변수가 있으면 반드시 ConfigMap으로 분리 (일반 설정)
- 비밀번호, 토큰, 시크릿 키 등 민감한 값은 반드시 Secret으로 분리 (base64 인코딩)
- 필요 시 HPA, Ingress, PersistentVolumeClaim 등 추가 가능

핵심 요구사항:
- 반드시 제공된 컨테이너에 대해서만 K8s 리소스를 생성. 제공되지 않은 서비스(예: DB, Redis 등)는 절대 Deployment/Service를 생성하지 마세요
- 환경변수에 외부 서비스(DB, Redis 등) 연결 정보가 있어도, 해당 서비스가 제공된 컨테이너 목록에 없으면 connections에만 기록하고 리소스는 생성하지 마세요
- 컨테이너의 환경변수(DB_HOST, REDIS_URL, API_URL 등), 포트, 이미지 이름을 분석하여 제공된 서비스 간 연결을 자동 감지
- 연결된 서비스 간에는 K8s DNS 사용: <service-name>.<namespace>.svc.cluster.local
- deploy_order에는 제공된 컨테이너에 대응하는 서비스만 포함 (의존성 순서 기반)
- 각 Deployment에 resource requests/limits, liveness/readiness probe, SecurityContext 포함
- Deployment에서 환경변수는 ConfigMap/Secret을 envFrom 또는 valueFrom으로 참조
- DB 서비스는 ClusterIP, 프론트엔드는 LoadBalancer, 백엔드는 ClusterIP
- 환경변수에서 다른 서비스를 참조하는 값은 K8s DNS로 치환

중요: reasoning 필드는 반드시 한국어로 작성하세요.`
}

func buildStackUserPrompt(info StackContainerInfo, history []models.DeploymentHistory) string {
	var b strings.Builder

	fmt.Fprintf(&b, "## Stack: %s (namespace: %s)\n\n", info.StackName, info.Namespace)

	if len(history) > 0 {
		b.WriteString("## 유사 배포 이력\n")
		for i, h := range history {
			if i >= 3 {
				break
			}
			fmt.Fprintf(&b, "- %s (%s:%s), CPU: %s/%s, Memory: %s/%s, Replicas: %d\n",
				h.ServiceName, h.ImageName, h.ImageTag,
				h.CPURequest, h.CPULimit, h.MemoryRequest, h.MemoryLimit, h.Replicas)
		}
		b.WriteString("\n")
	}

	for i, c := range info.Containers {
		fmt.Fprintf(&b, "## Container %d: %s\n", i+1, c.Name)
		fmt.Fprintf(&b, "- Image: %s:%s\n", c.Image, c.ImageTag)

		if len(c.Ports) > 0 {
			portStrs := make([]string, len(c.Ports))
			for j, p := range c.Ports {
				portStrs[j] = fmt.Sprintf("%d", p)
			}
			fmt.Fprintf(&b, "- Ports: %s\n", strings.Join(portStrs, ", "))
		}

		if len(c.EnvVars) > 0 {
			b.WriteString("- Environment Variables:\n")
			keys := make([]string, 0, len(c.EnvVars))
			for k := range c.EnvVars {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&b, "    %s=%s\n", k, c.EnvVars[k])
			}
		}

		if c.CPUUsage != "" {
			fmt.Fprintf(&b, "- CPU Usage: %s\n", c.CPUUsage)
		}
		if c.MemoryUsage != "" {
			fmt.Fprintf(&b, "- Memory Usage: %s\n", c.MemoryUsage)
		}
		if len(c.Volumes) > 0 {
			fmt.Fprintf(&b, "- Volumes: %s\n", strings.Join(c.Volumes, ", "))
		}
		b.WriteString("\n")
	}

	b.WriteString("위 컨테이너들을 분석하여 서비스 간 연결을 감지하고, 연결된 K8s 스택 manifest를 JSON 형식으로 생성하세요.")
	return b.String()
}

func buildStackRefinePrompt(manifest *StackManifestResult, feedback string) string {
	var b strings.Builder

	b.WriteString("## 현재 생성된 Stack Manifests\n\n")

	b.WriteString("### Topology\n")
	b.WriteString(fmt.Sprintf("- Deploy Order: %s\n", strings.Join(manifest.Topology.DeployOrder, " → ")))
	for _, conn := range manifest.Topology.Connections {
		fmt.Fprintf(&b, "- %s → %s (port: %d, env: %s)\n", conn.From, conn.To, conn.Port, conn.EnvVar)
	}
	b.WriteString("\n")

	for kind, resources := range manifest.Manifests {
		for name, yamlContent := range resources {
			fmt.Fprintf(&b, "### %s: %s\n```yaml\n%s\n```\n\n", kind, name, yamlContent)
		}
	}

	b.WriteString("## 사용자 피드백\n")
	b.WriteString(feedback)
	b.WriteString("\n\n위 피드백을 반영하여 manifest를 수정해주세요. 동일한 JSON 형식으로 응답하세요. reasoning은 한국어로 작성하세요.")

	return b.String()
}

func parseStackManifestResponse(response string) (*StackManifestResult, error) {
	trimmed := strings.TrimSpace(response)

	tryParse := func(data string) *StackManifestResult {
		var result StackManifestResult
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return nil
		}
		if len(result.Manifests) > 0 {
			return &result
		}
		// Fallback: try legacy format with top-level deployments/services keys
		var legacy struct {
			Topology    models.StackTopology       `json:"topology"`
			Deployments map[string]string          `json:"deployments"`
			Services    map[string]string          `json:"services"`
			ConfigMaps  map[string]string          `json:"configmaps"`
			HPAs        map[string]string          `json:"hpas"`
			Secrets     map[string]string          `json:"secrets"`
			Manifests   map[string]map[string]string `json:"manifests"`
			Reasoning   string                     `json:"reasoning"`
			Confidence  float64                    `json:"confidence"`
		}
		if err := json.Unmarshal([]byte(data), &legacy); err != nil || len(legacy.Deployments) == 0 {
			return nil
		}
		// Convert legacy to new format
		m := map[string]map[string]string{}
		if len(legacy.Deployments) > 0 {
			m["Deployment"] = legacy.Deployments
		}
		if len(legacy.Services) > 0 {
			m["Service"] = legacy.Services
		}
		if len(legacy.ConfigMaps) > 0 {
			m["ConfigMap"] = legacy.ConfigMaps
		}
		if len(legacy.Secrets) > 0 {
			m["Secret"] = legacy.Secrets
		}
		if len(legacy.HPAs) > 0 {
			m["HPA"] = legacy.HPAs
		}
		return &StackManifestResult{
			Topology:   legacy.Topology,
			Manifests:  m,
			Reasoning:  legacy.Reasoning,
			Confidence: legacy.Confidence,
		}
	}

	// Method 1: Direct JSON
	if r := tryParse(trimmed); r != nil {
		return r, nil
	}

	// Method 2: Code block extraction
	codeBlockRe := regexp.MustCompile("(?s)```(?:json)?\\s*\n(.*?)\n\\s*```")
	codeMatches := codeBlockRe.FindStringSubmatch(response)
	if len(codeMatches) > 1 {
		if r := tryParse(strings.TrimSpace(codeMatches[1])); r != nil {
			return r, nil
		}
	}

	// Method 3: First/last brace
	firstBrace := strings.Index(trimmed, "{")
	lastBrace := strings.LastIndex(trimmed, "}")
	if firstBrace >= 0 && lastBrace > firstBrace {
		if r := tryParse(trimmed[firstBrace : lastBrace+1]); r != nil {
			return r, nil
		}
	}

	slog.Warn("could not parse stack AI response", "response_len", len(response), "response_preview", truncate(response, 500))
	return nil, fmt.Errorf("could not parse stack manifest from response")
}

func (s *aiService) generateFallbackStackManifest(info StackContainerInfo) *StackManifestResult {
	manifests := map[string]map[string]string{
		"Deployment": {},
		"Service":    {},
		"ConfigMap":  {},
	}
	svcInfos := make([]models.StackServiceInfo, 0, len(info.Containers))
	deployOrder := make([]string, 0, len(info.Containers))

	for _, c := range info.Containers {
		name := c.Name
		if name == "" {
			parts := strings.Split(c.Image, "/")
			name = strings.Split(parts[len(parts)-1], ":")[0]
		}

		image := c.Image
		if c.ImageTag != "" && c.ImageTag != "latest" {
			image = c.Image + ":" + c.ImageTag
		}

		port := 80
		if len(c.Ports) > 0 {
			port = c.Ports[0]
		}

		svcType := detectServiceType(c)
		deployOrder = append(deployOrder, name)
		svcInfos = append(svcInfos, models.StackServiceInfo{
			ServiceName: name,
			ServiceType: svcType,
			Image:       c.Image + ":" + c.ImageTag,
		})

		// ConfigMap for env vars
		configMapName := name + "-config"
		if len(c.EnvVars) > 0 {
			var cmData strings.Builder
			keys := make([]string, 0, len(c.EnvVars))
			for k := range c.EnvVars {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&cmData, "  %s: %q\n", k, c.EnvVars[k])
			}
			manifests["ConfigMap"][configMapName] = fmt.Sprintf(`apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
  labels:
    app: %s
    stack: %s
data:
%s`, configMapName, name, info.StackName, cmData.String())
		}

		envFromBlock := ""
		if len(c.EnvVars) > 0 {
			envFromBlock = fmt.Sprintf(`
        envFrom:
        - configMapRef:
            name: %s`, configMapName)
		}

		manifests["Deployment"][name] = fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
    stack: %s
spec:
  replicas: 1
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
        stack: %s
    spec:
      containers:
      - name: %s
        image: %s
        ports:
        - containerPort: %d%s
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"`, name, name, info.StackName, name, name, info.StackName, name, image, port, envFromBlock)

		k8sSvcType := "ClusterIP"
		if svcType == "web-server" || svcType == "web-application" {
			k8sSvcType = "LoadBalancer"
		}

		manifests["Service"][name] = fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
  labels:
    stack: %s
spec:
  selector:
    app: %s
  ports:
  - protocol: TCP
    port: %d
    targetPort: %d
  type: %s`, name, info.StackName, name, port, port, k8sSvcType)
	}

	// Remove empty resource kinds
	for kind, resources := range manifests {
		if len(resources) == 0 {
			delete(manifests, kind)
		}
	}

	// Sort: databases first, then backends, then frontends
	sort.SliceStable(deployOrder, func(i, j int) bool {
		typeOrder := map[string]int{"database": 0, "message-queue": 1, "application": 2, "web-application": 3, "web-server": 4}
		iType := detectServiceType(info.Containers[i])
		jType := detectServiceType(info.Containers[j])
		return typeOrder[iType] < typeOrder[jType]
	})

	return &StackManifestResult{
		Topology: models.StackTopology{
			Services:    svcInfos,
			Connections: []models.ServiceConnection{},
			DeployOrder: deployOrder,
		},
		Manifests:  manifests,
		Reasoning:   "", // caller sets context-specific reasoning
		Confidence:  0.3,
	}
}

func detectServiceType(info ContainerInfo) string {
	image := strings.ToLower(info.Image)
	switch {
	case strings.Contains(image, "nginx") || strings.Contains(image, "httpd") || strings.Contains(image, "apache"):
		return "web-server"
	case strings.Contains(image, "node") || strings.Contains(image, "express"):
		return "web-application"
	case strings.Contains(image, "postgres") || strings.Contains(image, "mysql") || strings.Contains(image, "mongo") || strings.Contains(image, "redis"):
		return "database"
	case strings.Contains(image, "rabbitmq") || strings.Contains(image, "kafka"):
		return "message-queue"
	default:
		return "application"
	}
}
