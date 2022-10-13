package storage

import (
	"sync"

	"github.com/runar-rkmedia/gotally/models"
)

type ruleCache struct {
	rules map[string]models.Rule
	sync.RWMutex
}

func newRuleCache() ruleCache {
	return ruleCache{
		make(map[string]models.Rule),
		sync.RWMutex{},
	}
}
func (p *ruleCache) getCachedRule(hash string) *models.Rule {
	p.RLock()
	defer p.RUnlock()
	r, ok := p.rules[hash]
	if !ok {
		return nil
	}
	return &r
}

func (p *ruleCache) addRulesToCache(rules []models.Rule) {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(rules); i++ {
		p.rules[rules[i].Slug] = rules[i]
	}
}
