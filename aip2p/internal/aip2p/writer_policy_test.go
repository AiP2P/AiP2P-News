package aip2p

import (
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
		AllowUnsigned:   true,
		AllowedAgentIDs: []string{"agent://writer/allowed"},
	}
	allowed := &MessageOrigin{AgentID: "agent://writer/allowed"}
	denied := &MessageOrigin{AgentID: "agent://writer/other"}

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

	open := WriterPolicy{SyncMode: WriterSyncModeAll, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadWrite}
	if !open.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be accepted when allowed")
	}

	restricted := WriterPolicy{SyncMode: WriterSyncModeMixed, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadOnly}
	if restricted.CapabilityForOrigin(nil) != WriterCapabilityReadOnly {
		t.Fatalf("unsigned capability = %q, want read_only", restricted.CapabilityForOrigin(nil))
	}
	if !restricted.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be accepted in mixed mode when allow_unsigned is true")
	}

	trustedOnly := WriterPolicy{SyncMode: WriterSyncModeTrustedWritersOnly, AllowUnsigned: true, DefaultCapability: WriterCapabilityReadWrite}
	if trustedOnly.AcceptsOrigin(nil) {
		t.Fatal("expected unsigned writer to be rejected in trusted_writers_only mode")
	}
}

func TestWriterPolicySyncModeAllAcceptsReadOnlyWritersButStillBlocksBlockedOnes(t *testing.T) {
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

func TestWriterPolicyTrustedWritersOnlyRequiresReadWriteCapability(t *testing.T) {
	t.Parallel()

	policy := WriterPolicy{
		SyncMode:          WriterSyncModeTrustedWritersOnly,
		DefaultCapability: WriterCapabilityReadOnly,
		AgentCapabilities: map[string]WriterCapability{
			"agent://writer/trusted": WriterCapabilityReadWrite,
		},
	}
	trusted := &MessageOrigin{AgentID: "agent://writer/trusted"}
	untrusted := &MessageOrigin{AgentID: "agent://writer/readonly"}

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
