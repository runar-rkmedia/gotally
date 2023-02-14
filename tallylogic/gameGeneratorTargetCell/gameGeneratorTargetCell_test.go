package gamegenerator_target_cell

import (
	"testing"
	"time"

	"github.com/MarvinJWendt/testza"
	"github.com/go-test/deep"
	"github.com/runar-rkmedia/gotally/tallylogic"
)

func Test_getRequiredCellsHighestForm(t *testing.T) {
	type args struct {
		min    uint64
		max    uint64
		target uint64
	}
	tests := []struct {
		name    string
		args    args
		want    []uint64
		wantErr bool
	}{
		{
			"Should return the required cells in their highest form",
			args{min: 0, max: 12, target: 768},
			[]uint64{384, 192, 96, 48, 24, 12, 12},
			false,
		},
		{
			"Should err on target below min",
			args{min: 4, max: 12, target: 3},
			[]uint64{},
			true,
		},
		{
			"Should err on target non-divisable by two above max",
			args{min: 4, max: 12, target: 303},
			[]uint64{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getRequiredCellsHighestForm(tt.args.min, tt.args.max, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("getNeededCellsHighestForm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("getNeededCellsHighestForm() = %v, want %v\ndiff: %#v", got, tt.want, diff)
			}
		})
	}
}
func init() {
	testza.SetShowStartupMessage(false)
}
func Test_createMorePrimes(t *testing.T) {
	tests := []struct {
		name  string
		limit uint
		want  []uint64
	}{
		{
			"Should create some primes",
			10,
			[]uint64{2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41},
		},
	}
	for _, tt := range tests {
		listOfPrimes = []uint64{2, 3}
		t.Run(tt.name, func(t *testing.T) {

			// in case of an infinite loop
			testza.AssertCompletesIn(t, time.Second, func() { createMorePrimes(tt.limit) })

			if diff := deep.Equal(listOfPrimes, tt.want); diff != nil {
				t.Errorf("getNeededCellsHighestForm() = %v, want %v\ndiff: %#v", listOfPrimes, tt.want, diff)
			}
		})
	}
}

func Test_gameGeneratorTargetCell_GenerateGame(t *testing.T) {
	tests := []struct {
		name          string
		options       GameGeneratorTargetCellOptions
		solutionCount int
		wantErr       bool
	}{
		{
			"Should generate game for 768 on a 5x5",
			GameGeneratorTargetCellOptions{
				TargetCell:       768,
				Rows:             5,
				Columns:          5,
				RandomCellChance: -1,
				MaxMoves:         120,
			},
			1,
			false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewGameGeneratorForTargetCell(tt.options)
			if err != nil {
				t.Fatalf("failed to create game: %v", err)
			}
			if s, ok := gen.Randomizer.(tallylogic.IntRandomizer); ok {
				if false {
					s.SetSeed(0, 0)
				}

			}

			_, solutions, err := gen.GenerateGame()
			if (err != nil) != tt.wantErr {
				t.Errorf("gameGeneratorTargetCell.GenerateGame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(solutions) != tt.solutionCount {
				t.Fatalf("Expected %d solutions, but only got %d", tt.solutionCount, len(solutions))
			}
			for _, v := range solutions {
				t.Logf("game can be solved in %d moves with a score of %d and a HighestCellValue of %d", v.Moves(), v.Score(), v.HighestCellValue())

			}
		})
	}
}
func Benchmark_gameGeneratorTargetCell_GenerateGame(b *testing.B) {
	gen, err := NewGameGeneratorForTargetCell(
		GameGeneratorTargetCellOptions{
			TargetCell: 768,
			Rows:       5,
			Columns:    5,
			MaxMoves:   12,
		},
	)
	if err != nil {
		b.Fatalf("faile to run NewGameGeneratorForTargetCell: %v", err)
	}

	for i := 0; i < b.N; i++ {
		gen.generateGame()
	}

}
