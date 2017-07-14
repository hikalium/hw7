package othello

import (
	"io"
	"io/ioutil"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"

	"encoding/json"
	"fmt"
	"net/http"
)

var ScoreMap [8][8]int

func init() {
	http.HandleFunc("/", getMove)
	ScoreMap = [8][8]int{
		{15, 2, 5, 5, 5, 5, 2, 15},
		{2, 0, 1, 1, 1, 1, 0, 2},
		{5, 1, 1, 1, 1, 1, 1, 5},
		{5, 1, 1, 1, 1, 1, 1, 5},
		{5, 1, 1, 1, 1, 1, 1, 5},
		{5, 1, 1, 1, 1, 1, 1, 5},
		{2, 0, 1, 1, 1, 1, 0, 2},
		{15, 2, 5, 5, 5, 5, 2, 15},
	}
}

type Game struct {
	Board Board `json:board`
}

var ctx context.Context

func getMove(w http.ResponseWriter, r *http.Request) {
	ctx = appengine.NewContext(r)
	var js []byte
	defer r.Body.Close()
	js, _ = ioutil.ReadAll(r.Body)
	if len(js) < 1 {
		js = []byte(r.FormValue("json"))
	}
	if len(js) < 1 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<body><form method=get>
Paste JSON here:<p/><textarea name=json cols=80 rows=24></textarea>
<p/><input type=submit>
</form>
</body>`)
		return
	}
	var game Game
	err := json.Unmarshal(js, &game)
	if err != nil {
		fmt.Fprintf(w, "invalid json %v? %v", string(js), err)
		return
	}
	board := game.Board
	log.Infof(ctx, "Next: %v", board.Next)
	board.PrintLog(ctx)
	board.Eval()
	log.Infof(ctx, "%v", board.EvalScore)
	board.Moves = board.ValidMoves()
	if len(board.Moves) < 1 {
		fmt.Fprintf(w, "PASS")
		return
	}
	board.ExpandEvalTree("root:", 5)
	bestIndex := board.Moves.GetBestEvalIndex(board.Next)
	board.Moves.Send(bestIndex, w, ctx)
}

type Piece int8

const (
	Empty Piece = iota
	Black Piece = iota
	White Piece = iota

	// Red/Blue are aliases for Black/White
	Red  = Black
	Blue = White
)

func (p Piece) Opposite() Piece {
	switch p {
	case White:
		return Black
	case Black:
		return White
	default:
		return Empty
	}
}

type Board struct {
	// Layout says what pieces are where.
	Pieces [8][8]Piece
	// Next says what the color of the next piece played must be.
	Next      Piece
	EvalScore int
	Moves     MoveList
}

func (b Board) PrintLog(ctx context.Context) {
	for y := 0; y < 8; y++ {
		var s string

		for x := 0; x < 8; x++ {
			switch b.Pieces[y][x] {
			case White:
				s += "w "
			case Black:
				s += "b "
			default:
				s += "  "
			}
		}
		log.Infof(ctx, "%v\n", s)
	}
}

// At returns a pointer to the piece at a given position.
func (b *Board) At(p Position) *Piece {
	return &b.Pieces[p[1]-1][p[0]-1]
}

// Get returns the piece at a given position.
func (b *Board) Get(p Position) Piece {
	return *b.At(p)
}

// Exec runs a move on a given Board, updating the given board, and
// returning it. Returns error if the move is illegal.
func (b *Board) Exec(m Move) (*Board, error) {
	if !m.Where.Pass() {
		if _, err := b.realMove(m); err != nil {
			return b, err
		}
	} else {
		// Attempting to pass.
		valid := b.ValidMoves()
		if len(valid) > 0 {
			return nil, fmt.Errorf("%v illegal move: there are valid moves available: %v", m, valid)
		}
	}
	b.Next = b.Next.Opposite()
	return b, nil
}

// realMove executes a move that isn't a PASS.
func (b *Board) realMove(m Move) (*Board, error) {
	captures, err := b.tryMove(m)
	if err != nil {
		return nil, err
	}

	for _, p := range append(captures, m.Where) {
		*b.At(p) = m.As
	}
	return b, nil
}

func (b Board) GetMovedBoard(m Move) Board {
	b.Exec(m)
	return b
}

func (b *Board) GetEval() int {
	var eval int
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			switch b.Pieces[y][x] {
			case White:
				eval -= ScoreMap[y][x]
			case Black:
				eval += ScoreMap[y][x]
			}
		}
	}
	return eval
}

func (b *Board) Eval() {
	var eval int
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			switch b.Pieces[y][x] {
			case White:
				eval -= 1
			case Black:
				eval += 1
			}
		}
	}
	b.EvalScore = eval
}

func (b *Board) ExpandEvalTree(prefix string, depth int) {
	var evalScore int
	if depth == 0 {
		b.Eval()
		evalScore = b.EvalScore
	} else {
		b.Moves = b.ValidMoves()
		if len(b.Moves) == 0 {
			return
		}
		b.Moves.GenMovedBoards(b)
		for _, m := range b.Moves {
			board := m.MovedBoard
			board.ExpandEvalTree(m.ToStr(), depth-1)
		}
		bestIndex := b.Moves.GetBestEvalIndex(b.Next)
		if bestIndex < 0 {
			b.Eval()
			evalScore = b.EvalScore
		} else {
			evalScore = b.Moves[bestIndex].MovedBoard.EvalScore
			b.EvalScore = evalScore
		}
		//log.Infof(ctx, "bestIndex: %d", bestIndex)
	}
	/*
		var s string
		for i := 0; i < depth; i++ {
			s += "\t"
		}
		s += prefix
		s += strconv.Itoa(evalScore)
		log.Infof(ctx, s)
	*/
}

type BoardList []Board

func (bl *BoardList) LogAll(ctx context.Context) {
	log.Infof(ctx, "boards:")
	for _, v := range *bl {
		v.PrintLog(ctx)
		log.Infof(ctx, "----")
	}
}

// Position represents a position on the othello board. Valid board
// coordinates are 1-8 (not 0-7)!
type Position [2]int

// Valid returns true iff this is a valid board position.
func (p Position) Valid() bool {
	ok := func(i int) bool { return 1 <= i && i <= 8 }
	return ok(p[0]) && ok(p[1])
}

// Pass returns true iff this move position represents a pass.
func (p Position) Pass() bool {
	return !p.Valid()
}

// Move describes a move on an Othello board.
type Move struct {
	// Where a piece is going to be placed. If Where is zeros, or
	// another invalid coordinate, it indicates a pass.
	Where Position
	// As is the player taking the player taking the turn.
	As         Piece
	MovedBoard *Board
}

func (m Move) Send(w io.Writer, ctx context.Context) {
	fmt.Fprintf(w, "[%d,%d]", m.Where[0], m.Where[1])
	m.Log("Move to: ")
}

func (m Move) Log(prefix string) {
	log.Infof(ctx, "%s[%d,%d] (%d)", prefix, m.Where[0], m.Where[1], m.MovedBoard.EvalScore)
}

func (m *Move) ToStr() string {
	return fmt.Sprintf("[%d,%d]", m.Where[0], m.Where[1])
}

type MoveList []Move

func (ml MoveList) LogAll(ctx context.Context) {
	log.Infof(ctx, "moves:")
	for _, v := range ml {
		v.Log("")
	}
}

func (ml *MoveList) GenMovedBoards(baseBoard *Board) {
	for index, _ := range *ml {
		board := baseBoard.GetMovedBoard((*ml)[index])
		board.Eval()
		(*ml)[index].MovedBoard = &board
	}
}

func (ml *MoveList) GetBestEvalIndex(forPlayer Piece) int {
	var index int
	var m int
	if ml == nil {
		return -1
	}
	index = -1
	if len(*ml) > 0 {
		m = (*ml)[0].MovedBoard.EvalScore
		index = 0
	}
	if forPlayer == Black {
		for i := 1; i < len(*ml); i++ {
			if (*ml)[i].MovedBoard.EvalScore > m {
				m = (*ml)[i].MovedBoard.EvalScore
				index = i
			}
		}
	} else {
		for i := 1; i < len(*ml); i++ {
			if (*ml)[i].MovedBoard.EvalScore < m {
				m = (*ml)[i].MovedBoard.EvalScore
				index = i
			}
		}

	}
	return index
}

func (ml *MoveList) Send(index int, w io.Writer, ctx context.Context) {
	if index < 0 {
		fmt.Fprintf(w, "PASS")
		log.Infof(ctx, "Send: PASS")
		return
	}
	m := (*ml)[index]
	fmt.Fprintf(w, "[%d,%d]", m.Where[0], m.Where[1])
	m.Log("Send: ")
}

type direction Position

var dirs []direction

func init() {
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			if x == 0 && y == 0 {
				continue
			}
			dirs = append(dirs, direction{x, y})
		}
	}
}

// tryMove tries a non-PASS move without actually executing it.
// Returns the list of captures that would happen.
func (b *Board) tryMove(m Move) ([]Position, error) {
	if b.Get(m.Where) != Empty {
		return nil, fmt.Errorf("%v illegal move: %v is occupied by %v", m, m.Where, b.Get(m.Where))
	}

	var captures []Position
	for _, dir := range dirs {
		captures = append(captures, b.findCaptures(m, dir)...)
	}

	if len(captures) < 1 {
		return nil, fmt.Errorf("%v illegal move: no pieces were captured", m)
	}
	return captures, nil
}

func translate(p Position, d direction) Position {
	return Position{p[0] + d[0], p[1] + d[1]}
}

func (b *Board) findCaptures(m Move, dir direction) []Position {
	var caps []Position
	for p := m.Where; true; caps = append(caps, p) {
		p = translate(p, dir)
		if !p.Valid() {
			// End of board.
			return []Position{}
		}
		switch *b.At(p) {
		case m.As:
			return caps
		case Empty:
			return []Position{}
		}
	}
	panic("impossible")
}

func (b *Board) ValidMoves() MoveList {
	var moves MoveList
	for y := 1; y <= 8; y++ {
		for x := 1; x <= 8; x++ {
			m := Move{Where: Position{x, y}, As: b.Next}
			_, err := b.tryMove(m)
			if err == nil {
				moves = append(moves, m)
			}
		}
	}
	return moves
}
