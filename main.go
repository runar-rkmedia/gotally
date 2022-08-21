package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"

	"github.com/pelletier/go-toml/v2"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func main() {
	go func() {
		address := "localhost:6060"
		fmt.Println("pprof available at", address)
		fmt.Println(http.ListenAndServe(address, nil))
	}()
	// getCell()
	generateGame()
}

func getCell() {
	args := flag.Args()
	flag.Parse()
	fmt.Println(args)
	a, err := strconv.ParseInt(flag.Arg(0), 10, 64)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseInt(flag.Arg(1), 10, 64)
	if err != nil {
		panic(err)
	}
	cell := tallylogic.NewCell(a, int(b))
	fmt.Println(cell)

}
func generateGame() {
	type options struct {
		Rows, Columns, TargetCellValue, MaxBricks, MinBricks, MinMoves, MaxMoves, Concurrency, MaxIterations, MinGames int
	}
	f, err := os.ReadFile("./generator-config.toml")
	if err != nil {
		panic(err)
	}
	var op options
	err = toml.Unmarshal(f, &op)
	if err != nil {
		panic(err)
	}

	// op := options{}
	// optionsToml, err := toml.Marshal(op)

	// os.WriteFile("./generator-config.toml", optionsToml, 0677)
	gameCh := make(chan tallylogic.SolvableGame, 8)
	quit := make(chan struct{})

	go func() {
		fmt.Println("listening for games")
		for {
			select {
			case <-quit:
				fmt.Println("quitting")
				return
			case sg := <-gameCh:
				fmt.Println("got a gammmmmmmme", sg.Game.Print())
				cells := sg.Cells()
				out := tallylogic.GeneratedGame{
					GeneratorOptions: sg.GeneratorOptions,
					Preview:          sg.Print(),
					Hash:             sg.Hash(),
					Cells:            make([]int64, len(cells)),
				}
				for i, c := range cells {
					out.Cells[i] = c.Value()
				}
				buf := bytes.Buffer{}
				e := toml.NewEncoder(&buf)

				err := e.Encode(out)
				if err != nil {
					panic(err)
				}
				fmt.Println("buf", buf.String())
				dir := path.Join(
					"./",
					"generated",
					"games",
					fmt.Sprintf("%dx%d-target-%d-moves-%d", op.Columns, op.Rows, op.TargetCellValue, sg.Solutions[0].Moves()),
				)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					panic(err)
				}
				fp := path.Join(dir, sg.Game.Hash()+".toml")
				os.WriteFile(fp, buf.Bytes(), 0755)

			}

		}
	}()
	fmt.Printf("\n%#v", op)

	gb, err := tallylogic.NewGameGenerator(tallylogic.GameGeneratorOptions{
		GameSolutionChannel: gameCh,
		Rows:                op.Rows,
		Columns:             op.Columns,
		GoalChecker: tallylogic.GoalCheckLargestCell{
			TargetCellValue: int64(op.TargetCellValue),
		},
		TargetCellValue: uint64(op.TargetCellValue),
		MaxBricks:       op.MaxBricks,
		MinBricks:       op.MinBricks,
		MaxMoves:        op.MaxMoves,
		MinMoves:        op.MinMoves,
		MaxIterations:   op.MaxIterations,
		MinGames:        op.MinGames,
		Concurrency:     op.Concurrency,
		// CellGenerator: nil,
		// Randomizer:    rand.
		Randomizer: tallylogic.NewRandomizer(),
	})

	if err != nil {
		panic("gb: " + err.Error())
	}
	game, solutions, err := gb.GenerateGame()
	quit <- struct{}{}
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("\nGenerated a board with %d solutions %s\n", len(solutions), game.Print())
	fmt.Println("Games solved in moves:")
	for _, s := range solutions {
		fmt.Printf(" %d", s.Moves())
	}
	fmt.Println("")

}
