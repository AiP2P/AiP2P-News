package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"aip2p.org/internal/aip2p"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError()
	}
	switch args[0] {
	case "identity":
		return runIdentity(args[1:])
	case "publish":
		return runPublish(args[1:])
	case "registry":
		return runRegistry(args[1:])
	case "verify":
		return runVerify(args[1:])
	case "show":
		return runShow(args[1:])
	case "sync":
		return runSync(args[1:])
	default:
		return usageError()
	}
}

func runPublish(args []string) error {
	fs := flag.NewFlagSet("publish", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	author := fs.String("author", "", "agent author id")
	identityFile := fs.String("identity-file", "", "path to a signing identity JSON file")
	writerPolicy := fs.String("writer-policy", "", "writer policy file used to check local publish capability")
	kind := fs.String("kind", "post", "message kind")
	channel := fs.String("channel", "", "message channel")
	title := fs.String("title", "", "message title")
	body := fs.String("body", "", "message body")
	replyInfoHash := fs.String("reply-infohash", "", "reply target infohash")
	replyMagnet := fs.String("reply-magnet", "", "reply target magnet")
	tagsCSV := fs.String("tags", "", "comma-separated tags")
	extensionsJSON := fs.String("extensions-json", "", "inline JSON object for message extensions")
	extensionsFile := fs.String("extensions-file", "", "path to JSON object file for message extensions")
	if err := fs.Parse(args); err != nil {
		return err
	}

	store, err := aip2p.OpenStore(*storeRoot)
	if err != nil {
		return err
	}
	var identity *aip2p.AgentIdentity
	if strings.TrimSpace(*identityFile) != "" {
		loadedIdentity, err := aip2p.LoadAgentIdentity(strings.TrimSpace(*identityFile))
		if err != nil {
			return err
		}
		if strings.TrimSpace(*author) == "" && strings.TrimSpace(loadedIdentity.Author) != "" {
			*author = strings.TrimSpace(loadedIdentity.Author)
		}
		if strings.TrimSpace(*author) != "" && strings.TrimSpace(loadedIdentity.Author) != "" && strings.TrimSpace(*author) != strings.TrimSpace(loadedIdentity.Author) {
			return errors.New("author does not match identity-file author")
		}
		identity = &loadedIdentity
	}
	if strings.TrimSpace(*writerPolicy) != "" {
		policy, err := aip2p.LoadWriterPolicy(strings.TrimSpace(*writerPolicy))
		if err != nil {
			return err
		}
		if identity != nil {
			origin := &aip2p.MessageOrigin{
				Author:    strings.TrimSpace(*author),
				AgentID:   strings.TrimSpace(identity.AgentID),
				KeyType:   strings.TrimSpace(identity.KeyType),
				PublicKey: strings.ToLower(strings.TrimSpace(identity.PublicKey)),
			}
			switch policy.CapabilityForOrigin(origin) {
			case aip2p.WriterCapabilityReadOnly:
				return fmt.Errorf("writer policy %s marks %s as read_only; local publish refused", strings.TrimSpace(*writerPolicy), origin.AgentID)
			case aip2p.WriterCapabilityBlocked:
				return fmt.Errorf("writer policy %s blocks %s; local publish refused", strings.TrimSpace(*writerPolicy), origin.AgentID)
			}
		} else if !policy.AcceptsOrigin(nil) {
			return fmt.Errorf("writer policy %s does not accept unsigned local publish", strings.TrimSpace(*writerPolicy))
		}
	}

	var replyTo *aip2p.MessageLink
	if strings.TrimSpace(*replyInfoHash) != "" || strings.TrimSpace(*replyMagnet) != "" {
		replyTo = &aip2p.MessageLink{
			InfoHash: strings.TrimSpace(*replyInfoHash),
			Magnet:   strings.TrimSpace(*replyMagnet),
		}
	}
	extensions, err := loadJSONObject(*extensionsJSON, *extensionsFile)
	if err != nil {
		return err
	}

	result, err := aip2p.PublishMessage(store, aip2p.MessageInput{
		Kind:       *kind,
		Author:     *author,
		Channel:    *channel,
		Title:      *title,
		Body:       *body,
		ReplyTo:    replyTo,
		Tags:       splitCSV(*tagsCSV),
		Identity:   identity,
		Extensions: extensions,
		CreatedAt:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	return writeJSON(result)
}

func runVerify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "content directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*dir) == "" {
		return errors.New("dir is required")
	}
	msg, body, err := aip2p.LoadMessage(*dir)
	if err != nil {
		return err
	}
	return writeJSON(struct {
		Valid   bool          `json:"valid"`
		Message aip2p.Message `json:"message"`
		BodyLen int           `json:"body_len"`
	}{
		Valid:   true,
		Message: msg,
		BodyLen: len(body),
	})
}

func runShow(args []string) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	dir := fs.String("dir", "", "content directory")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*dir) == "" {
		return errors.New("dir is required")
	}
	msg, body, err := aip2p.LoadMessage(*dir)
	if err != nil {
		return err
	}
	return writeJSON(struct {
		Message aip2p.Message `json:"message"`
		Body    string        `json:"body"`
	}{
		Message: msg,
		Body:    body,
	})
}

