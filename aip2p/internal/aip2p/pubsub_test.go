package aip2p

import (
	"testing"
	"time"
)

func TestSubscribedAnnouncementTopics(t *testing.T) {
	t.Parallel()

	networkID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	topics := subscribedAnnouncementTopics(networkID, SyncSubscriptions{
		Topics: []string{"world", "WORLD"},
		Tags:   []string{"breaking"},
	})
	if len(topics) != 2 {
		t.Fatalf("topics len = %d, want 2", len(topics))
	}
	if topics[0] != "aip2p/announce/"+networkID+"/topic/world" && topics[1] != "aip2p/announce/"+networkID+"/topic/world" {
		t.Fatalf("missing topic subscription: %v", topics)
	}
	if topics[0] != "aip2p/announce/"+networkID+"/tag/breaking" && topics[1] != "aip2p/announce/"+networkID+"/tag/breaking" {
		t.Fatalf("missing tag subscription: %v", topics)
	}
}

func TestSubscribedAnnouncementTopicsReservedAllUsesGlobal(t *testing.T) {
	t.Parallel()

	networkID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	topics := subscribedAnnouncementTopics(networkID, SyncSubscriptions{
		Topics: []string{"all", "pc75"},
	})
	if len(topics) != 2 {
		t.Fatalf("topics len = %d, want 2", len(topics))
	}
	if topics[0] != "aip2p/announce/"+networkID+"/global" && topics[1] != "aip2p/announce/"+networkID+"/global" {
		t.Fatalf("missing global topic subscription: %v", topics)
	}
	if topics[0] != "aip2p/announce/"+networkID+"/topic/pc75" && topics[1] != "aip2p/announce/"+networkID+"/topic/pc75" {
		t.Fatalf("missing pc75 topic subscription: %v", topics)
	}
}

func TestAnnouncementTopicsAlwaysIncludeReservedAll(t *testing.T) {
	t.Parallel()

	networkID := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	topics := announcementTopics(networkID, SyncAnnouncement{Topics: []string{"pc75"}})
	found := false
	for _, topic := range topics {
		if topic == "aip2p/announce/"+networkID+"/topic/all" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing reserved all topic: %v", topics)
	}
}

func TestMatchesAnnouncement(t *testing.T) {
	t.Parallel()

	announcement := SyncAnnouncement{
		Channel:   "aip2p.news/world",
		NetworkID: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Topics:    []string{"world", "pc75"},
		Tags:      []string{"breaking"},
		Origin: &MessageOrigin{
			AgentID:   "agent://writer/test",
			PublicKey: "test-key",
		},
	}
	policy := defaultWriterPolicy()
	if !matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"pc75"}}, policy, DelegationStore{}) {
		t.Fatal("expected topic match")
	}
	if !matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"all"}}, policy, DelegationStore{}) {
		t.Fatal("expected reserved all topic match")
	}
	if !matchesAnnouncement(announcement, SyncSubscriptions{Channels: []string{"aip2p.news/world"}}, policy, DelegationStore{}) {
		t.Fatal("expected channel match")
	}
	if matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"markets"}}, policy, DelegationStore{}) {
		t.Fatal("unexpected topic match")
	}
}

func TestMatchesAnnouncementFiltersByMaxAgeDays(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	announcement := SyncAnnouncement{
		Channel:   "aip2p.news/world",
		CreatedAt: now.Add(-48 * time.Hour).Format(time.RFC3339),
		Topics:    []string{"world", "pc75"},
		Origin: &MessageOrigin{
			AgentID:   "agent://writer/test",
			PublicKey: "test-key",
		},
	}
	policy := defaultWriterPolicy()
	if matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"all"}, MaxAgeDays: 1}, policy, DelegationStore{}) {
		t.Fatal("expected stale announcement to be filtered")
	}
	if !matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"all"}, MaxAgeDays: 3}, policy, DelegationStore{}) {
		t.Fatal("expected announcement within max age")
	}
}

func TestMatchesAnnouncementFiltersByMaxBundleMB(t *testing.T) {
	t.Parallel()

	announcement := SyncAnnouncement{
		SizeBytes: 12 * 1024 * 1024,
		Topics:    []string{"world"},
		Origin: &MessageOrigin{
			AgentID:   "agent://writer/test",
			PublicKey: "test-key",
		},
	}
	policy := defaultWriterPolicy()
	if matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"all"}, MaxBundleMB: 10}, policy, DelegationStore{}) {
		t.Fatal("expected oversized announcement to be filtered")
	}
	if !matchesAnnouncement(announcement, SyncSubscriptions{Topics: []string{"all"}, MaxBundleMB: 20}, policy, DelegationStore{}) {
		t.Fatal("expected announcement within size limit")
	}
}
