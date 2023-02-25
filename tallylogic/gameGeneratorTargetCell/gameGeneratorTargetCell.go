package gamegenerator_target_cell

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic"
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

func uniqueUints(list []uint64) []uint64 {
	m := map[uint64]struct{}{}
	for _, v := range list {
		m[v] = struct{}{}
	}
	uniq := make([]uint64, len(m))
	var i int
	for k := range m {
		uniq[i] = k
		i++
	}
	return uniq
}

func (gen gameGeneratorTargetCell) GenerateGame() (tallylogic.Game, []tallylogic.Game, error) {
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
		solver := tallylogic.NewBruteSolver(options)
		solutions, err := solver.SolveGame(game)
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
			factors := primeFactors(n)
			uniquFactors := uniqueUints(factors)
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
	fmt.Println("new game")
	if err != nil {
		return game, fmt.Errorf("failed in gameGeneratorTargetCell.createGame: %w", err)
	}
	return game, nil
}

// func getSemiFactors(min, max, target uint64) []uint64 {
//   semiFactors := []
// }

func primeFactors(n uint64) []uint64 {
	switch n {
	case 0:
		return []uint64{}
	case 1:
	case 2:
	case 5:
	case 7:
	case 11:
		return []uint64{n}
	}

	var factors = []uint64{}
	var primeIndex int
	for {
		for primeIndex >= primeCount {
			createMorePrimes(100)
			primeCount = len(listOfPrimes)
		}
		prime := listOfPrimes[primeIndex]
		if prime == n {
			factors = append(factors, prime)
			return factors
		}
		mod := n % prime
		if mod > 0 {
			primeIndex++
			continue
		}
		n /= prime
		factors = append(factors, prime)
	}
}

func createMorePrimes(limit uint) {
	if limit == 0 {
		limit = 100
	}
	candidate := listOfPrimes[len(listOfPrimes)-1]
	if candidate < 3 {
		panic(fmt.Sprintf("candidate is too small: %v", candidate))
	}
	if candidate%2 == 0 {
		panic(fmt.Sprintf("candidate is divisible by 2"))
	}
	var i int
	var added uint
outer:
	for added <= limit {
		candidate += 2
		root := math.Sqrt(float64(candidate))
		for i = 0; i < len(listOfPrimes) && float64(listOfPrimes[i]) <= root; i++ {
			if candidate%listOfPrimes[i] == 0 {
				continue outer
			}
		}
		added++
		listOfPrimes = append(listOfPrimes, candidate)
	}
	primeCount = len(listOfPrimes)
}

var listOfPrimes = []uint64{
	2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 89, 97,
	101, 103, 107, 109, 113, 127, 131, 137, 139, 149, 151, 157, 163, 167, 173, 179, 181, 191, 193,
	197, 199, 211, 223, 227, 229, 233, 239, 241, 251, 257, 263, 269, 271, 277, 281, 283, 293, 307,
	311, 313, 317, 331, 337, 347, 349, 353, 359, 367, 373, 379, 383, 389, 397, 401, 409, 419, 421,
	431, 433, 439, 443, 449, 457, 461, 463, 467, 479, 487, 491, 499, 503, 509, 521, 523, 541, 547,
	557, 563, 569, 571, 577, 587, 593, 599, 601, 607, 613, 617, 619, 631, 641, 643, 647, 653, 659,
	661, 673, 677, 683, 691, 701, 709, 719, 727, 733, 739, 743, 751, 757, 761, 769, 773, 787, 797,
	809, 811, 821, 823, 827, 829, 839, 853, 857, 859, 863, 877, 881, 883, 887, 907, 911, 919, 929,
	937, 941, 947, 953, 967, 971, 977, 983, 991, 997, 1009, 1013, 1019, 1021, 1031, 1033, 1039, 1049,
	1051, 1061, 1063, 1069, 1087, 1091, 1093, 1097, 1103, 1109, 1117, 1123, 1129, 1151, 1153, 1163,
	1171, 1181, 1187, 1193, 1201, 1213, 1217, 1223, 1229, 1231, 1237, 1249, 1259, 1277, 1279, 1283,
	1289, 1291, 1297, 1301, 1303, 1307, 1319, 1321, 1327, 1361, 1367, 1373, 1381, 1399, 1409, 1423,
	1427, 1429, 1433, 1439, 1447, 1451, 1453, 1459, 1471, 1481, 1483, 1487, 1489, 1493, 1499, 1511,
	1523, 1531, 1543, 1549, 1553, 1559, 1567, 1571, 1579, 1583, 1597, 1601, 1607, 1609, 1613, 1619,
	1621, 1627, 1637, 1657, 1663, 1667, 1669, 1693, 1697, 1699, 1709, 1721, 1723, 1733, 1741, 1747,
	1753, 1759, 1777, 1783, 1787, 1789, 1801, 1811, 1823, 1831, 1847, 1861, 1867, 1871, 1873, 1877,
	1879, 1889, 1901, 1907, 1913, 1931, 1933, 1949, 1951, 1973, 1979, 1987, 1993, 1997, 1999, 2003,
	2011, 2017, 2027, 2029, 2039, 2053, 2063, 2069, 2081, 2083, 2087, 2089, 2099, 2111, 2113, 2129,
	2131, 2137, 2141, 2143, 2153, 2161, 2179, 2203, 2207, 2213, 2221, 2237, 2239, 2243, 2251, 2267,
	2269, 2273, 2281, 2287, 2293, 2297, 2309, 2311, 2333, 2339, 2341, 2347, 2351, 2357, 2371, 2377,
	2381, 2383, 2389, 2393, 2399, 2411, 2417, 2423, 2437, 2441, 2447, 2459, 2467, 2473, 2477, 2503,
	2521, 2531, 2539, 2543, 2549, 2551, 2557, 2579, 2591, 2593, 2609, 2617, 2621, 2633, 2647, 2657,
	2659, 2663, 2671, 2677, 2683, 2687, 2689, 2693, 2699, 2707, 2711, 2713, 2719, 2729, 2731, 2741,
	2749, 2753, 2767, 2777,
}
var (
	primeCount int = len(listOfPrimes)
)
