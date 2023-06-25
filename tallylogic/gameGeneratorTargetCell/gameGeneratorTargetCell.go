package gamegenerator_target_cell

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type GameGeneratorTargetCellOptions struct {
	TargetCell   uint64
	MinCellValue uint64
	MaxCellValue uint64
	// out of 120, how likely is it that the generator should add randomized cells to the board
	// The default is 12 (or 10%). Use a negative value to disable
	RandomCellChance   int
	MaxCells           int
	MaxAdditionalCells int
	Rows               int
	Columns            int
	MaxMoves           int
	MinMoves           int
	Seed               uint64
	Randomizer         tallylogic.Randomizer
	cellNeeded         []uint64
}
type gameGeneratorTargetCell struct {
	GameGeneratorTargetCellOptions
}

// getRequiredCellsHighestForm returns a list of numbers from the target.
// The list of numbers generated are the numbers that TallyBoard would need
// to have on the board, to be able to combine into the target.
//
// Based on the rules of TallyBoard, it is only possible to double one number.
// Therefore, the returned list is typically the half of the target
//
// TODO: improve the naming here
func getRequiredCellsHighestForm(min, max, target uint64) ([]uint64, error) {
	if target < min {
		return []uint64{}, fmt.Errorf("target cannot be smaller than the minimum value")
	}
	if target <= max {
		return []uint64{target}, nil
	}
	if target%2 != 0 {
		return []uint64{}, fmt.Errorf("target is above max-value, and not divisble by 2")
	}
	cellsNeeded := []uint64{}
	x := target
	for {
		if x <= max {
			break
		}
		x = x / 2
		cellsNeeded = append(cellsNeeded, x)
	}
	// duplicate the last cell, so that the we have a starting point
	cellsNeeded = append(cellsNeeded, cellsNeeded[len(cellsNeeded)-1])
	return cellsNeeded, nil
}
func getRequiredCellCount(min, max, target uint64) (uint64, error) {
	if target < min {
		return 0, fmt.Errorf("target cannot be smaller than the minimum value")
	}
	if target <= max {
		return 0, nil
	}
	if target%2 != 0 {
		return 0, fmt.Errorf("target is above max-value, and not divisble by 2")
	}

	f := math.Log(float64(target)/float64(max)) / math.Log(2)
	return uint64(math.Ceil(f)) + 1, nil
}

func NewGameGeneratorForTargetCell(options GameGeneratorTargetCellOptions) (gen gameGeneratorTargetCell, err error) {
	gen.GameGeneratorTargetCellOptions = options

	if gen.Randomizer == nil {
		gen.Randomizer = randomizer.NewRandomizer(options.Seed)
	}
	if gen.Rows <= 0 {
		return gen, fmt.Errorf("rows must be positive")
	}
	if gen.MaxMoves <= 0 {
		return gen, fmt.Errorf("maxMoves must be positive")
	}
	if gen.Columns <= 0 {
		return gen, fmt.Errorf("columns must be positive")
	}
	if gen.MinCellValue == 0 {
		gen.MinCellValue = 1
	}
	if gen.MaxCellValue == 0 {
		gen.MaxCellValue = 12
	}
	if gen.RandomCellChance == 0 {
		gen.RandomCellChance = 12
	}
	if gen.MaxCells > gen.Rows*gen.Columns {
		return gen, fmt.Errorf("Max-cells cannot be higher than the board-size")
	}
	if gen.MaxAdditionalCells < -1 {
		return gen, fmt.Errorf("maxAdditionalCells must bo non-negative. To Disable the behavior, use -1 as value")
	}
	if gen.TargetCell <= 0 {
		return gen, fmt.Errorf("targetcell must be positive")
	}
	c, err := getRequiredCellsHighestForm(gen.MinCellValue, gen.MaxCellValue, gen.TargetCell)
	if err != nil {
		return gen, fmt.Errorf("failed to calculate the required cells: %w", err)
	}
	gen.cellNeeded = c

	return
}

func intMinNonZero(ints ...int) (int, bool) {
	smallest := math.MaxInt
	ok := false
	for _, v := range ints {
		if v == 0 {
			continue
		}
		if v < smallest {
			ok = true
			smallest = v
		}
	}
	return smallest, ok
}
func intMaxNonZero(ints ...int) (int, bool) {
	biggest := -math.MaxInt
	ok := false
	for _, v := range ints {
		if v == 0 {
			continue
		}
		if v > biggest {
			ok = true
			biggest = v
		}
	}
	return biggest, ok
}

