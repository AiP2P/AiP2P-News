package latestapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type WriterCapability string

const (
	WriterCapabilityReadWrite WriterCapability = "read_write"
	WriterCapabilityReadOnly  WriterCapability = "read_only"
	WriterCapabilityBlocked   WriterCapability = "blocked"
)

type WriterSyncMode string

const (
	WriterSyncModeMixed              WriterSyncMode = "mixed"
	WriterSyncModeAll                WriterSyncMode = "all"
	WriterSyncModeTrustedWritersOnly WriterSyncMode = "trusted_writers_only"
	WriterSyncModeWhitelist          WriterSyncMode = "whitelist"
	WriterSyncModeBlacklist          WriterSyncMode = "blacklist"
)

type WriterPolicy struct {
	SyncMode              WriterSyncMode              `json:"sync_mode,omitempty"`
	AllowUnsigned         bool                        `json:"allow_unsigned"`
	DefaultCapability     WriterCapability            `json:"default_capability,omitempty"`
	AgentCapabilities     map[string]WriterCapability `json:"agent_capabilities,omitempty"`
	PublicKeyCapabilities map[string]WriterCapability `json:"public_key_capabilities,omitempty"`
	AllowedAgentIDs       []string                    `json:"allowed_agent_ids"`
	AllowedPublicKeys     []string                    `json:"allowed_public_keys"`
	BlockedAgentIDs       []string                    `json:"blocked_agent_ids"`
	BlockedPublicKeys     []string                    `json:"blocked_public_keys"`
	TrustedAuthorities    map[string]string           `json:"trusted_authorities,omitempty"`
	SharedRegistries      []string                    `json:"shared_registries,omitempty"`
	RelayDefaultTrust     RelayTrust                  `json:"relay_default_trust,omitempty"`
	RelayPeerTrust        map[string]RelayTrust       `json:"relay_peer_trust,omitempty"`
	RelayHostTrust        map[string]RelayTrust       `json:"relay_host_trust,omitempty"`
}

func LoadWriterPolicy(path string) (WriterPolicy, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return defaultWriterPolicy(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultWriterPolicy(), nil
		}
		return WriterPolicy{}, err
	}
	var policy WriterPolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		return WriterPolicy{}, err
	}
	policy.normalize()
	if err := policy.loadSharedRegistries(); err != nil {
		return WriterPolicy{}, err
	}
	return policy, nil
}

func defaultWriterPolicy() WriterPolicy {
	policy := WriterPolicy{
		SyncMode:          WriterSyncModeMixed,
		AllowUnsigned:     true,
		DefaultCapability: WriterCapabilityReadWrite,
		RelayDefaultTrust: RelayTrustNeutral,
	}
	policy.normalize()
	return policy
}

func (p *WriterPolicy) normalize() {
	if p == nil {
		return
	}
	p.SyncMode = normalizeWriterSyncMode(p.SyncMode, WriterSyncModeMixed)
	p.DefaultCapability = normalizeWriterCapability(p.DefaultCapability, WriterCapabilityReadWrite)
	p.AllowedAgentIDs = uniqueFold(p.AllowedAgentIDs)
	p.AllowedPublicKeys = normalizeHexList(p.AllowedPublicKeys)
	p.BlockedAgentIDs = uniqueFold(p.BlockedAgentIDs)
	p.BlockedPublicKeys = normalizeHexList(p.BlockedPublicKeys)
	p.TrustedAuthorities = normalizePublicKeyMap(p.TrustedAuthorities)
	p.SharedRegistries = uniqueTrim(p.SharedRegistries)
	p.RelayDefaultTrust = normalizeRelayTrust(p.RelayDefaultTrust, RelayTrustNeutral)
	p.RelayPeerTrust = normalizeRelayTrustMap(p.RelayPeerTrust, false)
	p.RelayHostTrust = normalizeRelayTrustMap(p.RelayHostTrust, true)
	p.AgentCapabilities = normalizeCapabilityMap(p.AgentCapabilities, false)
	p.PublicKeyCapabilities = normalizeCapabilityMap(p.PublicKeyCapabilities, true)
}

func (p WriterPolicy) Empty() bool {
	p.normalize()
	return p.AllowUnsigned &&
		p.SyncMode == WriterSyncModeMixed &&
		p.DefaultCapability == WriterCapabilityReadWrite &&
		len(p.AgentCapabilities) == 0 &&
		len(p.PublicKeyCapabilities) == 0 &&
		len(p.AllowedAgentIDs) == 0 &&
		len(p.AllowedPublicKeys) == 0 &&
		len(p.BlockedAgentIDs) == 0 &&
		len(p.BlockedPublicKeys) == 0 &&
		len(p.TrustedAuthorities) == 0 &&
		len(p.SharedRegistries) == 0 &&
		p.RelayDefaultTrust == RelayTrustNeutral &&
		len(p.RelayPeerTrust) == 0 &&
		len(p.RelayHostTrust) == 0
}

