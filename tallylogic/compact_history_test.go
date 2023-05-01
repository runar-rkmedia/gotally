package tallylogic

import (
	"bytes"

	"compress/gzip"
	"compress/zlib"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

type compactTestArgs struct {
	name          string
	columns, rows int
	// The test will append each of these
	history              string
	wantSize, wantLength int
}

func Benchmark_Instruction_Speed_Original_No_Compression(b *testing.B) {
	game := mustCreateNewGameForTest(GameModeRandom, nil)()
	path := []int{1, 2, 3, 4, 5, 6}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		game.History.AddPath(path)
		game.History.AddSwipe(SwipeDirectionUp)
	}
	if len(game.History) != b.N*2 {
		b.Fatalf("Expected history to be of length %d, but it was %d", b.N, len(game.History))
	}
}
func Benchmark_Instruction_Speed_Original_Compression(b *testing.B) {
	game := mustCreateNewGameForTest(GameModeRandom, nil)()
	path := []int{1, 2, 3, 4, 5, 6}
	b.ResetTimer()
	kind := "zlib"
	for i := 0; i < b.N; i++ {
		buf := bytes.Buffer{}
		compress(kind, &buf, game.History)
		r := bytes.NewReader(buf.Bytes())
		decompress(kind, r, &game.History)
		game.History.AddPath(path)
		game.History.AddSwipe(SwipeDirectionUp)
	}
	if len(game.History) != b.N*2 {
		b.Fatalf("Expected history to be of length %d, but it was %d", b.N, len(game.History))
	}
}
func Benchmark_Instruction_Speed_Compact(b *testing.B) {
	// game := mustCreateNewGameForTest(GameModeRandom, nil)()
	path := []int{1, 2, 3, 4, 5, 6}
	b.ResetTimer()
	c := NewCompactHistory(5, 5)
	for i := 0; i < b.N; i++ {
		c.AddPath(path)
		c.AddSwipe(SwipeDirectionUp)
	}
}

func compress(kind string, buf io.Writer, obj any) error {
	var writer io.Writer
	switch kind {
	case "naked":
		writer = buf
	case "gzip":
		writer = gzip.NewWriter(buf)
	case "zlib":
		writer, _ = zlib.NewWriterLevel(buf, zlib.DefaultCompression)

	}
	if writer == nil {
		return fmt.Errorf("no reader during compress for kind %s", kind)
	}
	if c, ok := writer.(io.Closer); ok {
		defer c.Close()
	}
	enc := gob.NewEncoder(writer)
	return enc.Encode(obj)
}
func decompress(kind string, buf io.Reader, obj any) error {
	var reader io.Reader
	switch kind {
	case "naked":
		reader = buf
	case "gzip":
		reader, _ = gzip.NewReader(buf)
	case "zlib":
		reader, _ = zlib.NewReader(buf)
	}
	if reader == nil {
		return fmt.Errorf("no reader during decompress for kind %s", kind)
	}
	enc := gob.NewDecoder(reader)
	if c, ok := reader.(io.Closer); ok {
		defer c.Close()
	}
	return enc.Decode(obj)
}

func Test_InstructionCompression(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			"small",
			5,
		},
		{
			"medium",
			500,
		},
	}

	gob.Register(SwipeDirection(""))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rnd := rand.NewSource(123)
			game := mustCreateNewGameForTest(GameModeRandom, nil)()
			for i := 0; i < tt.length; i++ {
				path, err := createValidTestpath(t, game.Rules.SizeX, game.Rules.SizeY, rnd)
				if err != nil {
					t.Fatalf("failed to create valid test path: %v", err)
				}
				game.History.AddPath(path)
				game.History.AddSwipe(SwipeDirectionUp)
			}
			baseline := 0
			for _, kind := range []string{"naked", "gzip", "zlib"} {
				buf := &bytes.Buffer{}
				if err := compress(kind, buf, game.History); err != nil {
					t.Fatal(err.Error())
				}
				bb := buf.Bytes()
				if kind == "naked" {
					baseline = len(bb)
				}
				t.Logf("%s %d %d %.2f%%", kind, len(bb), len(game.History), float64(len(bb))/float64(baseline)*100)
				var history Instruction
				bufr := bytes.NewReader(bb)
				if err := decompress(kind, bufr, &history); err != nil {
					t.Fatal(err.Error(), bb)
				}
				if len(history) != len(game.History) {
					t.Errorf("length mismatch")
				}
				if diff := deep.Equal(history, game.History); diff != nil {
					t.Errorf("history is not equal: %v", diff)
				}
			}
			c := NewCompactHistory(game.Rules.SizeX, game.Rules.SizeY)
			for _, ins := range game.History {
				insType := GetInstructionType(ins)
				switch insType {
				case InstructionTypeCombinePath:
					p, ok := GetInstructionAsPath(ins)
					if !ok {
						t.Fatalf("instruction failed")
					}
					c.AddPath(p)
				case InstructionTypeSwipe:
					p, ok := GetInstructionAsSwipe(ins)
					if !ok {
						t.Fatalf("instruction failed")
					}
					c.AddSwipe(p)
				default:
					t.Fatalf("instruction not handled for type %v", insType)
				}
			}
			t.Logf("compact: %d Length %d %.2f%%", c.c.Size(), c.c.Length(),
				float64(c.c.Size())/float64(baseline)*100,
			)
			// turn back into game-history to verify
			g2 := mustCreateNewGameForTest(GameModeRandom, nil)()
			c.Iterate(
				func(dir SwipeDirection, i int) {
					g2.History.AddSwipe(dir)
				},
				func(path []int, i int) {
					g2.History.AddPath(path)
				},
				func(helper helper, i int) {
					fmt.Println("???", helper, i)
					u, err := c.c.Unpack()
					fmt.Println("???", u, err)
					fmt.Println("???", c.Describe())
					panic("no helper")
				},
			)
			t.Log("boardsize", game.Rules.SizeX, game.Rules.SizeY, g2.Rules.SizeX, g2.Rules.SizeY, c.gameColumns, c.gameRows)
			if diff := deep.Equal(g2.History, game.History); diff != nil {
				for _, v := range diff {
					t.Errorf("game2-history is not equal: %v", v)

				}
				for i := 0; i < int(math.Min(float64(len(g2.History)), float64(len(game.History)))); i++ {
					a := fmt.Sprintf("%v", game.History[i])
					b := fmt.Sprintf("%v", g2.History[i])
					if a == b {
						continue
					}
					t.Logf("game[%d]: %v !=  %v", i, a, b)

					if i > 10 {
						break
					}
				}
			}
		})

	}
}