func (gen gameGeneratorTargetCell) GenerateGame(ctx context.Context) (tallylogic.Game, []tallylogic.Game, error) {
	for i := 0; i < 1000; i++ {
		game, err := gen.generateGame()
		if err != nil {
			return game, nil, err
		}
		// Ensure that the game can be solved.
		options := tallylogic.SolveOptions{
			MinMoves:     gen.MinMoves,
			MaxMoves:     gen.MaxMoves,
			MaxSolutions: 1,
			MaxTime:      time.Millisecond * 100,
		}
		_, solutions, err := tallylogic.SolveGame(options, game, nil)
		if err != nil {
			switch {
			case errors.Is(err, context.DeadlineExceeded):
				// ignore deadline, game is most likely not solveable
				// it is cheaper to just generate a new game
			default:
				return game, nil, fmt.Errorf("failed to create solution while generating game: %w", err)
			}
		}
		if solutions == nil || len(solutions) == 0 {
			continue
		}
		return game, solutions, err
	}
	return tallylogic.Game{}, nil, fmt.Errorf("too many retries")
}
func (gen gameGeneratorTargetCell) generateGame() (tallylogic.Game, error) {
	cellsForBoard := gen.Rows * gen.Columns
	// copy the cellsNeeded
	cellsNeeded := append([]uint64{}, gen.cellNeeded...)

	if len(cellsNeeded) > cellsForBoard {
		return tallylogic.Game{}, fmt.Errorf(
			"cannot create game for targetCellvValue %d, for a %dx%d-board, as it would require at least %d cells",
			gen.TargetCell, gen.Columns, gen.Rows, len(cellsNeeded))
	}
	if gen.MaxCells > 0 && len(cellsNeeded) > int(gen.MaxCells) {
		return tallylogic.Game{}, fmt.Errorf(
			"cannot create game for targetCellvValue %d, when MaxCells is set to %d. Requires at least %d",
			gen.TargetCell, gen.MaxCells, len(cellsNeeded))
	}

	var additionalCells int
	if gen.MaxAdditionalCells != -1 {
		// Calculate if and how many additional cells we can replace with an equal sum/product to create additional randomness.
		// For instance, given the number 12, we can replace it with for instance 3 and 4 ( 3 * 4 = 12 ), or 5 and 7 ( 5 + 7 = 12)
		// This must still fit within the bounds of the board, and the requirements set.
		max, ok := intMaxNonZero(cellsForBoard, gen.MaxCells)
		if !ok {
			return tallylogic.Game{}, fmt.Errorf("non-ok ints max %d", max)
		}

		maxAdditionalCells, ok := intMinNonZero(max-len(cellsNeeded), gen.MaxAdditionalCells)
		if !ok {
			return tallylogic.Game{}, fmt.Errorf("non-ok ints min %d", maxAdditionalCells)
		}

		if maxAdditionalCells > 0 && gen.MaxAdditionalCells != -1 {
			// TODO: we need to distrubute this better
			additionalCells = gen.Randomizer.Intn(maxAdditionalCells)
		}
	}
	if additionalCells > 0 {
		replacedCount := 0

		for replacedCount < additionalCells {
			if gen.RandomCellChance > 0 && gen.Randomizer.Intn(120) < gen.RandomCellChance {
				// Insert a random cell
				for {
					randomcell := uint64(gen.Randomizer.Int63n(int64(gen.MaxCellValue-gen.MinCellValue))) + gen.MinCellValue
					if randomcell == 0 {
						continue
					}
					cellsNeeded = append(cellsNeeded, randomcell)
					replacedCount++
					break
				}
			}

			// Replace one cell at a random index by values that either adds up to, or multiplies to the value at the index
			replaceIndex := gen.Randomizer.Intn(len(cellsNeeded))
			n := cellsNeeded[replaceIndex]
			// var divisor int
			switch {
			// not possible to do replacement here
			case n == 1:
				continue
			}
			// Find a factor that we can create a replacement with
			// for instance, given the number 42, which has factors 2, 3  and 7
			// we randomly select either 2, 3 or 7
			// in this example, we select 7.
			// We replace the value at the replaceIndex with 7
			// and then we add the value of 42 / 7 = 6
			factors := cell.NewFactors(n)
			uniquFactors := factors.UniqueFactors()
			randomFactorIndex := gen.Randomizer.Intn(len(uniquFactors))
			randomFactor := uniquFactors[randomFactorIndex]
			cellsNeeded[replaceIndex] = uint64(randomFactor)
			// cellsNeeded = append(cellsNeeded, n/uint64(randomFactor))
			cellsNeeded = append(cellsNeeded, 2)

			replacedCount++
		}
	}

	// place the cells randomly on the board
	cellIndexMap := make(map[int]uint64)
	for i := 0; i < len(cellsNeeded); i++ {
		for {
			r := gen.Randomizer.Intn(cellsForBoard)
			if _, ok := cellIndexMap[r]; ok {
				continue
			}
			cellIndexMap[r] = cellsNeeded[i]
			break
		}
	}
	cells := make([]uint64, cellsForBoard)
	for i, v := range cellIndexMap {
		cells[i] = v
	}

	game, err := gen.createGame(cells...)
	return game, err

}

func (gen gameGeneratorTargetCell) createGame(cellValues ...uint64) (tallylogic.Game, error) {

	template := tallylogic.NewGameTemplate(tallylogic.GameModeRandomChallenge, "gen", "generated", "", gen.Rows, gen.Columns).
		SetStartingLayoutUints(cellValues...).
		SetMaxMoves(gen.MaxMoves).
		SetGoalCheckerLargestValue(gen.TargetCell)
	game, err := tallylogic.NewGame(tallylogic.GameModeRandomChallenge, template)
	if err != nil {
		return game, fmt.Errorf("failed in gameGeneratorTargetCell.createGame: %w", err)
	}
	return game, nil
}
