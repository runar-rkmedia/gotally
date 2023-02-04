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

type Options struct {
	MaxItems int
}

func ReadGeneratedBoardsFromDisk(options ...Options) error {
	o := Options{}
	for _, x := range options {
		if x.MaxItems != 0 {
			o.MaxItems = x.MaxItems
		}
	}
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
		template := tallylogic.NewGameTemplate(tallylogic.GameModeRandomChallenge, gen.Hash, gen.Name, description, gen.GeneratorOptions.Columns, gen.GeneratorOptions.Rows).
			SetGoalCheckerLargestValue(gen.GeneratorOptions.TargetCellValue).
			SetMaxMoves(gen.GeneratorOptions.MaxMoves).
			SetStartingLayout(gen.Cells...)

		GeneratedTemplates = append(GeneratedTemplates, *template)
		if o.MaxItems > 0 && len(GeneratedTemplates) >= o.MaxItems {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}
