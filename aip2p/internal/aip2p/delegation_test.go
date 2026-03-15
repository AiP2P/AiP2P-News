package aip2p

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSignAndValidateWriterDelegation(t *testing.T) {
	t.Parallel()

	parent, err := NewAgentIdentity("agent://news/main", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(parent) error = %v", err)
	}
	child, err := NewAgentIdentity("agent://news/world-01", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(child) error = %v", err)
	}
	delegation, err := SignWriterDelegation(parent, child, []string{"post", "reply"}, time.Now().UTC(), time.Time{})
	if err != nil {
		t.Fatalf("SignWriterDelegation error = %v", err)
	}
	if err := ValidateWriterDelegation(delegation); err != nil {
		t.Fatalf("ValidateWriterDelegation error = %v", err)
	}
}

func TestDelegationStoreActiveDelegationRespectsRevocation(t *testing.T) {
	t.Parallel()

	parent, err := NewAgentIdentity("agent://news/main", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(parent) error = %v", err)
	}
	child, err := NewAgentIdentity("agent://news/world-01", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(child) error = %v", err)
	}
	createdAt := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
	delegation, err := SignWriterDelegation(parent, child, []string{"post"}, createdAt, time.Time{})
	if err != nil {
		t.Fatalf("SignWriterDelegation error = %v", err)
	}
	revocation, err := SignWriterRevocation(parent, child.AgentID, child.PublicKey, "rotated", createdAt.Add(2*time.Hour))
	if err != nil {
		t.Fatalf("SignWriterRevocation error = %v", err)
	}
	store := DelegationStore{
		Delegations: []WriterDelegation{delegation},
		Revocations: []WriterRevocation{revocation},
	}
	if _, ok := store.ActiveDelegationFor(child.AgentID, child.PublicKey, "post", createdAt.Add(time.Hour)); !ok {
		t.Fatal("expected delegation to be active before revocation")
	}
	if _, ok := store.ActiveDelegationFor(child.AgentID, child.PublicKey, "post", createdAt.Add(3*time.Hour)); ok {
		t.Fatal("expected delegation to be inactive after revocation")
	}
}

func TestLoadDelegationStore(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	parent, err := NewAgentIdentity("agent://news/main", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(parent) error = %v", err)
	}
	child, err := NewAgentIdentity("agent://news/world-01", "", time.Now().UTC())
	if err != nil {
		t.Fatalf("NewAgentIdentity(child) error = %v", err)
	}
	delegation, err := SignWriterDelegation(parent, child, []string{"post"}, time.Now().UTC(), time.Time{})
	if err != nil {
		t.Fatalf("SignWriterDelegation error = %v", err)
	}
	revocation, err := SignWriterRevocation(parent, child.AgentID, child.PublicKey, "rotated", time.Now().UTC())
	if err != nil {
		t.Fatalf("SignWriterRevocation error = %v", err)
	}
	delegationsDir := filepath.Join(root, "delegations")
	revocationsDir := filepath.Join(root, "revocations")
	if err := SaveWriterDelegation(filepath.Join(delegationsDir, "world-01.json"), delegation); err != nil {
		t.Fatalf("SaveWriterDelegation error = %v", err)
	}
	if err := SaveWriterRevocation(filepath.Join(revocationsDir, "world-01.json"), revocation); err != nil {
		t.Fatalf("SaveWriterRevocation error = %v", err)
	}
	store, err := LoadDelegationStore(delegationsDir, revocationsDir)
	if err != nil {
		t.Fatalf("LoadDelegationStore error = %v", err)
	}
	if len(store.Delegations) != 1 {
		t.Fatalf("delegations len = %d, want 1", len(store.Delegations))
	}
	if len(store.Revocations) != 1 {
		t.Fatalf("revocations len = %d, want 1", len(store.Revocations))
	}
}
