package storage

import (
	"sync"

	"github.com/runar-rkmedia/gotally/sqlite"
)

type ruleCacheSqLite struct {
	rules map[string]sqlite.Rule
	sync.RWMutex
}

func newRuleCacheSqlite() ruleCacheSqLite {
	return ruleCacheSqLite{
		make(map[string]sqlite.Rule),
		sync.RWMutex{},
	}
}
func (p *ruleCacheSqLite) getCachedRule(hashOrID string) *sqlite.Rule {
	p.RLock()
	defer p.RUnlock()
	r, ok := p.rules[hashOrID]
	if !ok {
		return nil
	}
	return &r
}

func (p *ruleCacheSqLite) addRulesToCache(rules []sqlite.Rule) {
	p.Lock()
	defer p.Unlock()
	for i := 0; i < len(rules); i++ {
		p.rules[rules[i].Slug] = rules[i]
		p.rules[rules[i].ID] = rules[i]
	}
}
