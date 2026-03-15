package aip2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
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

type WriterOriginDecision struct {
	Capability        WriterCapability
	Delegation        *WriterDelegation
	ExplicitReadWrite bool
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
	policy.Normalize()
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
	policy.Normalize()
	return policy
}

func (p *WriterPolicy) Normalize() {
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
	p.Normalize()
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

func (p WriterPolicy) AllowsOrigin(origin *MessageOrigin) bool {
	return p.AllowsOriginWithDelegation(origin, "", DelegationStore{})
}

func (p WriterPolicy) AcceptsOrigin(origin *MessageOrigin) bool {
	return p.AcceptsOriginWithDelegation(origin, "", DelegationStore{})
}

func (p WriterPolicy) AllowsOriginWithDelegation(origin *MessageOrigin, scope string, store DelegationStore) bool {
	return p.OriginDecision(origin, scope, store).Capability == WriterCapabilityReadWrite
}

func (p WriterPolicy) AcceptsOriginWithDelegation(origin *MessageOrigin, scope string, store DelegationStore) bool {
	p.Normalize()
	if origin == nil {
		switch p.SyncMode {
		case WriterSyncModeWhitelist, WriterSyncModeTrustedWritersOnly:
			return false
		default:
			return p.AllowUnsigned
		}
	}

	decision := p.OriginDecision(origin, scope, store)
	if decision.Capability == WriterCapabilityBlocked {
		return false
	}

	switch p.SyncMode {
	case WriterSyncModeAll:
		return true
	case WriterSyncModeBlacklist:
		return true
	case WriterSyncModeWhitelist:
		return decision.ExplicitReadWrite
	case WriterSyncModeTrustedWritersOnly:
		return decision.Capability == WriterCapabilityReadWrite
	case WriterSyncModeMixed:
		fallthrough
	default:
		return decision.Capability == WriterCapabilityReadWrite
	}
}

func (p WriterPolicy) CapabilityForOrigin(origin *MessageOrigin) WriterCapability {
	return p.CapabilityForOriginWithDelegation(origin, "", DelegationStore{})
}

func (p WriterPolicy) CapabilityForOriginWithDelegation(origin *MessageOrigin, scope string, store DelegationStore) WriterCapability {
	return p.OriginDecision(origin, scope, store).Capability
}

func (p WriterPolicy) OriginDecision(origin *MessageOrigin, scope string, store DelegationStore) WriterOriginDecision {
	p.Normalize()
	if origin == nil {
		if !p.AllowUnsigned {
			return WriterOriginDecision{Capability: WriterCapabilityBlocked}
		}
		if p.hasLegacyWhitelist() {
			return WriterOriginDecision{Capability: WriterCapabilityReadOnly}
		}
		return WriterOriginDecision{Capability: p.DefaultCapability}
	}

	child := p.capabilityState(origin.AgentID, origin.PublicKey)
	if child.Capability == WriterCapabilityBlocked {
		return WriterOriginDecision{Capability: WriterCapabilityBlocked}
	}

	decision := WriterOriginDecision{
		Capability:        child.Capability,
		ExplicitReadWrite: child.ExplicitReadWrite,
	}
	scope = normalizeFoldKey(scope)
	if delegation, ok := store.ActiveDelegationFor(strings.TrimSpace(origin.AgentID), strings.ToLower(strings.TrimSpace(origin.PublicKey)), scope, time.Time{}); ok {
		parent := p.capabilityState(delegation.ParentAgentID, delegation.ParentPublicKey)
		if parent.Capability == WriterCapabilityBlocked {
			return WriterOriginDecision{
				Capability:        WriterCapabilityBlocked,
				Delegation:        delegation,
				ExplicitReadWrite: false,
			}
		}
		decision.Delegation = delegation
		if !child.ExplicitlyConfigured && capabilityRank(parent.Capability) > capabilityRank(decision.Capability) {
			decision.Capability = parent.Capability
		}
		if parent.ExplicitReadWrite {
			decision.ExplicitReadWrite = true
		}
	}
	return decision
}

func (p WriterPolicy) RelayTrustFor(peerID, host string) RelayTrust {
	p.Normalize()
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

func (p WriterPolicy) AcceptsRelay(peerID, host string) bool {
	return p.RelayTrustFor(peerID, host) != RelayTrustBlocked
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

type writerCapabilityState struct {
	Capability           WriterCapability
	ExplicitlyConfigured bool
	ExplicitReadWrite    bool
}

func (p WriterPolicy) capabilityState(agentIDValue, publicKeyValue string) writerCapabilityState {
	agentID := normalizeFoldKey(agentIDValue)
	publicKey := strings.ToLower(strings.TrimSpace(publicKeyValue))

	if agentID != "" && containsFold(p.BlockedAgentIDs, agentID) {
		return writerCapabilityState{Capability: WriterCapabilityBlocked, ExplicitlyConfigured: true}
	}
	if publicKey != "" && containsFold(p.BlockedPublicKeys, publicKey) {
		return writerCapabilityState{Capability: WriterCapabilityBlocked, ExplicitlyConfigured: true}
	}
	if capability, ok := p.PublicKeyCapabilities[publicKey]; ok {
		return writerCapabilityState{
			Capability:           capability,
			ExplicitlyConfigured: true,
			ExplicitReadWrite:    capability == WriterCapabilityReadWrite,
		}
	}
	if capability, ok := p.AgentCapabilities[agentID]; ok {
		return writerCapabilityState{
			Capability:           capability,
			ExplicitlyConfigured: true,
			ExplicitReadWrite:    capability == WriterCapabilityReadWrite,
		}
	}
	if p.hasLegacyWhitelist() {
		if agentID != "" && containsFold(p.AllowedAgentIDs, agentID) {
			return writerCapabilityState{
				Capability:           WriterCapabilityReadWrite,
				ExplicitlyConfigured: true,
				ExplicitReadWrite:    true,
			}
		}
		if publicKey != "" && containsFold(p.AllowedPublicKeys, publicKey) {
			return writerCapabilityState{
				Capability:           WriterCapabilityReadWrite,
				ExplicitlyConfigured: true,
				ExplicitReadWrite:    true,
			}
		}
		return writerCapabilityState{Capability: WriterCapabilityReadOnly}
	}
	return writerCapabilityState{Capability: p.DefaultCapability}
}

func capabilityRank(capability WriterCapability) int {
	switch capability {
	case WriterCapabilityBlocked:
		return 0
	case WriterCapabilityReadOnly:
		return 1
	case WriterCapabilityReadWrite:
		return 2
	default:
		return 0
	}
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
	p.Normalize()
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
	merged.Normalize()
	*p = merged
	return nil
}