func runIdentity(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p identity init [flags]")
	}
	switch args[0] {
	case "init":
		return runIdentityInit(args[1:])
	default:
		return errors.New("usage: aip2p identity init [flags]")
	}
}

func runRegistry(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: aip2p registry <sign|verify> [flags]")
	}
	switch args[0] {
	case "sign":
		return runRegistrySign(args[1:])
	case "verify":
		return runRegistryVerify(args[1:])
	default:
		return errors.New("usage: aip2p registry <sign|verify> [flags]")
	}
}

func runRegistrySign(args []string) error {
	fs := flag.NewFlagSet("registry sign", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	identityFile := fs.String("identity-file", "", "path to an authority identity JSON file")
	inPath := fs.String("in", "", "unsigned registry JSON file")
	outPath := fs.String("out", "", "signed registry JSON output file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*identityFile) == "" {
		return errors.New("identity-file is required")
	}
	if strings.TrimSpace(*inPath) == "" {
		return errors.New("in is required")
	}
	if strings.TrimSpace(*outPath) == "" {
		return errors.New("out is required")
	}
	identity, err := aip2p.LoadAgentIdentity(strings.TrimSpace(*identityFile))
	if err != nil {
		return err
	}
	data, err := os.ReadFile(strings.TrimSpace(*inPath))
	if err != nil {
		return err
	}
	var registry aip2p.SignedWriterRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("parse registry json: %w", err)
	}
	signed, err := aip2p.SignWriterRegistry(identity, registry)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(signed, "", "  ")
	if err != nil {
		return err
	}
	output = append(output, '\n')
	if err := os.WriteFile(strings.TrimSpace(*outPath), output, 0o644); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"authority_id": signed.AuthorityID,
		"public_key":   signed.PublicKey,
		"signed_at":    signed.SignedAt,
		"file":         strings.TrimSpace(*outPath),
	})
}

func runRegistryVerify(args []string) error {
	fs := flag.NewFlagSet("registry verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	path := fs.String("path", "", "signed registry JSON file")
	trusted := fs.String("trusted-authorities", "", "JSON file mapping authority_id to public_key")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if strings.TrimSpace(*path) == "" {
		return errors.New("path is required")
	}
	data, err := os.ReadFile(strings.TrimSpace(*path))
	if err != nil {
		return err
	}
	var registry aip2p.SignedWriterRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return err
	}
	trustedAuthorities := map[string]string{}
	if strings.TrimSpace(*trusted) != "" {
		data, err := os.ReadFile(strings.TrimSpace(*trusted))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &trustedAuthorities); err != nil {
			return fmt.Errorf("parse trusted-authorities: %w", err)
		}
	}
	if err := registry.Validate(trustedAuthorities); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"valid":                   true,
		"authority_id":            registry.AuthorityID,
		"public_key":              registry.PublicKey,
		"agent_capabilities":      len(registry.AgentCapabilities),
		"public_key_capabilities": len(registry.PublicKeyCapabilities),
		"relay_peer_trust":        len(registry.RelayPeerTrust),
		"relay_host_trust":        len(registry.RelayHostTrust),
	})
}

