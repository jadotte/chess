package chess

import (
	"fmt"
	"math/bits"
	"strings"
)

type (
	Bitboard uint64 // Using bitboard to improve model efficiency
	Piece    uint8
	Color    uint8
	Square   uint8
)

const (
	Empty Piece = iota
	Pawns
	Knights
	Bishops
	Rooks
	Queens
	Kings
)

const (
	White Color = iota
	Black
)


var (
	FileA Bitboard = 0x0101010101010101
	FileB Bitboard = FileA << 1
	FileC Bitboard = FileA << 2
	FileD Bitboard = FileA << 3
	FileE Bitboard = FileA << 4
	FileF Bitboard = FileA << 5
	FileG Bitboard = FileA << 6
	FileH Bitboard = FileA << 7

	Rank1 Bitboard = 0x00000000000000FF
	Rank2 Bitboard = Rank1 << 8
	Rank3 Bitboard = Rank1 << 16
	Rank4 Bitboard = Rank1 << 24
	Rank5 Bitboard = Rank1 << 32
	Rank6 Bitboard = Rank1 << 40
	Rank7 Bitboard = Rank1 << 48
	Rank8 Bitboard = Rank1 << 56
)

func (bb Bitboard) GetBit(sq Square) bool {
	return (bb & (1 << sq)) != 0
}

func (bb *Bitboard) ZeroBit(sq Square) {
	*bb &= ^(1 << sq)
}

func (bb *Bitboard) SetBit(sq Square) {
	*bb |= 1 << sq
}

func (c Color) Other() Color {
	if c == White {
		return Black
	} else {
		return White
	}
}

func (sq Square) GetRank() Bitboard {
	ranks := []Bitboard{Rank1, Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8}
	return ranks[sq/8]
}

func (sq Square) GetFile() Bitboard {
	files := []Bitboard{FileA, FileB, FileC, FileD, FileE, FileF, FileG, FileH}
	return files[sq%8]
}

func getSymbol(c Color, p Piece) rune {
	// returns coresponding symbols for piece/color combo inputted.
	// White is uppercase and black is lowercase

	var sym rune
	switch p {
	case Empty:
		sym = '.'
	case Pawns:
		sym = 'P'
	case Knights:
		sym = 'N'
	case Bishops:
		sym = 'B'
	case Rooks:
		sym = 'R'
	case Queens:
		sym = 'Q'
	case Kings:
		sym = 'K'
	}
	if c == Black && sym != '.' {
		sym += 32
	}
	return sym
}

func PrintBB(bb Bitboard) string {
	var sb strings.Builder
	sb.Grow(90)
	pieces := [64]rune{}
	for i := range pieces {
		pieces[i] = '.'
	}
	var sym rune = 'x'
	for bb != 0 {
		sq := Square(bits.TrailingZeros64(uint64(bb)))
		pieces[sq] = sym
		bb &= bb - 1
	}

	for rank := 7; rank >= 0; rank-- {
		sb.WriteString(fmt.Sprintf("%d ", rank+1))

		for file := range 8 {
			sq := Square(rank*8 + file)
			sb.WriteString(string(pieces[sq]) + " ")
		}

		sb.WriteString("\n")
	}

	sb.WriteString("  a b c d e f g h\n")

	return sb.String()
}

func onBoard(start int, shift int) bool {
	// Checks if a piece at the start cooardinate can move
	// by the shift value without moving off the board.
	// So the end cooardinate must be within 0 and 63 and
	// cannot try crossing one of the edges.
	// For example a Knight on a2 cannot make any moves left.

	if (start+shift >= 0) && (start+shift < 64) && (start%8+shift%8 < 8) && (start%8+shift%8 >= 0) {
		return true
	}
	return false
}

func GenAllKnightMoves() [64]Bitboard {
	// Generates an array of bitboards where each bitbaord i
	// contains all of the valid moves for a knight starting at
	// cooardinate i if they were on an empty board.

	var bbs [64]Bitboard
	var dists [8]int = [8]int{-17, -15, -10, -6, 6, 10, 15, 17}
	for sq := Square(0); sq < 64; sq++ {
		rank := sq.GetRank()
		file := sq.GetFile()
		for _, dist := range dists {
			if (rank == Rank1 && dist <= -6) || (rank == Rank2 && dist <= -15) {
				continue
			} else if rank == Rank8 && dist >= 6 || (rank == Rank7 && dist >= 15) {
				continue
			} else if (file == FileA || file == FileB) && (dist == 6 || dist == -10) {
				continue
			} else if (file == FileA) && (dist == 15 || dist == -17) {
				continue
			} else if (file == FileH || file == FileG) && (dist == -6 || dist == 10) {
				continue
			} else if (file == FileH) && (dist == 17 || dist == -15) {
				continue
			}

			end := int(sq) + dist
			bbs[sq] |= 1 << uint(end)
		}
	}
	return bbs
}

