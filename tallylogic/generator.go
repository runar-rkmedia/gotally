package tallylogic

import (
	"fmt"
	"os"
	"time"

	"github.com/runar-rkmedia/gotally/randomizer"
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
	if options.MinBricks <= 0 {
		options.MinBricks = 1
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
		gb.CellGenerator = NewCellGenerator(r)
	}
	if options.Randomizer == nil {
		gb.Randomizer = r
	}
	return
}

// generateGame randomly generates a new board.
func (gb GameGenerator) generateGame() Game {
	template := NewGameTemplate("gen", "generated", "", gb.Rows, gb.Columns)
	template.GoalChecker = gb.GoalChecker
	template.Board.cells = gb.GenerateBoardValues()
	game, err := NewGame(GameModeTemplate, template)
	if err != nil {
		panic(err)
	}
	return game
}

// GenerateGame randomly generates a new board that is solvable within the requirements set
func (gb GameGenerator) GenerateGame() (Game, []Game, error) {
	options := SolveOptions{
		MinMoves:     gb.MinMoves,
		MaxMoves:     gb.MaxMoves,
		MaxSolutions: 1,
	}
	solver := NewBruteSolver(options)

	ch := make(chan SolvableGame, 100)
	jobs := make(chan Game, gb.Concurrency)
	doneCh := make(chan struct{})
	quit := make(chan struct{})
	errorsCh := make(chan string)
	errors := map[string]int{}
	start := time.Now()
	solvableGames := 0
	done := 0
	errorCount := 0
	go func() {
		for {
			select {
			case <-quit:
				return
			case game := <-jobs:
				go func(game Game) {
					sb, err := gb.solveGame(solver, game)
					if err != nil {
						errorsCh <- err.Error()
						return

					}
					if sb != nil {
						ch <- *sb
					} else {
						doneCh <- struct{}{}
					}
				}(game)
			}
		}
	}()
	jobHash := map[string]struct{}{}
	generateJob := func() {
		// jobs <- gb.generateGame()
		// return

		for {
			game := gb.generateGame()
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
		_, err := writer.WriteString(
			fmt.Sprintf("\n[%5.1f%% in %s (%s)] %.2f g/s. Unique: %d, Failure: %5.1f%%, ErrorMap: %v",
				perc*100,
				(expectedToBeDone - sinceStart).Round(100*time.Millisecond).String(),
				expectedToBeDoneAt.Format("15:04:05"),
				ratePerSecond,
				uniques,
				float64(errorCount)/float64(total)*100,
				errors),
		)
		if err != nil {
			panic(err)
		}
	}
	ticker := time.NewTicker(time.Millisecond * 500)
	for {
		select {
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
}

type SolvableGame struct {
	GeneratorOptions GameGeneratorOptions
	Game
	Solutions []Game
}

func (gb GameGenerator) solveGame(solver bruteSolver, game Game) (*SolvableGame, error) {
	solutions, err := solver.SolveGame(game)
	if err != nil {
		return nil, err
	}
	if len(solutions) > 0 {
		return &SolvableGame{gb.GameGeneratorOptions, game, solutions}, nil
	}
	return nil, nil
}

func (gb GameGenerator) GenerateBoardValues() []Cell {

	length := gb.Columns * gb.Rows
	board := make([]Cell, length)

	bricks := gb.Randomizer.Intn(gb.MaxBricks-gb.MinBricks) + gb.MinBricks
	for i := 0; i < length; i++ {
		board[i] = NewCell(0, 0)
	}
	for i := 0; i < bricks; i++ {
		index := gb.Randomizer.Intn(length - 1)
		for board[index].baseValue != 0 {
			index = gb.Randomizer.Intn(length - 1)
		}
		board[index] = gb.CellGenerator.Generate()
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
}

type GeneratedSolution struct {
	History          []any
	HighestCellValue int64
	Score            int64
	Moves            int
}
