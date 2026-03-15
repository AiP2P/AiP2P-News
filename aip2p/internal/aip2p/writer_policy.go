package aip2p

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type WriterPolicy struct {
	AllowUnsigned     bool     `json:"allow_unsigned"`
	AllowedAgentIDs   []string `json:"allowed_agent_ids"`
	AllowedPublicKeys []string `json:"allowed_public_keys"`
	BlockedAgentIDs   []string `json:"blocked_agent_ids"`
	BlockedPublicKeys []string `json:"blocked_public_keys"`
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
	return policy, nil
}

func defaultWriterPolicy() WriterPolicy {
	policy := WriterPolicy{AllowUnsigned: true}
	policy.Normalize()
	return policy
}

func (p *WriterPolicy) Normalize() {
	if p == nil {
		return
	}
	p.AllowedAgentIDs = uniqueFold(p.AllowedAgentIDs)
	p.AllowedPublicKeys = normalizeHexList(p.AllowedPublicKeys)
	p.BlockedAgentIDs = uniqueFold(p.BlockedAgentIDs)
	p.BlockedPublicKeys = normalizeHexList(p.BlockedPublicKeys)
}

func (p WriterPolicy) Empty() bool {
	p.Normalize()
	return p.AllowUnsigned && len(p.AllowedAgentIDs) == 0 && len(p.AllowedPublicKeys) == 0 && len(p.BlockedAgentIDs) == 0 && len(p.BlockedPublicKeys) == 0
}

func (p WriterPolicy) AllowsOrigin(origin *MessageOrigin) bool {
	p.Normalize()
	if origin == nil {
		return p.AllowUnsigned && len(p.AllowedAgentIDs) == 0 && len(p.AllowedPublicKeys) == 0
	}
	agentID := strings.TrimSpace(origin.AgentID)
	publicKey := strings.ToLower(strings.TrimSpace(origin.PublicKey))
	if agentID != "" && containsFold(p.BlockedAgentIDs, agentID) {
		return false
	}
	if publicKey != "" && containsFold(p.BlockedPublicKeys, publicKey) {
		return false
	}
	if len(p.AllowedAgentIDs) == 0 && len(p.AllowedPublicKeys) == 0 {
		return true
	}
	if agentID != "" && containsFold(p.AllowedAgentIDs, agentID) {
		return true
	}
	if publicKey != "" && containsFold(p.AllowedPublicKeys, publicKey) {
		return true
	}
	return false
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
