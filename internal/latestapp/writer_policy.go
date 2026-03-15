package latestapp

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
	policy.normalize()
	return policy, nil
}

func defaultWriterPolicy() WriterPolicy {
	policy := WriterPolicy{AllowUnsigned: true}
	policy.normalize()
	return policy
}

func (p *WriterPolicy) normalize() {
	if p == nil {
		return
	}
	p.AllowedAgentIDs = uniqueFold(p.AllowedAgentIDs)
	p.AllowedPublicKeys = normalizeHexList(p.AllowedPublicKeys)
	p.BlockedAgentIDs = uniqueFold(p.BlockedAgentIDs)
	p.BlockedPublicKeys = normalizeHexList(p.BlockedPublicKeys)
}

func (p WriterPolicy) Empty() bool {
	p.normalize()
	return p.AllowUnsigned && len(p.AllowedAgentIDs) == 0 && len(p.AllowedPublicKeys) == 0 && len(p.BlockedAgentIDs) == 0 && len(p.BlockedPublicKeys) == 0
}

func (p WriterPolicy) allowsOrigin(origin *MessageOrigin) bool {
	p.normalize()
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
			if !policy.allowsOrigin(bundle.Message.Origin) {
				continue
			}
			allowed[strings.ToLower(bundle.InfoHash)] = struct{}{}
			filtered = append(filtered, bundle)
		}
	}
	for _, bundle := range index.Bundles {
		switch bundle.Message.Kind {
		case "reply":
			if !policy.allowsOrigin(bundle.Message.Origin) {
				continue
			}
			if bundle.Message.ReplyTo != nil {
				if _, ok := allowed[strings.ToLower(bundle.Message.ReplyTo.InfoHash)]; ok {
					filtered = append(filtered, bundle)
				}
			}
		case "reaction":
			if !policy.allowsOrigin(bundle.Message.Origin) {
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
