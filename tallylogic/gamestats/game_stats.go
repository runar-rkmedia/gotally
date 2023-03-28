package gamestats

import (
	"fmt"
	"sort"

	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

func NewGameStats(game tallylogic.Game) (GameStats, error) {
	stats := GameStats{}
	cells := game.Cells()
	stats.CellCount = len(cells)
	cellValues := []uint64{}
	combinedFactors := cell.NewFactors(0)
	for _, c := range cells {
		if !c.IsEmpty() {
			cellValues = append(cellValues, uint64(c.Value()))
		}
		factors := c.Factors().Factors()
		for _, v := range factors {
			combinedFactors.AddFactor(v)
		}
	}
	stats.Hints = game.Hinter.GetHints()
	hintPaths := map[string]struct{}{}
	for _, v := range stats.Hints {
		p := fmt.Sprintf("%v", v.Path)
		if _, ok := hintPaths[p]; ok {
			continue
		}
		reversed := ReverseSlice(v.Path)
		rev := fmt.Sprintf("%v", reversed)
		if _, ok := hintPaths[rev]; ok {
			continue
		}
		hintPaths[p] = struct{}{}
	}
	stats.UniqueHints = len(hintPaths)
	stats.WithValueCount = len(cellValues)
	stats.UniqueValues = unique(cellValues)
	sort.Slice(stats.UniqueValues, func(i, j int) bool { return stats.UniqueValues[i] < stats.UniqueValues[j] })
	stats.UniqueFactors = combinedFactors.UniqueFactors()

	stats.DuplicateFactors = len(combinedFactors.Factors()) - len(stats.UniqueFactors)
	stats.DuplicateValues = len(cellValues) - len(stats.UniqueFactors)
	return stats, nil
}

type GameStats struct {
	// List of unique factors across all cells
	UniqueFactors []uint64
	// List of unique values across all cells
	UniqueValues []uint64
	// Count of duplicate factors
	DuplicateFactors int
	// Count of duplicate values
	DuplicateValues int
	// Cells with value (non-empty)
	WithValueCount int
	// Total number of cells
	CellCount int
	// Unique hints at start
	UniqueHints int
	// Hints at start
	Hints map[string]tallylogic.Hint
}
