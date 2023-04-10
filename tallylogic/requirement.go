package tallylogic

func (s StatsRequirement) Excludes(stats GameStats) bool {
	if s.CellCount != nil && s.CellCount.Excludes(stats.CellCount) {
		return true
	}
	if s.DuplicateFactors != nil && s.DuplicateFactors.Excludes(stats.DuplicateFactors) {
		return true
	}
	if s.DuplicateValues != nil && s.DuplicateValues.Excludes(stats.DuplicateValues) {
		return true
	}
	if s.UniqueFactorCount != nil && s.UniqueFactorCount.Excludes(len(stats.UniqueFactors)) {
		return true
	}
	if s.WithValueCount != nil && s.WithValueCount.Excludes(stats.WithValueCount) {
		return true
	}

	if s.UniquFactors != nil && s.UniquFactors.Excludes(stats.UniqueFactors) {
		return true
	}
	if s.UniqeValues != nil && s.UniqeValues.Excludes(stats.UniqueValues) {
		return true
	}
	return false
}

func (r IntListRequirement) Excludes(list []uint64) bool {
	if r.IncludesItems != nil {
	ox:
		for _, r := range *r.IncludesItems {
			for _, v := range list {
				if r == v {
					continue ox
				}
			}
			return true
		}
	}
	if r.ExcludesItems != nil {
		for _, r := range *r.IncludesItems {
			for _, v := range list {
				if r == v {
					return true
				}
			}
		}
	}
	if r.OnlyItems != nil {
		if len(*r.OnlyItems) != len(list) {
			return true
		}
	oi:
		for _, r := range *r.OnlyItems {
			for _, v := range list {
				if r == v {
					continue oi
				}
			}
			return true
		}
	}
	return false
}
func (r IntRequirement) Excludes(n int) bool {
	if r.GT != nil && !(n > *r.GT) {
		return true
	}
	if r.GTE != nil && !(n >= *r.GTE) {
		return true
	}
	if r.EQ != nil && !(n == *r.EQ) {
		return true
	}
	if r.LT != nil && !(n < *r.LT) {
		return true
	}
	if r.LTE != nil && !(n <= *r.LTE) {
		return true
	}

	return false
}

type StatsRequirement struct {
	CellCount, DuplicateFactors, DuplicateValues, UniqueFactorCount, WithValueCount *IntRequirement
	UniquFactors, UniqeValues                                                       *IntListRequirement
}
type IntListRequirement struct {
	// All items must exist in result, but there can be more items
	IncludesItems *[]uint64
	// All items must not exist in result
	ExcludesItems *[]uint64
	// Must be these exact items (order does not matter)
	OnlyItems *[]uint64
}
type IntRequirement struct {
	GT, GTE, EQ, LT, LTE *int
}
