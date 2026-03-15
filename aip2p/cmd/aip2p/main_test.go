package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aip2p.org/internal/aip2p"
)

func TestRunPublishRejectsReadOnlyIdentityWhenWriterPolicyIsProvided(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store := filepath.Join(root, "store")
	if _, err := aip2p.OpenStore(store); err != nil {
		t.Fatalf("OpenStore error = %v", err)
	}
	identity, err := aip2p.NewAgentIdentity("agent://writer/readonly", "agent://writer/readonly", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity error = %v", err)
	}
	identityPath := filepath.Join(root, "identity.json")
	if err := aip2p.SaveAgentIdentity(identityPath, identity); err != nil {
		t.Fatalf("SaveAgentIdentity error = %v", err)
	}
	policyPath := filepath.Join(root, "writer_policy.json")
	policy := map[string]any{
		"sync_mode":          "trusted_writers_only",
		"allow_unsigned":     false,
		"default_capability": "read_only",
		"agent_capabilities": map[string]string{
			identity.AgentID: "read_only",
		},
	}
	data, err := json.Marshal(policy)
	if err != nil {
		t.Fatalf("Marshal(policy) error = %v", err)
	}
	if err := os.WriteFile(policyPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	err = run([]string{
		"publish",
		"--store", store,
		"--author", identity.Author,
		"--identity-file", identityPath,
		"--writer-policy", policyPath,
		"--title", "Blocked publish",
		"--body", "hello world",
	})
	if err == nil {
		t.Fatal("expected publish to be refused")
	}
	if !strings.Contains(err.Error(), "read_only") {
		t.Fatalf("error = %v, want read_only refusal", err)
	}
}
