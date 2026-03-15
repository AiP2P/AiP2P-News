package aip2p

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadWriterPolicyMergesSignedRegistryAndLocalOverrides(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	identity, err := NewAgentIdentity("authority://news/main", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity error = %v", err)
	}
	registry, err := SignWriterRegistry(identity, SignedWriterRegistry{
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/shared": WriterCapabilityReadWrite,
		},
		RelayHostTrust: map[string]RelayTrust{
			"mirror.example": RelayTrustBlocked,
		},
	})
	if err != nil {
		t.Fatalf("SignWriterRegistry error = %v", err)
	}
	registryPath := filepath.Join(root, "shared-registry.json")
	registryData, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent(registry) error = %v", err)
	}
	registryData = append(registryData, '\n')
	if err := os.WriteFile(registryPath, registryData, 0o644); err != nil {
		t.Fatalf("WriteFile(registry) error = %v", err)
	}

	policyPath := filepath.Join(root, "writer_policy.json")
	policyData := `{
  "sync_mode": "trusted_writers_only",
  "allow_unsigned": false,
  "default_capability": "read_only",
  "trusted_authorities": {
    "authority://news/main": "` + identity.PublicKey + `"
  },
  "shared_registries": [
    "` + registryPath + `"
  ],
  "agent_capabilities": {
    "agent://writer/local": "read_write"
  },
  "relay_host_trust": {
    "local.example": "trusted"
  }
}`
	if err := os.WriteFile(policyPath, []byte(policyData), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}

	policy, err := LoadWriterPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadWriterPolicy error = %v", err)
	}
	if !policy.AcceptsOrigin(&MessageOrigin{AgentID: "agent://writer/shared"}) {
		t.Fatal("expected shared registry writer to be accepted")
	}
	if !policy.AcceptsOrigin(&MessageOrigin{AgentID: "agent://writer/local"}) {
		t.Fatal("expected local writer override to be accepted")
	}
	if policy.AcceptsRelay("", "mirror.example") {
		t.Fatal("expected blocked relay host from shared registry to be rejected")
	}
	if !policy.AcceptsRelay("", "local.example") {
		t.Fatal("expected local relay host trust override to be accepted")
	}
}
