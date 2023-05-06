package tallylogic

import (
	"fmt"
	"strconv"
)

func NewSolutionsStats(game Game, solutions []Game) (SolutionStats, error) {
	solutionStats := SolutionStats{Stats: make([]SolutionStat, len(solutions))}

	for i, s := range solutions {
		if s.Moves() < solutionStats.IdealMoves || solutionStats.IdealMoves == 0 {
			solutionStats.IdealMoves = s.Moves()
			solutionStats.ScoreOnIdeal = uint64(s.Score())
			solutionStats.IdealMovesSolutionIndex = i
		}
		if s.Score() > int64(solutionStats.MaxScore) {
			solutionStats.MaxScore = uint64(s.Score())
			solutionStats.MaxScoreSolutionIndex = i
		}
		stat, err := calculateStat(game, s)
		if err != nil {
			return solutionStats, err
		}
		solutionStats.Stats[i] = stat
	}
	return solutionStats, nil
}
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
func String(list []int) []string {
	s := make([]string, len(list))
	for i := 0; i < len(list); i++ {
		s[i] = strconv.FormatInt(int64(list[i]), 10)
	}
	return s
}
func calculateStat(original Game, solution Game) (SolutionStat, error) {
	instructionLength := original.History.Length()
	if instructionLength > 0 {
		return SolutionStat{}, fmt.Errorf("Not implemented. CalculateState currently only supports calculating Games where the original haz no History")
	}
	s := SolutionStat{
		Moves:           solution.Moves() - original.Moves(),
		Score:           uint64(solution.Score()) - uint64(original.Score()),
		InstructionTags: make([]InstructionTag, instructionLength),
	}
	if instructionLength > 0 {
		gameCopy := original.Copy()

		err := solution.History.Iterate(
			func(dir SwipeDirection, i int) error {
				s.InstructionTags[i] = InstructionTag{IsSwipe: true, Ok: true}
				if !gameCopy.Swipe(dir) {
					return fmt.Errorf("failed to instuct game to swipe")
				}
				return nil
			},
			func(path []int, i int) error {
				_, method, err := gameCopy.SoftEvaluatesForPath(path)
				if err != nil {
					return err
				}
				switch method {
				case EvalMethodSum:
					s.InstructionTags[i] = InstructionTag{IsAddition: true, Ok: true}
				case EvalMethodProduct:
					s.InstructionTags[i] = InstructionTag{IsMultiplication: true, Ok: true}
				}
				if !gameCopy.EvaluateForPath(path) {
					return fmt.Errorf("failed to instuct game to combine")
				}
				_, twoPow := gameCopy.Cells()[path[len(path)-1]].Raw()
				s.InstructionTags[i].TwoPow = twoPow
				return nil
			},
			func(helper Helper, i int) error {
				return fmt.Errorf("Not implemented: Helper in calculateStat")
			},
		)
		if err != nil {
			return s, err
		}
	}
	fmt.Println("Calculating status for solution")
	fmt.Println(original.Print())
	fmt.Println(solution.History.Describe())
	fmt.Printf("solutons has %d moves\n", s.Moves)
	fmt.Printf("solutons reached a score of %d \n", s.Score)

	return s, nil
}

type SolutionStats struct {
	IdealMovesSolutionIndex int
	MaxScoreSolutionIndex   int
	IdealMoves              int
	ScoreOnIdeal            uint64
	MaxScore                uint64
	Stats                   []SolutionStat
}
type SolutionStat struct {
	Moves           int
	Score           uint64
	InstructionTags []InstructionTag
}
type InstructionTag struct {
	Ok               bool
	IsMultiplication bool
	IsAddition       bool
	IsSwipe          bool
	TwoPow           int64
}
