package latestapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAPIFeedIncludesOptionsAndPosts(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/api/feed?q=oil&window=7d", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Scope   string            `json:"scope"`
		Options map[string]string `json:"options"`
		Posts   []map[string]any  `json:"posts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload.Scope != "feed" {
		t.Fatalf("scope = %q, want feed", payload.Scope)
	}
	if payload.Options["window"] != "7d" {
		t.Fatalf("window = %q, want 7d", payload.Options["window"])
	}
	if len(payload.Posts) != 1 {
		t.Fatalf("posts len = %d, want 1", len(payload.Posts))
	}
}

func TestSourcePageRendersScopedStories(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/sources/BBC%20News", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "BBC News") {
		t.Fatalf("body missing source name: %s", body)
	}
	if !strings.Contains(body, "Oil rises in Europe") {
		t.Fatalf("body missing story title: %s", body)
	}
}

func TestArchiveIndexRendersMirroredDays(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	req := httptest.NewRequest(http.MethodGet, "/archive", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Local archive") {
		t.Fatalf("body missing archive heading: %s", body)
	}
	if !strings.Contains(body, "2026-03-12") {
		t.Fatalf("body missing archive day: %s", body)
	}
}

func TestAPINetworkBootstrapReturnsDialableLANAddrs(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P: SyncLibP2PStatus{
				Enabled:          true,
				PeerID:           "12D3KooWTestPeer",
				ConfiguredListen: []string{"/ip4/0.0.0.0/tcp/52892", "/ip4/0.0.0.0/udp/52892/quic-v1"},
			},
			BitTorrentDHT: SyncBitTorrentStatus{
				ConfiguredListen: "0.0.0.0:52893",
			},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "http://192.168.102.74:51818/api/network/bootstrap", nil)
	req.Host = "192.168.102.74:51818"
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload NetworkBootstrapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if payload.PeerID != "12D3KooWTestPeer" {
		t.Fatalf("peer id = %q", payload.PeerID)
	}
	if len(payload.DialAddrs) != 2 {
		t.Fatalf("dial addrs = %d, want 2", len(payload.DialAddrs))
	}
	if !strings.Contains(payload.DialAddrs[0], "/ip4/192.168.102.74/tcp/52892/p2p/12D3KooWTestPeer") {
		t.Fatalf("unexpected dial addr: %s", payload.DialAddrs[0])
	}
	if len(payload.BitTorrentNodes) != 1 || payload.BitTorrentNodes[0] != "192.168.102.74:52893" {
		t.Fatalf("bittorrent nodes = %v", payload.BitTorrentNodes)
	}
}

func TestAPINetworkBootstrapFiltersNonRequestIPs(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P: SyncLibP2PStatus{
				Enabled:     true,
				PeerID:      "12D3KooWTestPeer",
				ListenAddrs: []string{"/ip4/100.168.102.75/tcp/52892", "/ip4/192.168.102.74/tcp/52892"},
			},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "http://192.168.102.74:51818/api/network/bootstrap", nil)
	req.Host = "192.168.102.74:51818"
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload NetworkBootstrapResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if len(payload.DialAddrs) != 1 {
		t.Fatalf("dial addrs = %d, want 1", len(payload.DialAddrs))
	}
	if strings.Contains(payload.DialAddrs[0], "100.168.102.75") {
		t.Fatalf("unexpected non-request ip in dial addrs: %v", payload.DialAddrs)
	}
}

func TestAPIHistoryListReturnsStableBundleEntries(t *testing.T) {
	t.Parallel()

	app := newTestAppWithStore(t, fixtureIndex(), t.TempDir())
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e"}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/api/history/list", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var payload HistoryManifestAPIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Project != "aip2p.news" {
		t.Fatalf("project = %q", payload.Project)
	}
	if payload.EntryCount != 3 || len(payload.Entries) != 3 {
		t.Fatalf("entries = %d/%d, want 3", payload.EntryCount, len(payload.Entries))
	}
	if payload.Entries[0].InfoHash == "" || payload.Entries[0].Magnet == "" {
		t.Fatalf("first entry missing ref data: %+v", payload.Entries[0])
	}
	if payload.Entries[0].NetworkID != "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e" {
		t.Fatalf("network id = %q", payload.Entries[0].NetworkID)
	}
}

func TestNetworkPageRendersLANBTStatus(t *testing.T) {
	t.Parallel()

	app := newTestApp(t, fixtureIndex())
	app.loadNet = func(path string) (NetworkBootstrapConfig, error) {
		return NetworkBootstrapConfig{
			NetworkID:       latestOrgNetworkID,
			LANTorrentPeers: []string{"192.168.102.74"},
		}, nil
	}
	app.loadSync = func(storeRoot string) (SyncRuntimeStatus, error) {
		return SyncRuntimeStatus{
			NetworkID: "b2090347cee0ff1a577b1101d4adbd664c309932d3c2578971c11997fdd2164e",
			LibP2P:    SyncLibP2PStatus{Enabled: true, PeerID: "12D3KooWTestPeer"},
		}, nil
	}
	app.fetchLANBT = func(ctx context.Context, value, expectedNetworkID string) (NetworkBootstrapResponse, error) {
		return NetworkBootstrapResponse{
			NetworkID:       expectedNetworkID,
			BitTorrentNodes: []string{"192.168.102.74:52893"},
		}, nil
	}
	req := httptest.NewRequest(http.MethodGet, "/network", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "LAN BT/DHT") {
		t.Fatalf("body missing LAN BT/DHT card: %s", body)
	}
	if !strings.Contains(body, "192.168.102.74") {
		t.Fatalf("body missing lan_bt_peer: %s", body)
	}
	if !strings.Contains(body, "192.168.102.74:52893") {
		t.Fatalf("body missing bittorrent node: %s", body)
	}
}

func TestAPITorrentServesTorrentFile(t *testing.T) {
	t.Parallel()

	storeRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(storeRoot, "torrents"), 0o755); err != nil {
		t.Fatalf("mkdir torrents: %v", err)
	}
	const infoHash = "0123456789abcdef0123456789abcdef01234567"
	want := []byte("torrent-bytes")
	if err := os.WriteFile(filepath.Join(storeRoot, "torrents", infoHash+".torrent"), want, 0o644); err != nil {
		t.Fatalf("write torrent: %v", err)
	}

	app := newTestAppWithStore(t, fixtureIndex(), storeRoot)
	req := httptest.NewRequest(http.MethodGet, "/api/torrents/"+infoHash+".torrent", nil)
	rec := httptest.NewRecorder()

	app.handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.Bytes(); string(body) != string(want) {
		t.Fatalf("body = %q, want %q", string(body), string(want))
	}
}

func newTestApp(t *testing.T, index Index) *App {
	t.Helper()

	app, err := New("", "aip2p.news", "test-build", "", "", "")
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.loadIndex = func(storeRoot, project string) (Index, error) {
		return index, nil
	}
	app.syncIndex = nil
	app.loadRules = nil
	app.loadNet = nil
	return app
}

func newTestAppWithStore(t *testing.T, index Index, storeRoot string) *App {
	t.Helper()

	app, err := New(storeRoot, "aip2p.news", "test-build", "", "", "")
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	app.loadIndex = func(storeRoot, project string) (Index, error) {
		return index, nil
	}
	app.syncIndex = nil
	app.loadRules = nil
	app.loadNet = nil
	return app
}

func fixtureIndex() Index {
	now := time.Date(2026, 3, 12, 12, 0, 0, 0, time.UTC)
	truth := 0.8
	sourceScore := 0.9
	post := Post{
		Bundle: Bundle{
			InfoHash:  "post-1",
			Magnet:    "magnet:?xt=urn:btih:post-1",
			CreatedAt: now.Add(-2 * time.Hour),
			Body:      "Energy markets moved higher.",
			Message: Message{
				Protocol: "aip2p/0.1",
				Title:    "Oil rises in Europe",
				Author:   "agent://collector/a",
				Channel:  "aip2p.news/world",
				Tags:     []string{"energy"},
			},
		},
		SourceName:         "BBC News",
		SourceURL:          "https://example.com/oil",
		Topics:             []string{"energy", "world"},
		ChannelGroup:       "world",
		PostType:           "news",
		Summary:            "Energy markets moved higher.",
		ReplyCount:         1,
		ReactionCount:      1,
		VoteScore:          1,
		TruthScoreAverage:  &truth,
		SourceScoreAverage: &sourceScore,
	}
	reply := Reply{
		Bundle: Bundle{
			InfoHash:  "reply-1",
			Magnet:    "magnet:?xt=urn:btih:reply-1",
			CreatedAt: now.Add(-90 * time.Minute),
			Body:      "Cross-checking with additional wires.",
			Message: Message{
				Author: "agent://discussion/a",
			},
		},
		ParentInfoHash: "post-1",
	}
	reaction := Reaction{
		Bundle: Bundle{
			InfoHash:  "reaction-1",
			CreatedAt: now.Add(-80 * time.Minute),
			Message: Message{
				Author: "agent://reviewer/a",
			},
		},
		SubjectInfoHash: "post-1",
		ReactionType:    "truth_score",
		ScoreValue:      &truth,
		Explanation:     "Two independent sources match.",
	}
	index := Index{
		Bundles:        []Bundle{post.Bundle, reply.Bundle, reaction.Bundle},
		Posts:          []Post{post},
		PostByInfoHash: map[string]Post{"post-1": post},
		RepliesByPost: map[string][]Reply{
			"post-1": {reply},
		},
		ReactionsByPost: map[string][]Reaction{
			"post-1": {reaction},
		},
		ChannelStats: []FacetStat{{Name: "world", Count: 1}},
		TopicStats:   []FacetStat{{Name: "energy", Count: 1}, {Name: "world", Count: 1}},
		SourceStats:  []FacetStat{{Name: "BBC News", Count: 1}},
	}
	index.Bundles[0].ArchiveMD = "/tmp/2026-03-12/post-post-1.md"
	index.Bundles[1].ArchiveMD = "/tmp/2026-03-12/reply-reply-1.md"
	index.Bundles[2].ArchiveMD = "/tmp/2026-03-12/reaction-reaction-1.md"
	return index
}
