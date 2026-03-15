package aip2p

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriterPolicyDefaultCapabilityReadOnlyRequiresExplicitWriter(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeMixed,
		AllowUnsigned:     false,
		DefaultCapability: WriterCapabilityReadOnly,
		PublicKeyCapabilities: map[string]WriterCapability{
			"abcd": WriterCapabilityReadWrite,
		},
	}
	allowed := &MessageOrigin{AgentID: "agent://writer/allowed", PublicKey: "abcd"}
	denied := &MessageOrigin{AgentID: "agent://writer/denied", PublicKey: "efgh"}

	if !policy.AllowsOrigin(allowed) {
		t.Fatal("expected explicitly trusted writer to be accepted")
	}
	if policy.CapabilityForOrigin(denied) != WriterCapabilityReadOnly {
		t.Fatalf("denied capability = %q, want read_only", policy.CapabilityForOrigin(denied))
	}
	if policy.AcceptsOrigin(denied) {
		t.Fatal("expected read_only writer to be rejected in mixed mode")
	}
}

func TestWriterPolicyLegacyAllowListStillWhitelists(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:        WriterSyncModeWhitelist,
		AllowUnsigned:   false,
		AllowedAgentIDs: []string{"agent://writer/allowed"},
	}
	allowed := &MessageOrigin{AgentID: "agent://writer/allowed", PublicKey: "allowed-key"}
	denied := &MessageOrigin{AgentID: "agent://writer/other", PublicKey: "other-key"}

	if !policy.AcceptsOrigin(allowed) {
		t.Fatal("expected legacy allow-list writer to be accepted")
	}
	if policy.CapabilityForOrigin(denied) != WriterCapabilityReadOnly {
		t.Fatalf("denied capability = %q, want read_only", policy.CapabilityForOrigin(denied))
	}
	if policy.AcceptsOrigin(denied) {
		t.Fatal("expected non-whitelisted writer to be rejected")
	}
}

func TestWriterPolicyBlockedListsOverrideExplicitWriteCapability(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		DefaultCapability: WriterCapabilityReadOnly,
		PublicKeyCapabilities: map[string]WriterCapability{
			"abcd": WriterCapabilityReadWrite,
		},
		BlockedPublicKeys: []string{"abcd"},
	}
	origin := &MessageOrigin{AgentID: "agent://writer/blocked", PublicKey: "abcd"}

	if policy.CapabilityForOrigin(origin) != WriterCapabilityBlocked {
		t.Fatalf("capability = %q, want blocked", policy.CapabilityForOrigin(origin))
	}
	if policy.AllowsOrigin(origin) {
		t.Fatal("expected blocked writer to be rejected")
	}
}

func TestWriterPolicyUnsignedUsesAllowUnsignedAndDefaultCapability(t *testing.T) {
	t.Parallel()

	closed := WriterPolicy{SyncMode: WriterSyncModeAll, AllowUnsigned: false, DefaultCapability: WriterCapabilityReadWrite}
	if closed.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be rejected by default")
	}

	open := WriterPolicy{SyncMode: WriterSyncModeAll, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadWrite}
	if !open.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be accepted when explicitly allowed")
	}

	restricted := WriterPolicy{SyncMode: WriterSyncModeMixed, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadOnly}
	if restricted.CapabilityForOrigin(nil) != WriterCapabilityReadOnly {
		t.Fatalf("unsigned capability = %q, want read_only", restricted.CapabilityForOrigin(nil))
	}
	if !restricted.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be accepted in mixed mode when explicitly allowed")
	}

	trustedOnly := WriterPolicy{SyncMode: WriterSyncModeTrustedWritersOnly, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadWrite}
	if !trustedOnly.AcceptsOrigin(nil) {
		t.Fatal("expected allow_unsigned to override sync mode for unsigned content")
	}
}

