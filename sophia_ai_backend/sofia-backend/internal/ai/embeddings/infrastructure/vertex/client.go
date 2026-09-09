package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/transport"
)

const (
	cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
	vertexLocation     = "us-central1"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
}

func NewClient(ctx context.Context, project, model string) (*Client, error) {
	project = strings.TrimSpace(project)
	model = strings.TrimSpace(model)
	if project == "" {
		return nil, errors.New("Google Cloud project is required")
	}
	if model == "" {
		return nil, errors.New("Vertex embedding model is required")
	}
	httpClient, _, err := transport.NewHTTPClient(ctx, option.WithScopes(cloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("create Vertex AI client: %w", err)
	}
	endpoint := fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		vertexLocation,
		url.PathEscape(project),
		vertexLocation,
		url.PathEscape(model),
	)
	return &Client{httpClient: httpClient, endpoint: endpoint}, nil
}

func (c *Client) EmbedText(ctx context.Context, text string) ([]float32, error) {
	payload := struct {
		Instances []struct {
			Content  string `json:"content"`
			TaskType string `json:"task_type"`
		} `json:"instances"`
	}{Instances: []struct {
		Content  string `json:"content"`
		TaskType string `json:"task_type"`
	}{{Content: text, TaskType: "SEMANTIC_SIMILARITY"}}}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Vertex embedding request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("Vertex embedding request failed (%s): %s", resp.Status, strings.TrimSpace(string(message)))
	}
	var decoded struct {
		Predictions []struct {
			Embeddings struct {
				Values []float64 `json:"values"`
			} `json:"embeddings"`
		} `json:"predictions"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode Vertex embedding response: %w", err)
	}
	if len(decoded.Predictions) == 0 || len(decoded.Predictions[0].Embeddings.Values) == 0 {
		return nil, errors.New("Vertex embedding response was empty")
	}
	values := decoded.Predictions[0].Embeddings.Values
	embedding := make([]float32, len(values))
	for i, value := range values {
		embedding[i] = float32(value)
	}
	return embedding, nil
}
