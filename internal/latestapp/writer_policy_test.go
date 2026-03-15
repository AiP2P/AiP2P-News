package latestapp

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyWriterPolicyOnlyKeepsReadWriteOrigins(t *testing.T) {
	t.Parallel()

	postAllowed := Bundle{
		InfoHash: "post-allowed",
		Message: Message{
			Kind:   "post",
			Title:  "Allowed",
			Author: "agent://writer/allowed",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/allowed",
				PublicKey: "aaaa",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}
	postReadOnly := Bundle{
		InfoHash: "post-readonly",
		Message: Message{
			Kind:   "post",
			Title:  "Read only",
			Author: "agent://writer/readonly",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/readonly",
				PublicKey: "bbbb",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}
	replyReadOnly := Bundle{
		InfoHash: "reply-readonly",
		Message: Message{
			Kind:   "reply",
			Author: "agent://writer/readonly",
			ReplyTo: &MessageLink{
				InfoHash: "post-allowed",
			},
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/readonly",
				PublicKey: "bbbb",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}

	index := buildIndex([]Bundle{postAllowed, postReadOnly, replyReadOnly}, "aip2p.news")
	policy := WriterPolicy{
		SyncMode:          WriterSyncModeMixed,
		DefaultCapability: WriterCapabilityReadOnly,
		PublicKeyCapabilities: map[string]WriterCapability{
			"aaaa": WriterCapabilityReadWrite,
		},
	}

	filtered := ApplyWriterPolicy(index, "aip2p.news", policy)
	if len(filtered.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(filtered.Posts))
	}
	if filtered.Posts[0].InfoHash != "post-allowed" {
		t.Fatalf("post = %q, want post-allowed", filtered.Posts[0].InfoHash)
	}
	if got := len(filtered.RepliesByPost["post-allowed"]); got != 0 {
		t.Fatalf("reply count = %d, want 0", got)
	}
}

func TestWriterPolicyCapabilityPrefersExplicitMap(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeMixed,
		AllowUnsigned:     false,
		DefaultCapability: WriterCapabilityReadOnly,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/allowed": WriterCapabilityReadWrite,
		},
	}
	allowed := &MessageOrigin{AgentID: "agent://writer/allowed"}
	denied := &MessageOrigin{AgentID: "agent://writer/other"}

	if !policy.allowsOrigin(allowed) {
		t.Fatal("expected explicit read_write writer to be accepted")
	}
	if policy.capabilityForOrigin(denied) != WriterCapabilityReadOnly {
		t.Fatalf("denied capability = %q, want read_only", policy.capabilityForOrigin(denied))
	}
	if policy.acceptsOrigin(denied) {
		t.Fatal("expected read_only writer to be rejected in mixed mode")
	}
}

func TestApplyWriterPolicyWhitelistAcceptsOnlyExplicitWriters(t *testing.T) {
	t.Parallel()

	postAllowed := Bundle{
		InfoHash: "post-allowed",
		Message: Message{
			Kind:   "post",
			Title:  "Allowed",
			Author: "agent://writer/allowed",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/allowed",
				PublicKey: "aaaa",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}
	postOther := Bundle{
		InfoHash: "post-other",
		Message: Message{
			Kind:   "post",
			Title:  "Other",
			Author: "agent://writer/other",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/other",
				PublicKey: "bbbb",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}

	index := buildIndex([]Bundle{postAllowed, postOther}, "aip2p.news")
	policy := WriterPolicy{
		SyncMode:          WriterSyncModeWhitelist,
		DefaultCapability: WriterCapabilityReadWrite,
		AllowedAgentIDs:   []string{"agent://writer/allowed"},
	}

	filtered := ApplyWriterPolicy(index, "aip2p.news", policy)
	if len(filtered.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(filtered.Posts))
	}
	if filtered.Posts[0].InfoHash != "post-allowed" {
		t.Fatalf("post = %q, want post-allowed", filtered.Posts[0].InfoHash)
	}
}

func TestApplyWriterPolicyAllModeKeepsReadOnlyWritersUnlessBlocked(t *testing.T) {
	t.Parallel()

	postReadOnly := Bundle{
		InfoHash: "post-readonly",
		Message: Message{
			Kind:   "post",
			Title:  "Read only",
			Author: "agent://writer/readonly",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/readonly",
				PublicKey: "aaaa",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}
	postBlocked := Bundle{
		InfoHash: "post-blocked",
		Message: Message{
			Kind:   "post",
			Title:  "Blocked",
			Author: "agent://writer/blocked",
			Origin: &MessageOrigin{
				AgentID:   "agent://writer/blocked",
				PublicKey: "bbbb",
			},
			Extensions: map[string]any{"project": "aip2p.news", "topics": []any{"all", "world"}},
		},
	}

	index := buildIndex([]Bundle{postReadOnly, postBlocked}, "aip2p.news")
	policy := WriterPolicy{
		SyncMode:          WriterSyncModeAll,
		DefaultCapability: WriterCapabilityReadOnly,
		BlockedPublicKeys: []string{"bbbb"},
	}

	filtered := ApplyWriterPolicy(index, "aip2p.news", policy)
	if len(filtered.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(filtered.Posts))
	}
	if filtered.Posts[0].InfoHash != "post-readonly" {
		t.Fatalf("post = %q, want post-readonly", filtered.Posts[0].InfoHash)
	}
}

func TestLoadWriterPolicyMergesSharedRegistry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey error = %v", err)
	}
	registry := SignedWriterRegistry{
		AuthorityID: "authority://news/main",
		KeyType:     latestAppKeyTypeEd25519,
		PublicKey:   hex.EncodeToString(publicKey),
		SignedAt:    "2026-03-15T00:00:00Z",
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/shared": WriterCapabilityReadWrite,
		},
		RelayHostTrust: map[string]RelayTrust{
			"mirror.example": RelayTrustBlocked,
		},
	}
	registry.Normalize()
	payload, err := registry.payloadBytes()
	if err != nil {
		t.Fatalf("payloadBytes error = %v", err)
	}
	copyRegistry := registry
	copyRegistry.Signature = hex.EncodeToString(ed25519.Sign(privateKey, payload))
	registryPath := filepath.Join(root, "registry.json")
	data, err := json.MarshalIndent(copyRegistry, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent error = %v", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(registryPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile(registry) error = %v", err)
	}
	_ = payload
	policyPath := filepath.Join(root, "writer_policy.json")
	policyJSON := `{
  "sync_mode": "trusted_writers_only",
  "allow_unsigned": false,
  "default_capability": "read_only",
  "trusted_authorities": {
    "authority://news/main": "` + registry.PublicKey + `"
  },
  "shared_registries": [
    "` + registryPath + `"
  ]
}`
	if err := os.WriteFile(policyPath, []byte(policyJSON), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}
	policy, err := LoadWriterPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadWriterPolicy error = %v", err)
	}
	if !policy.acceptsOrigin(&MessageOrigin{AgentID: "agent://writer/shared"}) {
		t.Fatal("expected shared registry capability to be merged")
	}
	if policy.acceptsRelay("", "mirror.example") {
		t.Fatal("expected relay host from shared registry to be blocked")
	}
}
