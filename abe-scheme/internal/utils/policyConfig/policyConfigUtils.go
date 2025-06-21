package policyConfig

import (
	"github.com/pzkt/abe-scripts/abe-scheme/internal/crypto"
	"github.com/pzkt/abe-scripts/abe-scheme/internal/utils"
)

type Config struct {
	PurposeTrees []*utils.Tree
	Scheme       crypto.ABEscheme
}

func (p Config) ResolvePurpose(purpose string) []string {
	out := []string{}
	for _, pt := range p.PurposeTrees {
		node, found := pt.FindValue(purpose)
		if !found {
			continue
		}
		out = append(out, node.GetRootPath()...)
	}
	return out
}