func TestWriterPolicySyncModeAllAcceptsSignedReadOnlyWritersButStillBlocksBlockedOnes(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeAll,
		DefaultCapability: WriterCapabilityReadOnly,
	}
	readOnly := &MessageOrigin{AgentID: "agent://writer/readonly", PublicKey: "aaaa"}
	blocked := &MessageOrigin{AgentID: "agent://writer/blocked", PublicKey: "bbbb"}
	policy.BlockedPublicKeys = []string{"bbbb"}

	if !policy.AcceptsOrigin(readOnly) {
		t.Fatal("expected all mode to accept read_only writers")
	}
	if policy.AcceptsOrigin(blocked) {
		t.Fatal("expected blocked writer to be rejected even in all mode")
	}
}

func TestWriterPolicyOriginWithoutPublicKeyCountsAsUnsigned(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeAll,
		AllowUnsigned:     false,
		DefaultCapability: WriterCapabilityReadWrite,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/legacy": WriterCapabilityReadWrite,
		},
	}
	legacy := &MessageOrigin{AgentID: "agent://writer/legacy"}

	if policy.CapabilityForOrigin(legacy) != WriterCapabilityBlocked {
		t.Fatalf("capability = %q, want blocked", policy.CapabilityForOrigin(legacy))
	}
	if policy.AcceptsOrigin(legacy) {
		t.Fatal("expected origin without public key to be treated as unsigned and rejected")
	}
}

func TestWriterPolicyTrustedWritersOnlyRequiresReadWriteCapability(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeTrustedWritersOnly,
		DefaultCapability: WriterCapabilityReadOnly,
		PublicKeyCapabilities: map[string]WriterCapability{
			"trusted-key": WriterCapabilityReadWrite,
		},
	}
	trusted := &MessageOrigin{AgentID: "agent://writer/trusted", PublicKey: "trusted-key"}
	untrusted := &MessageOrigin{AgentID: "agent://writer/readonly", PublicKey: "readonly-key"}

	if !policy.AcceptsOrigin(trusted) {
		t.Fatal("expected trusted writer to be accepted")
	}
	if policy.AcceptsOrigin(untrusted) {
		t.Fatal("expected untrusted writer to be rejected")
	}
}

func TestWriterPolicyRelayTrustRejectsBlockedPeersAndHosts(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		RelayDefaultTrust: RelayTrustNeutral,
		RelayPeerTrust: map[string]RelayTrust{
			"12D3BlockedPeer": RelayTrustBlocked,
		},
		RelayHostTrust: map[string]RelayTrust{
			"mirror.example": RelayTrustBlocked,
		},
	}
	if policy.AcceptsRelay("12D3BlockedPeer", "") {
		t.Fatal("expected blocked relay peer to be rejected")
	}
	if policy.AcceptsRelay("", "mirror.example") {
		t.Fatal("expected blocked relay host to be rejected")
	}
	if !policy.AcceptsRelay("12D3TrustedPeer", "trusted.example") {
		t.Fatal("expected neutral relay to be accepted")
	}
}

func TestWriterPolicyDelegatedChildInheritsParentReadWrite(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeTrustedWritersOnly,
		DefaultCapability: WriterCapabilityReadOnly,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/parent": WriterCapabilityReadWrite,
		},
	}
	child := &MessageOrigin{
		AgentID:   "agent://writer/child",
		PublicKey: "child-key",
	}
	store := DelegationStore{
		Delegations: []WriterDelegation{
			{
				ParentAgentID:   "agent://writer/parent",
				ParentPublicKey: "parent-key",
				ChildAgentID:    "agent://writer/child",
				ChildPublicKey:  "child-key",
				Scopes:          []string{"post"},
				CreatedAt:       "2024-03-15T12:00:00Z",
			},
		},
	}

	decision := policy.OriginDecision(child, "post", store)
	if decision.Capability != WriterCapabilityReadWrite {
		t.Fatalf("capability = %q, want read_write", decision.Capability)
	}
	if decision.Delegation == nil || decision.Delegation.ParentAgentID != "agent://writer/parent" {
		t.Fatal("expected parent delegation metadata to be attached")
	}
	if !policy.AcceptsOriginWithDelegation(child, "post", store) {
		t.Fatal("expected delegated child to be accepted")
	}
}

