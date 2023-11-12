package repository

import (
	"bufio"
	"fmt"
	"io"
	"sort"
)

type line struct {
	number int
	text   string
}

func lines(r io.Reader) ([]*line, error) {

	scan := bufio.NewScanner(r)
	lines := []*line{}

	for i := 1; scan.Scan(); i++ {
		lines = append(lines, &line{i, scan.Text()})
	}

	if err := scan.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}

type myers struct {
	a, b []*line
}

func newMyers(a, b []*line) *myers {
	return &myers{a, b}
}

type symbol string

func (s symbol) String() string {
	return string(s)
}

const (
	Nochange  symbol = " "
	Insertion symbol = "+"
	Deletion  symbol = "-"
)

type edit struct {
	diff  symbol
	aline *line
	bline *line
	line  *line
}

func newEdit(diff symbol, aline, bline *line) *edit {

	var line *line
	if aline != nil {
		line = aline
	} else {
		line = bline
	}

	return &edit{diff, aline, bline, line}

}

func (e *edit) String() string {
	return fmt.Sprintf("%s%s", e.diff, e.line.text)
}

func (e *edit) Diff() symbol {
	return e.diff
}

type edits []*edit

func (es edits) filter(f func(e *edit) bool) (edits edits) {

	for _, e := range es {

		if f(e) {
			edits = append(edits, e)
		}

	}

	return
}

func (es edits) String() string {

	var s string
	for _, e := range es {
		s += fmt.Sprintln(e)
	}

	return s
}

const hunkContext int = 3

func (es edits) hunks() []*hunk {

	hunks := []*hunk{}
	offset := 0

	for {

		for offset < len(es) && es[offset].diff == Nochange {
			offset++
		}

		if offset >= len(es) {
			return hunks
		}

		offset -= hunkContext + 1

		aStart := 0
		if offset >= 0 {
			aStart = es[offset].aline.number
		}

		bStart := 0
		if offset >= 0 {
			bStart = es[offset].bline.number
		}

		hunk := &hunk{aStart, bStart, []*edit{}}
		offset = hunk.build(es, offset)
		hunks = append(hunks, hunk)

	}
}

type hunk struct {
	aStart, bStart int
	edits          edits
}

func (h *hunk) String() string {

	l := fmt.Sprintln(h.Header())

	for _, e := range h.edits {
		l += fmt.Sprintln(e)
	}

	return l
}

func (h *hunk) Edits() edits {
	return h.edits
}

func (h *hunk) Header() string {

	aStart, aSize := h.offset(func(e *edit) bool { return e.aline != nil }, func(e *edit) int { return e.aline.number }, h.aStart)
	bStart, bSize := h.offset(func(e *edit) bool { return e.bline != nil }, func(e *edit) int { return e.bline.number }, h.bStart)

	return fmt.Sprintf("@@ -%d,%d +%d,%d @@", aStart, aSize, bStart, bSize)
}

func (h *hunk) offset(filterFunc func(*edit) bool, startFunc func(*edit) int, defaultStart int) (int, int) {

	lines := h.edits.filter(filterFunc)

	start := defaultStart
	if len(lines) > 0 {
		start = startFunc(lines[0])
	}

	return start, len(lines)
}

func (h *hunk) build(edits edits, offset int) int {

	counter := -1
	for counter != 0 {

		if offset >= 0 && counter > 0 {
			h.edits = append(h.edits, edits[offset])
		}

		offset++
		if offset >= len(edits) {
			break
		}

		if offset+hunkContext < len(edits) {

			switch edits[offset+hunkContext].diff {
			case Insertion, Deletion:
				counter = 2*hunkContext + 1
			default:
				counter -= 1
			}

		}

	}

	return offset
}

func (my *myers) diff() edits {

	diff := edits{}

	moves := make(chan *move)
	go my.backtrack(moves)

	for {

		move, ok := <-moves
		if !ok {
			break
		}

		if move.x == move.prev.x {

			diff = append(diff, newEdit(Insertion, nil, my.b[move.prev.y]))

		} else if move.y == move.prev.y {

			diff = append(diff, newEdit(Deletion, my.a[move.prev.x], nil))

		} else {

			diff = append(diff, newEdit(Nochange, my.a[move.prev.x], my.b[move.prev.y]))

		}
	}

	// 後ろから前に向かってチェックしたので最後は順番を戻す
	sort.SliceStable(diff, func(i, j int) bool {
		return i > j
	})

	return diff

}

