package model

import (
	"bufio"
	"fmt"
	"io"
	"sort"
)

func lines(r io.Reader) ([]string, error) {

	scan := bufio.NewScanner(r)
	lines := []string{}
	for scan.Scan() {
		lines = append(lines, scan.Text())
	}

	if err := scan.Err(); err != nil {
		return lines, err
	}

	return lines, nil
}

type myers struct {
	a, b []string
}

func newMyers(a, b []string) *myers {
	return &myers{a, b}
}

type symbol string

func (s symbol) String() string {
	return string(s)
}

const (
	nochange  symbol = " "
	insertion symbol = "+"
	deletion  symbol = "-"
)

type edit struct {
	diff symbol
	str  string
}

func (e *edit) String() string {
	return fmt.Sprintf("%s%s", e.diff, e.str)
}

type edits []*edit

func (es edits) String() string {

	var s string
	for _, e := range es {
		s += fmt.Sprintln(e)
	}

	return s
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

			diff = append(diff, &edit{insertion, my.b[move.prev.y]})

		} else if move.y == move.prev.y {

			diff = append(diff, &edit{deletion, my.a[move.prev.x]})

		} else {

			diff = append(diff, &edit{nochange, my.a[move.prev.x]})

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
			for x < n && y < m && my.a[x] == my.b[y] {
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
