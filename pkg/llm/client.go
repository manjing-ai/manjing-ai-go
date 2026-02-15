package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ChatMessage OpenAI兼容的消息结构
type ChatMessage struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest 对话请求
type ChatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
	Temperature    float32       `json:"temperature,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

// ChatResponse 对话响应（OpenAI标准格式）
type ChatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatResult 封装后的调用结果
type ChatResult struct {
	Content          string // 模型返回的文本内容
	PromptTokens     int    // 输入Token数
	CompletionTokens int    // 输出Token数
	TotalTokens      int    // 总Token数
	DurationMs       int    // 调用耗时（毫秒）
}

// ClientConfig 客户端配置
type ClientConfig struct {
	BaseURL     string  // API端点URL
	APIKey      string  // API密钥
	Model       string  // 默认模型
	MaxTokens   int     // 默认最大Token
	Temperature float32 // 默认温度
	Timeout     int     // 超时时间（秒）
}

// Client OpenAI兼容的LLM客户端
type Client struct {
	config     ClientConfig
	httpClient *http.Client
}

// NewClient 创建LLM客户端
func NewClient(cfg ClientConfig) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60
	}
	if cfg.MaxTokens <= 0 {
		cfg.MaxTokens = 4096
	}
	if cfg.Temperature <= 0 {
		cfg.Temperature = 0.7
	}
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// ChatCompletion 发送对话请求
func (c *Client) ChatCompletion(ctx context.Context, messages []ChatMessage, opts ...ChatOption) (*ChatResult, error) {
	if len(messages) == 0 {
		return nil, errors.New("messages不能为空")
	}

	opt := c.defaultOptions()
	for _, o := range opts {
		o(&opt)
	}

	reqBody := ChatRequest{
		Model:       opt.Model,
		Messages:    messages,
		MaxTokens:   opt.MaxTokens,
		Temperature: opt.Temperature,
	}
	if opt.JSONMode {
		reqBody.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建URL
	baseURL := strings.TrimRight(opt.BaseURL, "/")
	url := baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+opt.APIKey)

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	durationMs := int(time.Since(start).Milliseconds())

	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("调用超时")
		}
		return nil, fmt.Errorf("调用失败: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("模型限流")
	}
	if resp.StatusCode != http.StatusOK {
		log.Errorf("LLM调用失败 status=%d body=%s", resp.StatusCode, string(respBytes))
		return nil, fmt.Errorf("模型调用失败: HTTP %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	content := ""
	if len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	}

	return &ChatResult{
		Content:          content,
		PromptTokens:     chatResp.Usage.PromptTokens,
		CompletionTokens: chatResp.Usage.CompletionTokens,
		TotalTokens:      chatResp.Usage.TotalTokens,
		DurationMs:       durationMs,
	}, nil
}

// ChatOption 调用选项
type ChatOption func(*chatOptions)

type chatOptions struct {
	BaseURL     string
	APIKey      string
	Model       string
	MaxTokens   int
	Temperature float32
	JSONMode    bool
}

func (c *Client) defaultOptions() chatOptions {
	return chatOptions{
		BaseURL:     c.config.BaseURL,
		APIKey:      c.config.APIKey,
		Model:       c.config.Model,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}
}

// WithModel 指定模型
func WithModel(model string) ChatOption {
	return func(o *chatOptions) { o.Model = model }
}

// WithMaxTokens 指定最大Token
func WithMaxTokens(n int) ChatOption {
	return func(o *chatOptions) { o.MaxTokens = n }
}

// WithTemperature 指定温度
func WithTemperature(t float32) ChatOption {
	return func(o *chatOptions) { o.Temperature = t }
}

// WithJSONMode 启用JSON模式
func WithJSONMode() ChatOption {
	return func(o *chatOptions) { o.JSONMode = true }
}

// WithEndpoint 覆盖BaseURL和APIKey
func WithEndpoint(baseURL, apiKey string) ChatOption {
	return func(o *chatOptions) {
		o.BaseURL = baseURL
		o.APIKey = apiKey
	}
}
