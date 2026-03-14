package aip2p

import "testing"

func TestCanonicalizeMagnetDropsTrackersAndPeerHints(t *testing.T) {
	raw := "magnet:?xt=urn:btih:73c3893f0ab6c7e8bcdd2da53fb4a49addf14240&dn=test-post&tr=https%3A%2F%2Ftracker.example%2Fannounce&x.pe=192.168.102.75:52893"
	got := CanonicalizeMagnet(raw)
	want := "magnet:?dn=test-post&xt=urn%3Abtih%3A73c3893f0ab6c7e8bcdd2da53fb4a49addf14240"
	if got != want {
		t.Fatalf("canonical magnet = %q, want %q", got, want)
	}
}

func TestCanonicalMessageLinkSanitizesReplyMagnet(t *testing.T) {
	link := canonicalMessageLink(&MessageLink{
		InfoHash: "73C3893F0AB6C7E8BCDD2DA53FB4A49ADDF14240",
		Magnet:   "magnet:?xt=urn:btih:73c3893f0ab6c7e8bcdd2da53fb4a49addf14240&dn=parent&tr=https%3A%2F%2Ftracker.example%2Fannounce",
	})
	if link == nil {
		t.Fatal("canonical message link is nil")
	}
	want := "magnet:?dn=parent&xt=urn%3Abtih%3A73c3893f0ab6c7e8bcdd2da53fb4a49addf14240"
	if link.Magnet != want {
		t.Fatalf("reply magnet = %q, want %q", link.Magnet, want)
	}
	if link.InfoHash != "73c3893f0ab6c7e8bcdd2da53fb4a49addf14240" {
		t.Fatalf("reply infohash = %q", link.InfoHash)
	}
}
