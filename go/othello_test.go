package othello

import "testing"

func TestInitState(t *testing.T) {
	b := Board{
		Pieces: [8][8]Piece{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 1, 2, 0, 0, 0},
			{0, 0, 0, 2, 1, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
		Next: Black,
	}
	moves := b.ValidMoves()
	if len(moves) != 4 {
		t.Fatalf("Get error: %v", moves)
	}
}

func TestNoWay(t *testing.T) {
	b := Board{
		Pieces: [8][8]Piece{
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0},
		},
		Next: Black,
	}
	moves := b.ValidMoves()
	if len(moves) != 0 {
		t.Fatalf("Get error: %v", moves)
	}
}