func TestWriterPolicyExplicitChildRestrictionOverridesParentDelegation(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeTrustedWritersOnly,
		DefaultCapability: WriterCapabilityReadOnly,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/parent": WriterCapabilityReadWrite,
			"agent://writer/child":  WriterCapabilityReadOnly,
		},
	}
	child := &MessageOrigin{
		AgentID:   "agent://writer/child",
		PublicKey: "child-key",
	}
	store := DelegationStore{
		Delegations: []WriterDelegation{
			{
				ParentAgentID:   "agent://writer/parent",
				ParentPublicKey: "parent-key",
				ChildAgentID:    "agent://writer/child",
				ChildPublicKey:  "child-key",
				Scopes:          []string{"post"},
				CreatedAt:       "2024-03-15T12:00:00Z",
			},
		},
	}

	if policy.AcceptsOriginWithDelegation(child, "post", store) {
		t.Fatal("expected explicit child read_only to override parent delegation")
	}
}

func TestWriterPolicyRevokedDelegationStopsGrantingWrite(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeTrustedWritersOnly,
		DefaultCapability: WriterCapabilityReadOnly,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/parent": WriterCapabilityReadWrite,
		},
	}
	child := &MessageOrigin{
		AgentID:   "agent://writer/child",
		PublicKey: "child-key",
	}
	store := DelegationStore{
		Delegations: []WriterDelegation{
			{
				ParentAgentID:   "agent://writer/parent",
				ParentPublicKey: "parent-key",
				ChildAgentID:    "agent://writer/child",
				ChildPublicKey:  "child-key",
				Scopes:          []string{"post"},
				CreatedAt:       "2024-03-15T12:00:00Z",
			},
		},
		Revocations: []WriterRevocation{
			{
				ParentAgentID:   "agent://writer/parent",
				ParentPublicKey: "parent-key",
				ChildAgentID:    "agent://writer/child",
				ChildPublicKey:  "child-key",
				CreatedAt:       "2024-03-15T12:30:00Z",
			},
		},
	}

	now := time.Date(2024, 3, 15, 13, 0, 0, 0, time.UTC)
	if _, ok := store.ActiveDelegationFor(child.AgentID, child.PublicKey, "post", now); ok {
		t.Fatal("expected revoked delegation to be inactive")
	}
	if policy.AcceptsOriginWithDelegation(child, "post", store) {
		t.Fatal("expected revoked delegation to stop granting write access")
	}
}

func TestLoadWriterPolicyMergesWhitelistAndBlacklistINF(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	policyPath := filepath.Join(root, "writer_policy.json")
	if err := os.WriteFile(policyPath, []byte("{\n  \"sync_mode\": \"all\"\n}\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(policy) error = %v", err)
	}
	whitelist := "# comment\nagent://news/publisher-01\npublic_key=abcd1234\n"
	if err := os.WriteFile(filepath.Join(root, writerWhitelistINFName), []byte(whitelist), 0o644); err != nil {
		t.Fatalf("WriteFile(whitelist) error = %v", err)
	}
	blacklist := "agent_id=agent://spam/bot-99\ndeadbeef9999\n"
	if err := os.WriteFile(filepath.Join(root, writerBlacklistINFName), []byte(blacklist), 0o644); err != nil {
		t.Fatalf("WriteFile(blacklist) error = %v", err)
	}

	policy, err := LoadWriterPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadWriterPolicy error = %v", err)
	}
	if !containsFold(policy.AllowedAgentIDs, "agent://news/publisher-01") {
		t.Fatal("expected whitelist agent to be merged")
	}
	if !containsFold(policy.AllowedPublicKeys, "abcd1234") {
		t.Fatal("expected whitelist public key to be merged")
	}
	if !containsFold(policy.BlockedAgentIDs, "agent://spam/bot-99") {
		t.Fatal("expected blacklist agent to be merged")
	}
	if !containsFold(policy.BlockedPublicKeys, "deadbeef9999") {
		t.Fatal("expected blacklist public key to be merged")
	}
}
