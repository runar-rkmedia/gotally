package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"

	"github.com/pelletier/go-toml/v2"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/gotally/randomizer"
	"github.com/runar-rkmedia/gotally/tallylogic"
	"github.com/runar-rkmedia/skiver/utils"
)

var log logger.AppLogger

// used to generate games and save them to disk.
func main() {
	logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: true,
	})
	log = logger.GetLogger("main")

	go func() {
		address := "localhost:6060"
		log.Info().Str("address", address).Msg("pprof available")
		log.Fatal().
			Str("address", address).
			Err(http.ListenAndServe(address, nil)).
			Msg(("failed setting up listener"))
	}()
	generateGame()
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

	gameCh := make(chan tallylogic.SolvableGame, 8)
	quit := make(chan struct{})

	go func() {
		log.Info().Msg("listening for games")
		for {
			select {
			case <-quit:
				log.Info().Msg("quitting")
				return
			case sg := <-gameCh:
				cells := sg.Cells()
				hashName, _ := utils.GetRandomName()
				out := tallylogic.GeneratedGame{
					GeneratorOptions: sg.GeneratorOptions,
					Solutions:        make([]tallylogic.GeneratedSolution, len(sg.Solutions)),
					Name:             hashName,
					Preview:          sg.Print(),
					Hash:             sg.Hash(),
					Cells:            make([]int64, len(cells)),
				}
				stats, err := sg.Stats()
				if err != nil {
					panic("Failed to generate gamestats: " + err.Error())
				}
				out.Stats = stats
				solutionStats, err := tallylogic.NewSolutionsStats(sg.Game, sg.Solutions)
				if err != nil {
					panic("Failed to generate solutionStats: " + err.Error())
				}
				out.SolutionStats = solutionStats
				for i, c := range cells {
					out.Cells[i] = c.Value()
				}
				for i, s := range sg.Solutions {
					out.Solutions[i] = tallylogic.GeneratedSolution{
						History:          s.History,
						HighestCellValue: s.HighestCellValue(),
						Score:            s.Score(),
						Moves:            s.Moves(),
					}
					gameCopy := sg.Copy()
					for _, ins := range s.History {
						t := tallylogic.GetInstructionType(ins)
						switch t {
						case tallylogic.InstructionTypeCombinePath, tallylogic.InstructionTypeSelectCoord, tallylogic.InstructionTypeSelectIndex:
							path, ok := tallylogic.GetInstructionAsPath(ins)
							if !ok {
								panic("The path was not the expected type")
							}
							s := "\n" + gameCopy.DescribePath(path)
							s += gameCopy.PrintForSelectionNoColor(path) + "\n"
							gameCopy.Instruct(ins)
							out.Solutions[i].VisualSolution += s
						case tallylogic.InstructionTypeSwipe:
							dir, ok := tallylogic.GetInstructionAsSwipe(ins)
							if !ok {
								fmt.Printf("%#v\n", ins)
								panic("invalid swipe-direction")
							}
							gameCopy.Instruct(ins)
							out.Solutions[i].VisualSolution += "\n" + string(dir) + gameCopy.Print() + "\n"
						}
					}
					out.Solutions[i].VisualSolution += "\nEnd: \n" + gameCopy.Print()
				}
				buf := bytes.Buffer{}
				e := toml.NewEncoder(&buf)

				err = e.Encode(out)
				if err != nil {
					panic(err)
				}
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
				fp := path.Join(dir, hashName+"_"+sg.Game.Hash()+".toml")
				err = os.WriteFile(fp, buf.Bytes(), 0755)
				if err != nil {
					panic(err)
				}

			}

		}
	}()

	gb, err := tallylogic.NewGameGenerator(tallylogic.GameGeneratorOptions{
		GameSolutionChannel: gameCh,
		Rows:                op.Rows,
		Columns:             op.Columns,
		GoalChecker: tallylogic.GoalCheckLargestCell{
			TargetCellValue: uint64(op.TargetCellValue),
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
		Randomizer: randomizer.NewSeededRandomizer(),
	})

	if err != nil {
		panic("gb: " + err.Error())
	}
	game, solutions, err := gb.GenerateGame(context.TODO())
	quit <- struct{}{}
	if err != nil {
		panic(err.Error())
	}
	log.Info().Str("game", game.Print()).Int("solutions", len(solutions)).Msg("Generated a board with solutions")

}