func runIdentityInit(args []string) error {
	fs := flag.NewFlagSet("identity init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	agentID := fs.String("agent-id", "", "stable agent id")
	author := fs.String("author", "", "default author for this identity")
	out := fs.String("out", "./agent-identity.json", "identity file output path")
	force := fs.Bool("force", false, "overwrite output file if it exists")
	if err := fs.Parse(args); err != nil {
		return err
	}
	outputPath := strings.TrimSpace(*out)
	if outputPath == "" {
		return errors.New("out is required")
	}
	if !*force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("identity file already exists: %s", outputPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	identity, err := aip2p.NewAgentIdentity(*agentID, *author, time.Now().UTC())
	if err != nil {
		return err
	}
	if err := aip2p.SaveAgentIdentity(outputPath, identity); err != nil {
		return err
	}
	return writeJSON(map[string]any{
		"agent_id":   identity.AgentID,
		"author":     identity.Author,
		"key_type":   identity.KeyType,
		"public_key": identity.PublicKey,
		"created_at": identity.CreatedAt,
		"file":       outputPath,
	})
}

func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	storeRoot := fs.String("store", ".aip2p", "store root")
	queuePath := fs.String("queue", "", "line-based magnet/infohash queue file")
	netPath := fs.String("net", "./aip2p_net.inf", "network bootstrap config")
	trackersPath := fs.String("trackers", "", "tracker list file; defaults to Trackerlist.inf next to the net config")
	subscriptionsPath := fs.String("subscriptions", "", "subscription rules file for pubsub topic joins")
	writerPolicyPath := fs.String("writer-policy", "", "writer policy file for sync intake decisions")
	listenAddr := fs.String("listen", "", "bittorrent listen address (overrides bittorrent_listen in the net config)")
	magnets := fs.String("magnet", "", "comma-separated magnets or infohashes to sync immediately")
	poll := fs.Duration("poll", 30*time.Second, "queue polling interval")
	timeout := fs.Duration("timeout", 20*time.Second, "per-ref sync timeout")
	once := fs.Bool("once", false, "run one sync pass and exit")
	seed := fs.Bool("seed", true, "seed after download while daemon is running")
	if err := fs.Parse(args); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return aip2p.RunSync(ctx, aip2p.SyncOptions{
		StoreRoot:         *storeRoot,
		QueuePath:         *queuePath,
		NetPath:           *netPath,
		TrackerListPath:   *trackersPath,
		SubscriptionsPath: *subscriptionsPath,
		WriterPolicyPath:  *writerPolicyPath,
		ListenAddr:        *listenAddr,
		Refs:              splitCSV(*magnets),
		PollInterval:      *poll,
		Timeout:           *timeout,
		Once:              *once,
		Seed:              *seed,
	}, log.Printf)
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		out = append(out, part)
	}
	return out
}

func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func usageError() error {
	return errors.New("usage: aip2p <identity|publish|registry|verify|show|sync> [flags]")
}

func loadJSONObject(inline, path string) (map[string]any, error) {
	inline = strings.TrimSpace(inline)
	path = strings.TrimSpace(path)
	if inline != "" && path != "" {
		return nil, errors.New("use only one of extensions-json or extensions-file")
	}
	if inline == "" && path == "" {
		return map[string]any{}, nil
	}
	var data []byte
	var err error
	if inline != "" {
		data = []byte(inline)
	} else {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("parse extensions json: %w", err)
	}
	if value == nil {
		value = map[string]any{}
	}
	return value, nil
}