func compactTestCreator(t *testing.T, name string, seed int64, columns, rows, historyLength int, wantSize, wantLength int) compactTestArgs {
	rnd := rand.NewSource(seed)
	res := compactTestArgs{name, columns, rows, historyStringCreator(t, columns, rows, rnd, historyLength), wantSize, wantLength}
	hCount := strings.Count(res.history, ";")
	res.name = fmt.Sprintf("%s %dx%d-%d-%d", name, columns, rows, hCount, seed)
	return res
}
func TestCompactHistory(t *testing.T) {
	type fields struct {
	}
	type args struct {
		dir SwipeDirection
	}

	tests := []compactTestArgs{
		{"Test simple history 5x5", 5, 5, "U;R;6,1,2,7,12,11;L;H;D;", 5, 13},
		compactTestCreator(t, "Test randomized history ", 1000, 5, 5, 8, 7, 17),
		compactTestCreator(t, "Test randomized history", 1001, 5, 5, 4, 3, 8),
		compactTestCreator(t, "Test randomized history", 1002, 5, 5, 30, 47, 123),
		compactTestCreator(t, "Test randomized history", 1003, 4, 4, 3, 3, 6),
		compactTestCreator(t, "Test randomized history", 1004, 4, 4, 400, 479, 1276),
		compactTestCreator(t, "Test randomized history on tiny board", 1005, 3, 2, 4, 8, 19),
		compactTestCreator(t, "Test randomized history on bigger board", 1006, 8, 8, 4, 12, 32),
	}
	pathRegex := regexp.MustCompile(`^[0-9,]{2,}$`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCompactHistory(tt.columns, tt.rows)
			t.Log("input-history", tt.history)
			tb := newTableBoard(t, tt.columns, tt.rows)
			parsePathStr := func(x string) ([]int, error, int) {
				pathStr := strings.Split(x, ",")
				path := make([]int, len(pathStr))
				for i := 0; i < len(pathStr); i++ {
					j, err := strconv.ParseInt(pathStr[i], 10, 64)
					if err != nil {
						t.Fatalf("Failed to parse from history as path-index: %v %s", err, pathStr[i])
					}
					path[i] = int(j)
				}
				err, invalidIndex := tb.ValidatePath(path)
				return path, err, invalidIndex
			}
			for _, v := range strings.Split(tt.history, ";") {
				x := strings.TrimSuffix(v, ";")
				switch x {
				case "":
					continue
				case "U":
					c.AddSwipe(SwipeDirectionUp)
				case "R":
					c.AddSwipe(SwipeDirectionRight)
				case "D":
					c.AddSwipe(SwipeDirectionDown)
				case "L":
					c.AddSwipe(SwipeDirectionLeft)
				case "H":
					c.AddHint()
				case "S":
					c.AddSwap()
				case "Z":
					c.AddUndo()
				default:
					if pathRegex.MatchString(x) {
						// interpret as a comma-seperated path
						path, err, invalidIndex := parsePathStr(x)
						if err != nil {
							t.Log(tb.PrintBoard(SelectionColoredHightlighter(path)))
							t.Fatalf("SanityTest: The test-input had an invalid path at index %d: %v", invalidIndex, err)
						}
						c.AddPath(path)
						continue
					}
					t.Fatalf("Invalid History-string from test-arguments: %s", x)
				}
			}
			unpacked, err := c.c.Unpack()
			if err != nil {
				t.Fatalf("failed to unpack: %v", err)
			}
			gotSize := c.c.Size()
			gotLength := c.c.Length()
			t.Logf("size %d bytes length %d. %.2f%%", gotSize, gotLength, float64(gotSize)/float64(gotLength)*100)
			described := c.Describe()
			if diff := deep.Equal(described, tt.history); diff != nil {
				t.Errorf("Describe did not match %v\ngot:  %s\nwant: %s\nUnpacked: %v", diff, described, tt.history, unpacked)
				d := strings.Split(described, ";")
				h := strings.Split(tt.history, ";")
				for i := 0; i < int(math.Min(float64(len(d)), float64(len(h)))); i++ {
					if d[i] == h[i] {
						continue
					}
					if pathRegex.MatchString(h[i]) {
						path, err, invalidIndex := parsePathStr(h[i])
						if err != nil {
							t.Log(tb.PrintBoard(SelectionColoredHightlighter(path)))
							t.Fatalf("SanityTest: The test-input had an invalid path at index %d: %v", invalidIndex, err)
						}
						t.Log("Expected path", tb.PrintBoard(SelectionColoredHightlighter(path)))
					} else {
						t.Logf("no match %v", d[i])
					}
					if pathRegex.MatchString(d[i]) {
						path, err, invalidIndex := parsePathStr(d[i])
						if err != nil {
							t.Log("Got path", tb.PrintBoard(SelectionColoredHightlighter(path)))
							t.Errorf("The path failed validation at %d: %v", invalidIndex, err)
						}
					} else {
						t.Logf("no match %v", d[i])
					}
					t.Errorf("Mismatch at index %d, %s != %s ", i, string(d[i]), string(h[i]))
					break

				}
			}
			if tt.wantSize != c.c.Size() {
				t.Errorf("Expected Size to be %d, but it was %d", tt.wantSize, c.c.Size())
			}
			if tt.wantLength != c.c.Length() {
				t.Errorf("Expected Length to be %d, but it was %d", tt.wantLength, c.c.Length())
			}
		})
	}
}
func newTableBoard(t *testing.T, columns, rows int) TableBoard {
	tb := NewTableBoard(columns, rows)
	boardSize := columns * rows
	for i := 0; i < boardSize; i++ {
		v := int64(i)
		if v == 0 {
			v = 99
		}
		err := tb.AddCellToBoard(cell.NewCell(v, 0), i, true)
		if err != nil {
			t.Fatalf("faile to add cell to board: %v", err)
		}
	}
	return tb
}
func createValidTestpath(t *testing.T, columns, rows int, rnd rand.Source) ([]int, error) {
	boardSize := columns * rows
	tb := newTableBoard(t, columns, rows)

	//Path
	l := int(rnd.Int63())%8 + 2
	path := []int{int(rnd.Int63()) % boardSize}

	maxn := 12
	for i := 1; i < l; i++ {
		if maxn < 0 {
			return path, fmt.Errorf("max retries triggered")
		}
		neighbours, ok := tb.NeighboursForCellIndex(path[i-1])
		if !ok {
			return path, fmt.Errorf("failed to get neighbours")
		}
		attempts := map[int]struct{}{}
		var n int
	o:
		for len(attempts) < len(neighbours) {

			n = neighbours[int(rnd.Int63())%len(neighbours)]
			if _, exists := attempts[n]; exists {
				continue o
			}

			for _, v := range path {
				if v == n {
					attempts[v] = struct{}{}
					continue o
				}
			}
			break o
		}
		if len(attempts) >= len(neighbours) {
			maxn = 12
			break
		}
		path = append(path, n)

	}
	if err, invalidIndex := tb.ValidatePath(path); err != nil {
		t.Log("Path", path)
		t.Fatalf("While generating paths for test, the resulting path is reported as invalid from TableBoard: at index %d %v", invalidIndex, err)
	}
	return path, nil
}
func historyStringCreator(t *testing.T, columns, rows int, rnd rand.Source, length int) string {
	s := ""
outer:
	for i := 0; i < length; i++ {

		kind := int(rnd.Int63()) % 3

		switch kind {
		case 0:
			//swipe
			dir := int(rnd.Int63()) % 4
			switch dir {
			case 0:
				s += "U;"
			case 1:
				s += "R;"
			case 2:
				s += "D;"
			case 3:
				s += "L;"
			}
		case 1:
			//Helper
			h := int(rnd.Int63()) % 3
			switch h {
			case 0:
				s += "H;"
			case 1:
				s += "S;"
			case 2:
				s += "Z;"
			}
		case 2:
			path, err := createValidTestpath(t, columns, rows, rnd)
			if err != nil {
				continue outer
			}
			for i := 0; i < len(path); i++ {
				s += strconv.FormatInt(int64(path[i]), 10) + ","
			}
			s = s[:len(s)-1]
			s += ";"
		}

	}
	return s
}
