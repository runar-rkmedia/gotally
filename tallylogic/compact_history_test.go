package tallylogic

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/runar-rkmedia/gotally/tallylogic/cell"
)

func TestCompactHistory(t *testing.T) {
	type fields struct {
	}
	type args struct {
		dir SwipeDirection
	}

	type testArgs struct {
		name          string
		columns, rows int
		// The test will append each of these
		history              string
		wantSize, wantLength int
	}
	testCreator := func(name string, seed int64, columns, rows, historyLength int, wantSize, wantLength int) testArgs {
		rnd := rand.NewSource(seed)
		res := testArgs{name, columns, rows, historyStringCreator(t, columns, rows, rnd, historyLength), wantSize, wantLength}
		hCount := strings.Count(res.history, ";")
		res.name = fmt.Sprintf("%s %dx%d-%d-%d", name, columns, rows, hCount, seed)
		return res
	}
	tests := []testArgs{
		{"Test simple history 5x5", 5, 5, "U;R;6,1,2,7,12,11;L;H;D;", 5, 13},
		testCreator("Test randomized history ", 1000, 5, 5, 8, 7, 17),
		testCreator("Test randomized history", 1001, 5, 5, 4, 3, 8),
		testCreator("Test randomized history", 1002, 5, 5, 30, 47, 123),
		testCreator("Test randomized history", 1003, 4, 4, 3, 3, 6),
		testCreator("Test randomized history", 1004, 4, 4, 400, 479, 1276),
		testCreator("Test randomized history on tiny board", 1005, 3, 2, 4, 8, 19),
		testCreator("Test randomized history on bigger board", 1006, 8, 8, 4, 12, 32),
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
			_, err := c.c.Unpack()
			if err != nil {
				t.Fatalf("failed to unpack: %v", err)
			}
			gotSize := c.c.Size()
			gotLength := c.c.Length()
			t.Logf("size %d bytes length %d. %.2f%%", gotSize, gotLength, float64(gotSize)/float64(gotLength)*100)
			described := c.Describe()
			if diff := deep.Equal(described, tt.history); diff != nil {
				t.Errorf("Described did not match %v\ngot:  %s\nwant: %s", diff, described, tt.history)
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
func historyStringCreator(t *testing.T, columns, rows int, rnd rand.Source, length int) string {
	s := ""
	boardSize := columns * rows
	tb := newTableBoard(t, columns, rows)
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
			//Path
			l := int(rnd.Int63())%8 + 2
			path := []int{int(rnd.Int63()) % boardSize}

			maxn := 12
			for i := 1; i < l; i++ {
				if maxn < 0 {
					continue outer
				}
				neighbours, ok := tb.NeighboursForCellIndex(path[i-1])
				if !ok {
					break outer
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
			for i := 0; i < len(path); i++ {
				s += strconv.FormatInt(int64(path[i]), 10) + ","
			}
			s = s[:len(s)-1]
			s += ";"
		}

	}
	return s
}