func GetPawnMoves(sq Square, fullBB Bitboard, color Color, otherColorBB Bitboard, enPassantSquare *Square) Bitboard {
	var moves Bitboard
	if color == White {
		// adds single push
		moves = (1 << (sq + 8)) & ^fullBB
		// adds double push
		moves |= ((moves & Rank3) << 8) & ^fullBB
		// adds takes
		moves |= (((1<<(sq + 7)) & ^FileH) & otherColorBB) | (((1<<(sq + 9)) & ^FileA) & otherColorBB)
    if enPassantSquare != nil {
      if (sq.GetRank() == Rank5) {
	  		epTarget := *enPassantSquare
  			if (sq+7 == epTarget && (1<<(sq+7)) & ^FileH != 0) ||
			    (sq+9 == epTarget && (1<<(sq+9)) & ^FileA != 0) {
				  moves |= (1 << epTarget)
			  }
      }
		}
	} else {
		// same for black
		moves = (1 << (sq - 8)) & ^fullBB
		moves |= ((moves & Rank6) >> 8) & ^fullBB
		moves |= (((1 << (sq - 7)) & ^FileH) & otherColorBB) | (((1 << (sq - 9)) & ^FileA) & otherColorBB)
    if enPassantSquare != nil {
      if (sq.GetRank() == Rank4){
			  epTarget := *enPassantSquare
  			if (sq-7 == epTarget && (1<<(sq-7)) & ^FileA != 0) ||
	  		   (sq-9 == epTarget && (1<<(sq-9)) & ^FileH != 0) {
		  		moves |= (1 << epTarget)
			  }
		  }
    }
	}
	return moves
}

func GetBishopMoves(sq Square, fullBB Bitboard, colorBB Bitboard) Bitboard {
	// Creates a bitboard of every legal move that a rook
	// of the coresponding color and square could make.
	rank := sq / 8
	file := sq % 8
	var moves Bitboard = 0

	// up right
	f := file + 1
	for r := rank + 1; r < 8; r++ {
		target := r*8 + f
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
		if f >= 7 {
			break
		}
		f++
	}
	// down right
	if rank != 0 {
		f = file + 1
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f >= 7 {
				break
			}
			f++
		}
	}
	// down left
	if rank != 0 && file != 0 {
		f = file - 1
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f == 0 {
				break
			}
			f--
		}
	}
	// up left
	if file != 0 {
		f = file - 1
		for r := rank + 1; r < 8; r++ {
			target := r*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
			if f == 0 {
				break
			}
			f--
		}
	}
	return moves &^ colorBB
}

func GetRookMoves(sq Square, fullBB Bitboard, colorBB Bitboard) Bitboard {
	// Creates a bitboard of every legal move that a rook
	// of the coresponding color and square could make.
	rank := sq / 8
	file := sq % 8
	var moves Bitboard = 0

	// up
	for r := rank + 1; r < 8; r++ {
		target := r*8 + file
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
	}
	// right
	for f := file + 1; f < 8; f++ {
		target := rank*8 + f
		moves |= 1 << uint(target)
		if (fullBB & (1 << uint(target))) != 0 {
			break
		}
	}
	// down
	if rank != 0 {
		for r := rank - 1; r >= 0; r-- {
			target := r*8 + file
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
		}
	}
	// left
	if file != 0 {
		for f := file - 1; f >= 0; f-- {
			target := rank*8 + f
			moves |= 1 << uint(target)
			if (fullBB & (1 << uint(target))) != 0 {
				break
			}
		}
	}
	return moves &^ colorBB
}

func GetQueenMoves(sq Square, fullBB Bitboard, colorBB Bitboard) Bitboard {
	return GetRookMoves(sq, fullBB, colorBB) | GetBishopMoves(sq, fullBB, colorBB)
}

func GetKingMoves(sq Square, fullBB Bitboard, color Color, RKR [3]bool, opBB, cBB Bitboard) Bitboard {
	king := Bitboard(1) << sq
	moves := (king << 8) |
		(king >> 8) |
		((king << 1) & ^FileA) |
		((king >> 1) & ^FileH) |
		((king << 9) & ^FileA) |
		((king << 7) & ^FileH) |
		((king >> 7) & ^FileA) |
		((king >> 9) & ^FileH)
	return (moves | GetCastles(color, fullBB, RKR, opBB)) &^ cBB
}

func GetCastles(color Color, fullBB Bitboard, RKR [3]bool, opBB Bitboard) Bitboard {
	var moves Bitboard
	var rank Bitboard
	if color == White {
		rank = Rank1
	} else {
		rank = Rank8
	}
	kingside := []Bitboard{FileF, FileG}
	queenside := []Bitboard{FileD, FileC, FileB}
	if !RKR[1] {
		if !RKR[0] {
			for i := range queenside {
				if (queenside[i]&rank&fullBB) != 0 || (queenside[i]&rank&opBB != 0) {
					break
				} else {
					moves |= FileC & rank
				}
			}
		}
		if !RKR[2] {
			for i := range kingside {
				if (kingside[i]&rank&fullBB) != 0 || (kingside[i]&rank&opBB != 0) {
					break
				} else {
					moves |= FileG & rank
				}
			}
		}
	}
	return moves
}

