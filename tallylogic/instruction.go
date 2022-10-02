package tallylogic

import (
	"encoding/gob"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type InstructionEnum uint64

const (
	InstructionEnumSwipeUp = iota + 1
	InstructionEnumSwipeRight
	InstructionEnumSwipeDown
	InstructionEnumSwipeLeft
)

var (
	// maploojup by map[column][row][InstructionEnumValue]combinePath
	instructionPathMap        map[int]map[int]map[InstructionEnum][]uint8
	instructionReversePathMap map[int]map[int]map[string]InstructionEnum
)

func expandInstruction(boardColumns int, boardRows int, instr InstructionEnum) any {
	switch instr {
	case 0:
		return nil
	case InstructionEnumSwipeUp:
		return SwipeDirectionUp
	case InstructionEnumSwipeRight:
		return SwipeDirectionRight
	case InstructionEnumSwipeDown:
		return SwipeDirectionDown
	case InstructionEnumSwipeLeft:
		return SwipeDirectionLeft
	}
	return instructionPathMap[boardColumns][boardRows][instr]
}

const (
	_sep = ";"
)

func populateInstuctionpath(boardColumns int, boardRows int) {
	if instructionPathMap == nil {
		instructionPathMap = map[int]map[int]map[InstructionEnum][]uint8{}
	}
	if instructionPathMap[boardColumns] == nil {
		instructionPathMap[boardColumns] = map[int]map[InstructionEnum][]uint8{}
	}
	if instructionReversePathMap == nil {
		instructionReversePathMap = map[int]map[int]map[string]InstructionEnum{}
	}
	if instructionReversePathMap[boardColumns] == nil {
		instructionReversePathMap[boardColumns] = map[int]map[string]InstructionEnum{}
	}
	length := boardColumns * boardRows
	board := NewTableBoard(boardColumns, boardRows)
	// set all cells to the magic value 0
	for i := 0; i < length; i++ {
		board.cells[i] = cell.NewCell(1, 0)
	}
	gh := NewHintCalculator(&board, &board, &board)

	fmt.Println(board.PrintBoard(nil))
	hints := gh.GetHints()
	if len(hints) == 0 {
		panic("recieved 0 paths")
	}
	fmt.Printf("\ngot %d hints", len(hints))
	tmp := make([][]uint8, len(hints))
	{
		var i InstructionEnum = 0
		for _, v := range hints {
			tmp[i] = pathSliceToUintPathSlice(v.Path)
			i++
		}
	}
	fmt.Printf("\nSorting paths...")

	// this is highly ineffecient
	sort.Slice(tmp, func(i, j int) bool {
		return hashUint8Path(tmp[i]) < hashUint8Path(tmp[j])
	})
	fmt.Printf("\nPopulating")
	instructionPathMap[boardColumns][boardRows] = make(map[InstructionEnum][]uint8, len(tmp))
	instructionReversePathMap[boardColumns][boardRows] = make(map[string]InstructionEnum, len(tmp))
	for j, v := range tmp {
		instructionPathMap[boardColumns][boardRows][InstructionEnum(j)] = v
		instructionReversePathMap[boardColumns][boardRows][hashUint8Path(v)] = InstructionEnum(j)
	}
	fmt.Printf("\nAll done with board %dx%d\n", boardColumns, boardRows)
}

func readInstrionsFromDisk(filePath string, j any) {
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	gobber := gob.NewDecoder(f)
	err = gobber.Decode(j)
	if err != nil {
		panic(err)
	}
	return
}

func init() {
	filePathInstr := "instructionPathMap.gob"
	filePathReverseInstr := "instructionPathMapReverse.gob"
	if _, err := os.Stat(filePathInstr); err == nil {
		readInstrionsFromDisk(filePathInstr, &instructionPathMap)
		// readInstrionsFromDisk(filePathReverseInstr, &instructionReversePathMap)
	} else {
		generateInstruionPathsToDisk()
		f, err := os.OpenFile(filePathInstr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		fmt.Println("Serializing")
		gobber := gob.NewEncoder(f)
		err = gobber.Encode(instructionPathMap)
		if err != nil {
			panic(err)
		}
		{

			f, err := os.OpenFile(filePathReverseInstr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			fmt.Println("Serializing")
			gobber := gob.NewEncoder(f)
			err = gobber.Encode(instructionReversePathMap)
			if err != nil {
				panic(err)
			}
		}
		fmt.Printf("\nAll done")
	}
	// fmt.Println(len(instructionReversePathMap[5][5]))
	fmt.Println(len(instructionPathMap[5][5]))
}

func pathSliceToUintPathSlice(path []int) []uint8 {
	m := make([]uint8, len(path))
	for i := 0; i < len(path); i++ {
		m[i] = uint8(path[i])
	}
	return m
}
func hashUint8Path(path []uint8) string {
	a := strings.Builder{}
	for i := 0; i < len(path); i++ {
		a.WriteString(strconv.FormatUint(uint64(path[i]), 10))
		a.WriteString(_sep)
	}
	return a.String()
}
func hashIntPath(path []uint) string {
	a := strings.Builder{}
	for i := 0; i < len(path); i++ {
		a.WriteString(strconv.FormatUint(uint64(path[i]), 10))
		a.WriteString(_sep)
	}
	return a.String()
}

func generateInstruionPathsToDisk() {
	populateInstuctionpath(3, 3)
	populateInstuctionpath(4, 4)
	populateInstuctionpath(5, 5)
}
