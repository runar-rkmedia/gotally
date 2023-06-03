package tallylogic

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
	"github.com/runar-rkmedia/gotally/tallylogic/cellgenerator"
	"golang.org/x/net/context"
)

type GameGeneratorOptions struct {
	Rows, Columns       int
	GoalChecker         GoalChecker `toml:"-"`
	TargetCellValue     uint64
	MaxBricks           int
	MinBricks           int
	MinMoves            int
	MaxMoves            int
	MaxIterations       int
	Concurrency         int
	CellGenerator       CellGenerator `toml:"-"`
	Randomizer          Randomizer    `toml:"-"`
	Seed                uint64
	MinGames            int
	GameSolutionChannel chan SolvableGame `toml:"-"`
}

type GameGenerator struct {
	GameGeneratorOptions
}

type Randomizer interface {
	Int63n(n int64) int64
	Intn(n int) int
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

func NewGameGenerator(options GameGeneratorOptions) (gb GameGenerator, err error) {
	if options.GameSolutionChannel == nil {
		err = fmt.Errorf("channel empty")
		return
	}
	if options.Concurrency == 0 {
		options.Concurrency = 2000
	}
	if options.MaxIterations == 0 {
		options.MaxIterations = 1_000_000
	}
	min, err := getRequiredCellCount(1, 12, options.TargetCellValue)
	if err != nil {
		return GameGenerator{}, fmt.Errorf("cannot create game: %w", err)
	}
	if options.MinBricks <= int(min) {
		options.MinBricks = int(min)
	}
	gb.GameGeneratorOptions = options
	if options.Columns < 1 {
		err = fmt.Errorf("Columns must be a non-zero positive number")
		return
	}
	if options.Rows < 1 {
		err = fmt.Errorf("Rows must be a non-zero positive number")
		return
	}
	length := options.Columns * options.Rows
	if options.MaxBricks > length {
		err = fmt.Errorf("MaxBricks cannot be more that total cell-count (columns * rows)")
		return
	}
	if options.GoalChecker == nil {
		err = fmt.Errorf("GoalChecker must be set")
		return
	}
	r := randomizer.NewRandomizer(options.Seed)
	if options.CellGenerator == nil {
		gb.CellGenerator = cellgenerator.NewCellGenerator(r)
	}
	if options.Randomizer == nil {
		gb.Randomizer = r
	}
	return
}

// generateGame randomly generates a new board.
func (gb GameGenerator) generateGame() Game {
	template := NewGameTemplate(GameModeRandomChallenge, "gen", "generated", "", gb.Rows, gb.Columns)
	template.GoalChecker = gb.GoalChecker
	template.Board.cells = gb.GenerateBoardValues()
	game, err := NewGame(GameModeTutorial, template)
	if err != nil {
		panic(err)
	}
	return game
}

// Does a quick check to see if a game is solvable to a certain target value
// Ignoring the order of any of the values
func (gb GameGenerator) isUnsolvableQuickCheck(values []uint64, targetValue uint64) bool {
	// This implementation needs to be fast about eliminating games
	targetDoubled := targetValue * 2
	var multiples uint64 = 1
	for i := 0; i < len(values); i++ {
		if values[i] >= targetValue {
			return false
		}
		if values[i] == 0 {
			continue
		}
		multiples *= values[i]
		if multiples >= targetDoubled {
			return false
		}
	}
	// We can eliminate this safely
	if multiples < targetDoubled {
		// fmt.Println("multiple below", multiples, targetValue, values)
		return true
	}
	// check multiples

	return false
}

// deprecated
// DEV-hack for now, see #23 at https://github.com/runar-rkmedia/gotally/issues/23
func send[T any](name string, ch chan T, msg T) bool {
	ch <- msg
	return true
	// Originial non-blocking send:
	// select {
	// case ch <- msg:
	// 	return true
	// default:
	// 	return false
	// }

}

// GenerateGame randomly generates a new board that is solvable within the requirements set
func (gb GameGenerator) GenerateGame(ctx context.Context) (Game, []Game, error) {
	options := GameSolverFactoryOptions{
		// BreadthFirst: true,
		SolveOptions: SolveOptions{
			MinMoves:     gb.MinMoves,
			MaxMoves:     gb.MaxMoves,
			MaxSolutions: 1,
		},
	}
	solver := GameSolverFactory(options)

	ch := make(chan SolvableGame, 100)
	jobs := make(chan Game, gb.Concurrency)
	doneCh := make(chan struct{})
	quit := make(chan struct{})
	quit2 := make(chan struct{})
	errorsCh := make(chan string)
	errors := map[string]int{}
	start := time.Now()
	solvableGames := 0
	done := 0
	errorCount := 0
	go func() {
		for {
			select {
			case <-ctx.Done():

				return
			case <-quit:
				send("q3", quit2, struct{}{})
				return
			case game := <-jobs:
				go func(game Game) {
					sb, err := gb.solveGame(solver, game, quit2)
					if err != nil {
						errorsCh <- err.Error()
						return

					}
					if sb != nil {
						ch <- *sb
						send("q4", quit2, struct{}{})
						return
					} else {
						ok1 := send("q1", quit2, struct{}{})
						ok2 := send("q2", doneCh, struct{}{})
						if !ok1 || !ok2 {
							return
						}
					}
				}(game)
			}
		}
	}()
	jobHash := map[string]struct{}{}
	skipped := 0
	total := 0
	generateJob := func() {
		// jobs <- gb.generateGame()
		// return

		for {
			game := gb.generateGame()
			total++
			cells := game.Cells()
			cellvalues := make([]uint64, len(cells))
			for i := 0; i < len(cells); i++ {
				cellvalues[i] = uint64(cells[i].Value())
			}
			if gb.TargetCellValue > 0 {
				unsolvable := gb.isUnsolvableQuickCheck(cellvalues, gb.TargetCellValue)
				if unsolvable {
					skipped++
					total++
					// continue
				}
			}
			// fmt.Printf("skipped %d of %d | %.2f\n", skipped, total, float64(skipped)/float64(total))

			hash := game.board.Hash()
			if _, ok := jobHash[hash]; !ok {
				jobHash[hash] = struct{}{}
				jobs <- game
				return
			}
		}
	}

	for i := 0; i < gb.Concurrency; i++ {
		generateJob()
	}
	writer := os.Stdout
	printStatus := func() {
		total := done + errorCount
		sinceStart := time.Since(start)
		ratePerSecond := float64(total) / float64(sinceStart/time.Second)
		perc := float64(total) / float64(gb.MaxIterations)
		expectedToBeDone := time.Duration(float64(gb.MaxIterations)/ratePerSecond) * time.Second
		expectedToBeDoneAt := start.Add(expectedToBeDone)
		uniques := len(jobHash)
		if sinceStart > time.Second {

			_, err := writer.WriteString(
				fmt.Sprintf("[%5.1f%% in % 12s (%s)] % 8.2f g/s. Found: % 6d Unique: % 12d Skipped: % 12d, Failure: % 12.1f%%, ErrorMap: %v\n",
					perc*100,
					(expectedToBeDone - sinceStart).Round(100*time.Millisecond).String(),
					expectedToBeDoneAt.Format("15:04:05"),
					ratePerSecond,
					solvableGames,
					uniques,
					skipped,
					float64(errorCount)/float64(total)*100,
					errors),
			)
			if err != nil {
				panic(err)
			}
		}
	}
	ticker := time.NewTicker(time.Millisecond * 500)
outer:
	for {
		select {
		case <-ctx.Done():
			break outer

		case <-ticker.C:
			printStatus()
		case sg := <-ch:
			solvableGames++
			gb.GameSolutionChannel <- sg
			if solvableGames >= gb.MinGames {
				quit <- struct{}{}
				return sg.Game, sg.Solutions, nil
			}
		case errMsg := <-errorsCh:
			errorCount++
			errors[errMsg]++
			if (done + errorCount) > gb.MaxIterations {
				quit <- struct{}{}
				return Game{}, []Game{}, fmt.Errorf("Too many errors %v", errors)
			}
			generateJob()
		case <-doneCh:
			done++
			if (done + errorCount) > gb.MaxIterations {
				return Game{}, []Game{}, fmt.Errorf("no games found")
			}
			generateJob()
		}
	}
	return Game{}, nil, ctx.Err()
}

type SolvableGame struct {
	GeneratorOptions GameGeneratorOptions
	Game
	Solutions []Game
}

func (gb GameGenerator) solveGame(solver Solver, game Game, quitCh chan struct{}) (*SolvableGame, error) {
	solutions, err := solver.SolveGame(game, quitCh)
	if err != nil {
		return nil, err
	}
	if len(solutions) > 0 {
		return &SolvableGame{gb.GameGeneratorOptions, game, solutions}, nil
	}
	return nil, nil
}

func (gb GameGenerator) GenerateBoardValues() []cell.Cell {

	length := gb.Columns * gb.Rows
	board := make([]cell.Cell, length)

	bricks := gb.Randomizer.Intn(gb.MaxBricks-gb.MinBricks) + gb.MinBricks
	for i := 0; i < length; i++ {
		board[i] = cell.NewCell(0, 0)
	}
	for i := 0; i < bricks; i++ {
		index := gb.Randomizer.Intn(length - 1)
		for !board[index].IsEmpty() {
			index = gb.Randomizer.Intn(length - 1)
		}
		board[index] = gb.CellGenerator.GeneratePure()
	}

	return board
}

type GeneratedGame struct {
	GeneratorOptions GameGeneratorOptions
	Solutions        []GeneratedSolution
	Name             string
	Preview          string `toml:",multiline,literal"`
	Hash             string
	Cells            []int64
	Stats            GameStats
	SolutionStats    SolutionStats
}

type GeneratedSolution struct {
	History          CompactHistory
	HighestCellValue int64
	Score            int64
	Moves            int
	VisualSolution   string `toml:",multiline,literal"`
}