func (p WriterPolicy) allowsOrigin(origin *MessageOrigin) bool {
	return p.capabilityForOrigin(origin) == WriterCapabilityReadWrite
}

func (p WriterPolicy) acceptsOrigin(origin *MessageOrigin) bool {
	p.normalize()
	if origin == nil {
		switch p.SyncMode {
		case WriterSyncModeWhitelist, WriterSyncModeTrustedWritersOnly:
			return false
		default:
			return p.AllowUnsigned
		}
	}

	capability := p.capabilityForOrigin(origin)
	if capability == WriterCapabilityBlocked {
		return false
	}

	switch p.SyncMode {
	case WriterSyncModeAll:
		return true
	case WriterSyncModeBlacklist:
		return true
	case WriterSyncModeWhitelist:
		return p.isExplicitlyAllowed(origin)
	case WriterSyncModeTrustedWritersOnly:
		return capability == WriterCapabilityReadWrite
	case WriterSyncModeMixed:
		fallthrough
	default:
		return capability == WriterCapabilityReadWrite
	}
}

func (p WriterPolicy) capabilityForOrigin(origin *MessageOrigin) WriterCapability {
	p.normalize()
	if origin == nil {
		if !p.AllowUnsigned {
			return WriterCapabilityBlocked
		}
		if p.hasLegacyWhitelist() {
			return WriterCapabilityReadOnly
		}
		return p.DefaultCapability
	}

	agentID := normalizeFoldKey(origin.AgentID)
	publicKey := strings.ToLower(strings.TrimSpace(origin.PublicKey))

	if agentID != "" && containsFold(p.BlockedAgentIDs, agentID) {
		return WriterCapabilityBlocked
	}
	if publicKey != "" && containsFold(p.BlockedPublicKeys, publicKey) {
		return WriterCapabilityBlocked
	}
	if capability, ok := p.PublicKeyCapabilities[publicKey]; ok {
		return capability
	}
	if capability, ok := p.AgentCapabilities[agentID]; ok {
		return capability
	}
	if p.hasLegacyWhitelist() {
		if agentID != "" && containsFold(p.AllowedAgentIDs, agentID) {
			return WriterCapabilityReadWrite
		}
		if publicKey != "" && containsFold(p.AllowedPublicKeys, publicKey) {
			return WriterCapabilityReadWrite
		}
		return WriterCapabilityReadOnly
	}
	return p.DefaultCapability
}

func (p WriterPolicy) relayTrustFor(peerID, host string) RelayTrust {
	p.normalize()
	peerID = strings.TrimSpace(peerID)
	host = normalizeFoldKey(host)
	if peerID != "" {
		if trust, ok := p.RelayPeerTrust[peerID]; ok {
			return trust
		}
	}
	if host != "" {
		if trust, ok := p.RelayHostTrust[host]; ok {
			return trust
		}
	}
	return p.RelayDefaultTrust
}

func (p WriterPolicy) acceptsRelay(peerID, host string) bool {
	return p.relayTrustFor(peerID, host) != RelayTrustBlocked
}

func (p WriterPolicy) hasLegacyWhitelist() bool {
	return len(p.AllowedAgentIDs) > 0 || len(p.AllowedPublicKeys) > 0
}

func (p WriterPolicy) isExplicitlyAllowed(origin *MessageOrigin) bool {
	if origin == nil {
		return false
	}
	agentID := normalizeFoldKey(origin.AgentID)
	publicKey := strings.ToLower(strings.TrimSpace(origin.PublicKey))
	if publicKey != "" {
		if capability, ok := p.PublicKeyCapabilities[publicKey]; ok {
			return capability == WriterCapabilityReadWrite
		}
	}
	if agentID != "" {
		if capability, ok := p.AgentCapabilities[agentID]; ok {
			return capability == WriterCapabilityReadWrite
		}
	}
	if agentID != "" && containsFold(p.AllowedAgentIDs, agentID) {
		return true
	}
	if publicKey != "" && containsFold(p.AllowedPublicKeys, publicKey) {
		return true
	}
	return false
}

func normalizeWriterSyncMode(value, fallback WriterSyncMode) WriterSyncMode {
	switch WriterSyncMode(strings.ToLower(strings.TrimSpace(string(value)))) {
	case WriterSyncModeMixed:
		return WriterSyncModeMixed
	case WriterSyncModeAll:
		return WriterSyncModeAll
	case WriterSyncModeTrustedWritersOnly:
		return WriterSyncModeTrustedWritersOnly
	case WriterSyncModeWhitelist:
		return WriterSyncModeWhitelist
	case WriterSyncModeBlacklist:
		return WriterSyncModeBlacklist
	default:
		return fallback
	}
}

