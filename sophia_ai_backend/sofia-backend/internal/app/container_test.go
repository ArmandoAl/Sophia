package app

import (
	"testing"

	runtimeinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure"
	deepseekinfra "github.com/armandoalvarado/sofia-backend/internal/ai/runtime/infrastructure/deepseek"
	"github.com/armandoalvarado/sofia-backend/internal/config"
)

func TestBuildUsesMemoryRepositories(t *testing.T) {
	cfg := config.Config{
		Env:                   "test",
		Port:                  "8080",
		JWTSecret:             "test-secret",
		RequestBodyLimitBytes: 1 << 20,
		PersistenceDriver:     "memory",
		AIModelProvider:       "fake",
		AIRuntimeEnabled:      true,
		AIRuntimeProposalOnly: true,
	}

	container, err := Build(cfg)
	if err != nil {
		t.Fatalf("Build returned error: %v", err)
	}
	defer container.Close()
	if container.Server == nil {
		t.Fatal("expected server")
	}
	if container.FirestoreStatus != "disabled" {
		t.Fatalf("expected disabled firestore, got %q", container.FirestoreStatus)
	}
}

func TestBuildModelClientDefaultsToFake(t *testing.T) {
	client, err := BuildModelClient(config.Config{AIModelProvider: "fake", AIRuntimeEnabled: true})
	if err != nil {
		t.Fatalf("BuildModelClient returned error: %v", err)
	}
	if _, ok := client.(*runtimeinfra.FakeModelClient); !ok {
		t.Fatalf("expected fake model client, got %T", client)
	}
}

func TestBuildModelClientSupportsDeepSeek(t *testing.T) {
	client, err := BuildModelClient(config.Config{
		AIModelProvider:  "deepseek",
		AIRuntimeEnabled: true,
		DeepSeekAPIKey:   "test-key",
		DeepSeekModels:   map[string]string{"": "deepseek-test"},
		DeepSeekBaseURL:  "https://deepseek.test",
	})
	if err != nil {
		t.Fatalf("BuildModelClient returned error: %v", err)
	}
	if _, ok := client.(*deepseekinfra.Client); !ok {
		t.Fatalf("expected DeepSeek client, got %T", client)
	}
}
