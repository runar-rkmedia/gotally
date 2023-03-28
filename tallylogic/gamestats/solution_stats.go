package gamestats

import (
	"fmt"
	"strconv"

	"github.com/runar-rkmedia/gotally/tallylogic"
)

func NewSolutionsStats(game tallylogic.Game, solutions []tallylogic.Game) (SolutionStats, error) {
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
func calculateStat(original tallylogic.Game, solution tallylogic.Game) (SolutionStat, error) {
	if len(original.History) > 0 {
		return SolutionStat{}, fmt.Errorf("Not implemented. CalculateState currently only supports calculating Games where the original haz no History")
	}
	instructionLength := len(solution.History)
	s := SolutionStat{
		Moves:           solution.Moves() - original.Moves(),
		Score:           uint64(solution.Score()) - uint64(original.Score()),
		InstructionTags: make([]InstructionTag, instructionLength),
	}
	if len(solution.History) > 0 {
		gameCopy := original.Copy()

		for i, ins := range solution.History {

			t := tallylogic.GetInstructionType(ins)
			switch t {
			case tallylogic.InstructionTypeSwipe:
				s.InstructionTags[i] = InstructionTag{IsSwipe: true, Ok: true}
				if !gameCopy.Instruct(ins) {
					return s, fmt.Errorf("failed to instuct game to swipe")
				}
			case tallylogic.InstructionTypeCombinePath, tallylogic.InstructionTypeSelectCoord, tallylogic.InstructionTypeSelectIndex:
				path, ok := tallylogic.GetInstructionAsPath(ins)
				if !ok {
					return s, fmt.Errorf("failed to get instruction as path for instuction")
				}
				_, method, err := gameCopy.SoftEvaluatesForPath(path)
				if err != nil {
					return s, err
				}
				switch method {
				case tallylogic.EvalMethodSum:
					s.InstructionTags[i] = InstructionTag{IsAddition: true, Ok: true}
				case tallylogic.EvalMethodProduct:
					s.InstructionTags[i] = InstructionTag{IsMultiplication: true, Ok: true}
				}
				if !gameCopy.Instruct(ins) {
					return s, fmt.Errorf("failed to instuct game to combine")
				}
			default:
				return s, fmt.Errorf("Unhandled instructiontype %d with value %#v", t, ins)

			}

		}
	}
	fmt.Println("Calculating status for solution")
	fmt.Println(original.Print())
	fmt.Println(solution.DescribeInstruction(solution.History))
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
	TwoPow           uint64
}