func normalizeWriterCapability(value, fallback WriterCapability) WriterCapability {
	switch WriterCapability(strings.ToLower(strings.TrimSpace(string(value)))) {
	case WriterCapabilityReadWrite:
		return WriterCapabilityReadWrite
	case WriterCapabilityReadOnly:
		return WriterCapabilityReadOnly
	case WriterCapabilityBlocked:
		return WriterCapabilityBlocked
	default:
		return fallback
	}
}

func normalizeCapabilityMap(items map[string]WriterCapability, hexKeys bool) map[string]WriterCapability {
	if len(items) == 0 {
		return nil
	}
	normalized := make(map[string]WriterCapability, len(items))
	for key, capability := range items {
		if hexKeys {
			key = strings.ToLower(strings.TrimSpace(key))
		} else {
			key = normalizeFoldKey(key)
		}
		if key == "" {
			continue
		}
		normalized[key] = normalizeWriterCapability(capability, WriterCapabilityReadWrite)
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func normalizeFoldKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeHexList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.ToLower(strings.TrimSpace(item))
		if item == "" {
			continue
		}
		normalized = append(normalized, item)
	}
	return uniqueFold(normalized)
}

func (p *WriterPolicy) loadSharedRegistries() error {
	if p == nil {
		return nil
	}
	p.normalize()
	if len(p.SharedRegistries) == 0 {
		return nil
	}
	local := *p
	merged := WriterPolicy{
		SyncMode:           local.SyncMode,
		AllowUnsigned:      local.AllowUnsigned,
		DefaultCapability:  local.DefaultCapability,
		TrustedAuthorities: local.TrustedAuthorities,
		SharedRegistries:   append([]string(nil), local.SharedRegistries...),
		RelayDefaultTrust:  local.RelayDefaultTrust,
	}
	for _, source := range local.SharedRegistries {
		registry, err := loadSignedWriterRegistrySource(source)
		if err != nil {
			return err
		}
		if err := registry.Validate(local.TrustedAuthorities); err != nil {
			return fmt.Errorf("verify shared registry %s: %w", source, err)
		}
		merged.AgentCapabilities = mergeRegistryCapabilities(merged.AgentCapabilities, registry.AgentCapabilities)
		merged.PublicKeyCapabilities = mergeRegistryCapabilities(merged.PublicKeyCapabilities, registry.PublicKeyCapabilities)
		merged.RelayPeerTrust = mergeRegistryRelayTrust(merged.RelayPeerTrust, registry.RelayPeerTrust)
		merged.RelayHostTrust = mergeRegistryRelayTrust(merged.RelayHostTrust, registry.RelayHostTrust)
	}
	merged.AgentCapabilities = mergeRegistryCapabilities(merged.AgentCapabilities, local.AgentCapabilities)
	merged.PublicKeyCapabilities = mergeRegistryCapabilities(merged.PublicKeyCapabilities, local.PublicKeyCapabilities)
	merged.RelayPeerTrust = mergeRegistryRelayTrust(merged.RelayPeerTrust, local.RelayPeerTrust)
	merged.RelayHostTrust = mergeRegistryRelayTrust(merged.RelayHostTrust, local.RelayHostTrust)
	merged.AllowedAgentIDs = append(append([]string(nil), merged.AllowedAgentIDs...), local.AllowedAgentIDs...)
	merged.AllowedPublicKeys = append(append([]string(nil), merged.AllowedPublicKeys...), local.AllowedPublicKeys...)
	merged.BlockedAgentIDs = append(append([]string(nil), merged.BlockedAgentIDs...), local.BlockedAgentIDs...)
	merged.BlockedPublicKeys = append(append([]string(nil), merged.BlockedPublicKeys...), local.BlockedPublicKeys...)
	merged.normalize()
	*p = merged
	return nil
}

func ApplyWriterPolicy(index Index, project string, policy WriterPolicy) Index {
	policy.normalize()
	if policy.Empty() {
		return index
	}
	filtered := make([]Bundle, 0, len(index.Bundles))
	allowed := make(map[string]struct{})
	for _, bundle := range index.Bundles {
		switch bundle.Message.Kind {
		case "post":
			if !policy.acceptsOrigin(bundle.Message.Origin) {
				continue
			}
			allowed[strings.ToLower(bundle.InfoHash)] = struct{}{}
			filtered = append(filtered, bundle)
		}
	}
	for _, bundle := range index.Bundles {
		switch bundle.Message.Kind {
		case "reply":
			if !policy.acceptsOrigin(bundle.Message.Origin) {
				continue
			}
			if bundle.Message.ReplyTo != nil {
				if _, ok := allowed[strings.ToLower(bundle.Message.ReplyTo.InfoHash)]; ok {
					filtered = append(filtered, bundle)
				}
			}
		case "reaction":
			if !policy.acceptsOrigin(bundle.Message.Origin) {
				continue
			}
			subject := strings.ToLower(nestedString(bundle.Message.Extensions, "subject", "infohash"))
			if _, ok := allowed[subject]; ok {
				filtered = append(filtered, bundle)
			}
		}
	}
	return buildIndex(filtered, project)
}
