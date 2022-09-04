package generated

import (
	"embed"
	"fmt"
	"io/fs"

	"github.com/pelletier/go-toml/v2"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

//go:embed games
var GenDir embed.FS

var GeneratedTemplates []tallylogic.GameTemplate

func ReadGeneratedBoardsFromDisk() error {
	// generatorDir := path.Join("./generated")
	// generatorDir := generated.GenDir
	err := fs.WalkDir(GenDir, ".", func(p string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := fs.ReadFile(GenDir, p)
		if err != nil {
			return err
		}
		var gen tallylogic.GeneratedGame
		err = toml.Unmarshal(b, &gen)
		if err != nil {
			return err
		}
		// if gen.GeneratorOptions.Rows != 3 {
		// 	return nil
		// }
		var description string
		if len(gen.Solutions) > 0 {
			description = fmt.Sprintf("Get at least one cell to a value of %d. This game can be solved in %d moves, with the highest cell at %d", gen.GeneratorOptions.TargetCellValue, gen.Solutions[0].Moves, gen.Solutions[0].HighestCellValue)

		}
		template := tallylogic.NewGameTemplate(gen.Hash, gen.Name, description, gen.GeneratorOptions.Columns, gen.GeneratorOptions.Rows).
			SetGoalCheckerLargestValue(int64(gen.GeneratorOptions.TargetCellValue)).
			SetMaxMoves(gen.GeneratorOptions.MaxMoves).
			SetStartingLayout(gen.Cells...)

		GeneratedTemplates = append(GeneratedTemplates, *template)
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}