// intarray
type intarray []int

func (ns *intarray) set(i, v int) {

	(*ns)[ns.index(i)] = v
}

func (ns *intarray) get(i int) int {
	return (*ns)[ns.index(i)]
}

func (ns *intarray) index(i int) uint {
	if i >= 0 {
		return uint(i)
	} else if l := len(*ns) + i; l >= 0 {
		return uint(l)
	} else {
		panic(fmt.Sprintf("index out of range %d", i))
	}
}

// shortestEdit aをbにする最短の編集
//
// ex. a -> "ABCABBA" , b -> "CBABAC", k -> x-y
//
//   k| -3 -2 -1  0  1  2  3  4
// d --------------------------
// 0  |  0  0  0  0  0  0  0  0
// 1  |  0  0  0  0  1  0  0  0
// 2  |  0  2  0  2  1  3  0  0
// 3  |  3  2  4  2  5  3  5  0
// 4  |  3  4  4  5  5  7  5  7
// 4  |  3  4  5  5  7  7  5  7
func (my *myers) shortestEdit() []intarray {
	n, m := len(my.a), len(my.b)
	max := n + m

	if max == 0 {
		return []intarray{}
	}

	xs := make(intarray, 2*max+1)
	// (0, 0)の前は(0, -1)とする
	xs.set(1, 0)

	trace := []intarray{}

	// d -> 移動回数
	for d := 0; d <= max; d++ {

		// min(src, dist)の長さだけしかcopyされないので、
		// srcの長さだけdistの配列を作って全部コピる
		t := make(intarray, len(xs))
		copy(t, xs)
		trace = append(trace, t)

		// k -> x-y
		for k := -d; k <= d; k += 2 {

			//　k == -d （xが0）は全部下(y)を選択
			//  k != d && (xs.get(k-1) < xs.get(k+1)) は kが大きい方（削除優先）を選択
			x := 0
			if k == -d || (k != d && (xs.get(k-1) < xs.get(k+1))) {
				// 下
				x = xs.get(k + 1)
			} else {
				// 右
				x = xs.get(k-1) + 1
			}

			y := x - k

			// 同じ文字列は飛ばす
			for x < n && y < m && my.a[x].text == my.b[y].text {
				x, y = x+1, y+1
			}

			xs.set(k, x)

			// 右下に到達したら終わり
			if x >= n && y >= m {
				return trace
			}

		}
	}

	return trace
}

type vector2 struct {
	x, y int
}

type move struct {
	prev *vector2
	*vector2
}

// backtrack 終端の len(a),len(b) から 0,0 への最短編集経路のパスを生成する
//
// ex. a -> "ABCABBA" , b -> "CBABAC", k -> x-y
//
//   k| -3 -2 -1  0  1  2  3  4
// d --------------------------
// 0  |  0  0  0  G  0  0  0  0
// 1  |  0  0  0  0  X  0  0  0
// 2  |  0  2  0  2  1  X  0  0
// 3  |  3  2  4  2  X  3  5  0
// 4  |  3  4  4  5  5  X  5  7
// 4  |  3  4  5  5  S  7  5  7
func (my *myers) backtrack(moves chan<- *move) {
	x, y := len(my.a), len(my.b)
	trace := my.shortestEdit()

	// d -> 移動回数
	for d := len(trace) - 1; 0 <= d; d-- {

		xs := trace[d]

		// 削除優先でkが大きいところを進む
		k := x - y
		prevk := 0
		if k == -d || (k != d && (xs.get(k-1) < xs.get(k+1))) {
			prevk = k + 1 // prevk -> k は下
		} else {
			prevk = k - 1 // prevk -> k は右
		}

		prevx := xs.get(prevk)
		prevy := prevx - prevk
		// 飛ばした分を移動する
		for x > prevx && y > prevy {
			moves <- &move{&vector2{x - 1, y - 1}, &vector2{x, y}}
			x, y = x-1, y-1
		}

		if d > 0 {
			moves <- &move{&vector2{prevx, prevy}, &vector2{x, y}}
		}

		x, y = prevx, prevy
	}

	close(moves)

}
