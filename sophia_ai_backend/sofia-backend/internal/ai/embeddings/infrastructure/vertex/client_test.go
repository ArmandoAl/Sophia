package vertex

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestEmbedTextUsesSemanticSimilarityTask(t *testing.T) {
	client := &Client{
		endpoint: "https://vertex.test/predict",
		httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), `"task_type":"SEMANTIC_SIMILARITY"`) {
				t.Fatalf("unexpected request body: %s", body)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader(`{"predictions":[{"embeddings":{"values":[0.25,0.75]}}]}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}

	embedding, err := client.EmbedText(context.Background(), "meeting")
	if err != nil {
		t.Fatal(err)
	}
	if len(embedding) != 2 || embedding[0] != 0.25 || embedding[1] != 0.75 {
		t.Fatalf("unexpected embedding: %v", embedding)
	}
}
