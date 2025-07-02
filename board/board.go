package chess

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"strings"
)

type Board struct {
	PieceBB [2][7]Bitboard

	ColorBB [2]Bitboard

	FullBB Bitboard

	Turn Color

	KnightMoves [64]Bitboard

	RKRmoved [2][3]bool

	EnPassantSquare *Square

	MoveCounter uint8

  History map[uint64] int

  allKnightMoves [64]Bitboard
  
  NotationToIndex map[string]Square

  TotalMoves uint8
}

type Move struct {
  Start Square

  End Square

  Promotion Piece
}

func (b *Board) GetZobristHash() uint64 {
  var hash uint64
  for c := White; c <= Black; c ++ {
    for p := Pawns; p< Kings; p++ {
      bb := b.PieceBB[c][p]
      for bb != 0 {
        s := Square(bits.TrailingZeros64(uint64(bb)))
        hash ^= zobristKeys.Pieces[c][p][s]
        bb &= bb -1
      }
    }
  }
  if b.Turn == White {
    hash ^= zobristKeys.Turn
  }
  // White king side castle => 1___
  // WQSK => _1__
  // etc
  castling := uint8(0)
  if !b.RKRmoved[White][1] && !b.RKRmoved[White][2] {
  castling |= (1 << 3)
  }
  if !b.RKRmoved[White][1] && !b.RKRmoved[White][0] {
    castling |= (1 << 2)
  }
  if !b.RKRmoved[Black][1] && !b.RKRmoved[Black][2] {
    castling |= (1 << 1)
  }
  if !b.RKRmoved[Black][1] && !b.RKRmoved[Black][0] {
    castling |= (1 << 0)
  }
  hash ^= zobristKeys.Castling[castling]
  if b.EnPassantSquare != nil {
    hash ^= zobristKeys.EnPassant[uint8(*b.EnPassantSquare)%8] 
  }
  return hash

}

func NotationToIndex (str string) Square {
  var nti = map[string]Square{
		"a1": 0, "b1": 1, "c1": 2, "d1": 3, "e1": 4, "f1": 5, "g1": 6, "h1": 7,
		"a2": 8, "b2": 9, "c2": 10, "d2": 11, "e2": 12, "f2": 13, "g2": 14, "h2": 15,
		"a3": 16, "b3": 17, "c3": 18, "d3": 19, "e3": 20, "f3": 21, "g3": 22, "h3": 23,
		"a4": 24, "b4": 25, "c4": 26, "d4": 27, "e4": 28, "f4": 29, "g4": 30, "h4": 31,
		"a5": 32, "b5": 33, "c5": 34, "d5": 35, "e5": 36, "f5": 37, "g5": 38, "h5": 39,
		"a6": 40, "b6": 41, "c6": 42, "d6": 43, "e6": 44, "f6": 45, "g6": 46, "h6": 47,
		"a7": 48, "b7": 49, "c7": 50, "d7": 51, "e7": 52, "f7": 53, "g7": 54, "h7": 55,
		"a8": 56, "b8": 57, "c8": 58, "d8": 59, "e8": 60, "f8": 61, "g8": 62, "h8": 63,
	}
  return nti[str]
}

func NewBoard() *Board {
	// initializes new board with starting chess possition
	// filled in all sub-bitboards 
  
	b := &Board{
		Turn: White,
	}
	b.PieceBB[White][Pawns] = Rank2
	b.PieceBB[White][Knights] = (1 << NotationToIndex("b1")) | (1 << NotationToIndex("g1"))
	b.PieceBB[White][Bishops] = (1 << NotationToIndex("c1")) | (1 << NotationToIndex("f1"))
	b.PieceBB[White][Rooks] = (1 << NotationToIndex("a1")) | (1 << NotationToIndex("h1"))
	b.PieceBB[White][Queens] = (1 << NotationToIndex("d1"))
	b.PieceBB[White][Kings] = (1 << NotationToIndex("e1"))

	b.PieceBB[Black][Pawns] = Rank7
	b.PieceBB[Black][Knights] = (1 << NotationToIndex("b8")) | (1 << NotationToIndex("g8"))
	b.PieceBB[Black][Bishops] = (1 << NotationToIndex("c8")) | (1 << NotationToIndex("f8"))
	b.PieceBB[Black][Rooks] = (1 << NotationToIndex("a8")) | (1 << NotationToIndex("h8"))
	b.PieceBB[Black][Queens] = (1 << NotationToIndex("d8"))
	b.PieceBB[Black][Kings] = (1 << NotationToIndex("e8"))
	b.CombineBB()
	b.EnPassantSquare = nil
  b.History = make(map[uint64]int)
  b.allKnightMoves = GenAllKnightMoves()

	return b
}

func (b *Board) IsLegal(move Move) bool {
	color := b.Turn
  allLegalMoves := b.GetAllLegalMoves(color)
  for _, m := range allLegalMoves {
    if m == move {
    return true
    }
  }
  return false
}
func (b *Board) IsAttacked(color Color, sq Square) bool {
  attacksByColor := b.AllAttacks(color)
  return attacksByColor.GetBit(sq)
}

func (b *Board) IsCheck(color Color) bool {
	otherColor := color.Other()
  if (b.PieceBB[color][Kings] & b.AllAttacks(otherColor)) != 0 {
		return true
	}
  return false
}

func (b *Board) IsCheckmate() bool {
  color := b.Turn
  if !b.IsCheck(color) {
    return false
  }
  allMoves := b.GetAllLegalMoves(color)
  return (len(allMoves) == 0)
}

func (b *Board) IsStalemate() bool {
  color := b.Turn
  if b.IsCheck(color) {
    return false
  }
  allMoves := b.GetAllLegalMoves(color)
  return (len(allMoves) == 0)
}

func (b *Board) IsThreefold() bool {
  return (b.History[b.GetZobristHash()] >= 3)
}
func (b *Board) Is50Moves() bool {
  return (b.MoveCounter >= 100)
}

func (b *Board) GetAllLegalMoves(color Color) []Move {
  var  legalMoves[]Move

  for p := Pawns; p <= Kings; p++ {
    pieceBB := b.PieceBB[color][p]
    for pieceBB != 0 {
      start := Square(bits.TrailingZeros64(uint64(pieceBB)))
      pieceBB.ZeroBit(start)
      var tempMoves Bitboard
      switch p {
      case Pawns:
        tempMoves = GetPawnMoves(start, b.FullBB, color, b.ColorBB[color.Other()], b.EnPassantSquare)
      case Knights:
        tempMoves = b.allKnightMoves[start] &^ b.ColorBB[color]
      case Bishops:
        tempMoves = GetBishopMoves(start, b.FullBB, b.ColorBB[color])
      case Rooks:
        tempMoves = GetRookMoves(start, b.FullBB, b.ColorBB[color])
      case Queens:
        tempMoves = GetQueenMoves(start, b.FullBB, b.ColorBB[color])
      case Kings:
        tempMoves = GetKingMoves(start, b.FullBB, color, b.RKRmoved[color], b.AllAttacks(color.Other()), b.ColorBB[color])
      }
      // remove illegal moves
      for tempMoves != 0 {
        end := Square(bits.TrailingZeros64(uint64(tempMoves)))
        tempMoves.ZeroBit(end)

        if p == Pawns && (end.GetRank() == Rank8 || end.GetRank() == Rank1) {
          for _, promotion := range []Piece{Queens, Rooks, Bishops, Knights} {
            test := Move{start, end, promotion}
            if b.IsSimMoveLegal(test, color) {
              legalMoves = append(legalMoves, test)
            }
          }
        } else {
          test := Move{start, end, Empty}
          if b.IsSimMoveLegal(test, color) {
              legalMoves = append(legalMoves, test)
            }
        }
      }
    }
  }
    return legalMoves
}

func (b *Board) IsSimMoveLegal(move Move, color Color) bool {
  simBoard := b.Clone()
  _ = simBoard.MovePiece(simBoard.GetPieceAt(move.Start, color), move)

  if simBoard.IsCheck(color) {
    return false
  }
  return true
}

func (b *Board) Clone() *Board {
  cloned := &Board{
    FullBB: b.FullBB,
    Turn: b.Turn,
    MoveCounter: b.MoveCounter,
    EnPassantSquare: b.EnPassantSquare,
    TotalMoves: b.TotalMoves,
    History: make(map[uint64]int, len(b.History)),
  }
  copy(cloned.PieceBB[0][:], b.PieceBB[0][:])
  copy(cloned.PieceBB[1][:], b.PieceBB[1][:])
  cloned.ColorBB[0] = b.ColorBB[0]
  cloned.ColorBB[1] = b.ColorBB[1]
  copy(cloned.RKRmoved[0][:], b.RKRmoved[0][:])
  copy(cloned.RKRmoved[1][:], b.RKRmoved[1][:])
  for k, v := range b.History {
        cloned.History[k] = v
    }
  return cloned
}

func (b *Board) MovePiece(piece Piece, move Move) bool {
	// moves piece at start to end and promotes if specified
	// returns true if there is a piece captured
  start := move.Start
  end := move.End
  promotion := move.Promotion
	color := b.Turn
	otherColor := color.Other()
	startFile := start.GetFile()
	startRank := start.GetRank()
	endFile := end.GetFile()
	endRank := end.GetRank()

	var capture bool = false

	if piece == Empty {
		return false
	}
	// checks through the other colored bb to remove captured piece if relevent
	for p := Pawns; p <= Kings; p++ {
		if b.PieceBB[otherColor][p]&(1<<end) != 0 {
			b.PieceBB[otherColor][p].ZeroBit(end)
			capture = true
		}
	}
	// checks if castling
	b.EnPassantSquare = nil
	if piece == Kings {
		if startFile == FileE && endFile == FileC {
			if color == White {
				b.MovePiece(Rooks, Move{1, 4, Empty})
			} else {
				b.MovePiece(Rooks, Move{56, 59, Empty})
			}
			b.RKRmoved[color][2] = true
		} else if startFile == FileE && endFile == FileG {
			if color == White {
				b.MovePiece(Rooks, Move{7, 5, Empty})
			} else {
				b.MovePiece(Rooks, Move{63, 61, Empty})
			}
			b.RKRmoved[color][0] = true
		}
		b.RKRmoved[color][1] = true
	} else if piece == Rooks {
		if startFile == FileA {
			b.RKRmoved[color][0] = true
		} else if startFile == FileH {
			b.RKRmoved[color][2] = true
		}
    // or en passant square created
	} else if piece == Pawns {
		if color == White && startRank == Rank2 && endRank == Rank4 {
			newSqVal := end - 8
      newSq := new(Square)
      *newSq = newSqVal
			b.EnPassantSquare = newSq
		} else if color == Black && startRank == Rank7 && endRank == Rank5 {
			newSqVal := end + 8
      newSq := new(Square)
      *newSq = newSqVal
			b.EnPassantSquare = newSq
		}
  }
	// actually moves the selected piece with promotion check
	if promotion == Empty {
		b.PieceBB[color][piece].SetBit(end)
	} else {
		b.PieceBB[color][promotion].SetBit(end)
	}
	b.PieceBB[color][piece].ZeroBit(start)
	// pushes through the changed piecebb to affect all other bbs
	b.CombineBB()
	return capture
}

func (b *Board) CombineBB() {
	b.ColorBB[White] = 0
	b.ColorBB[Black] = 0

	for p := Pawns; p <= Kings; p++ {
		b.ColorBB[White] |= b.PieceBB[White][p]
		b.ColorBB[Black] |= b.PieceBB[Black][p]
	}

	b.FullBB = b.ColorBB[White] | b.ColorBB[Black]
}

func (b *Board) GetPieceAt(sq Square, color Color) Piece {
	for p := Pawns; p <= Kings; p++ {
		if b.PieceBB[color][p].GetBit(sq) {
			return p
		}
	}
	return Empty
}

func (b *Board) AllAttacks(color Color) Bitboard {
	var moves Bitboard
	moves |= b.GetPawnAttacks(color) | b.GetKnightMoves(color)
	moves |= b.GetBishopMoves(color) | b.GetRookMoves(color)
	moves |= b.GetQueenMoves(color) | b.GetKingMoves(color, true)
	return moves
}

func (b *Board) PrintBoard() string {
	var sb strings.Builder
	sb.Grow(90)
	pieces := [64]rune{}
	for i := range pieces {
		pieces[i] = '.'
	}
	for c := White; c <= Black; c++ {
		for p := Empty; p <= Kings; p++ {
			bb := b.PieceBB[c][p]
			var sym rune = getSymbol(c, p)
			for bb != 0 {
				sq := Square(bits.TrailingZeros64(uint64(bb)))
				pieces[sq] = sym
				bb &= bb - 1
			}
		}
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

func (b *Board) GetPawnMoves(color Color) Bitboard {
	// creates a bitboard of every legal move that a pawn of the
	// coresponding color could make (excluding attacks).
	var moves Bitboard
	pawns := b.PieceBB[color][Pawns]

	if color == White {
		// adds single push
		moves = (pawns << 8) & ^b.FullBB
		// adds double push
		moves |= ((moves & Rank3) << 8) & ^b.FullBB
	} else {
		// same for black
		moves = (pawns >> 8) & ^b.FullBB
		moves |= ((moves & Rank6) >> 8) & ^b.FullBB
	}
	return moves
}

func (b *Board) GetPawnAttacks(color Color) Bitboard {
	// creates a bitboard of every legal attack that a pawn of the
	// coresponding color could make.
	var moves Bitboard
	pawns := b.PieceBB[color][Pawns]
	if color == White {
		moves = (((pawns << 7) & ^FileA) & b.ColorBB[Black]) | (((pawns << 9) & ^FileH) & b.ColorBB[Black])
    if b.EnPassantSquare != nil {
			epTarget := *b.EnPassantSquare
			if epTarget.GetRank() == Rank6 {
				if (pawns & (epTarget.GetFile() << 7 & ^FileH & Rank5)) != 0 {
					moves |= (1 << epTarget)
				}
				if (pawns & (epTarget.GetFile() >> 7 & ^FileA & Rank5)) != 0 {
					moves |= (1 << epTarget)
				}
			}
		}
	} else {
		moves = (((pawns >> 7) & ^FileA) & b.ColorBB[White]) | (((pawns >> 9) & ^FileH) & b.ColorBB[White])
    if b.EnPassantSquare != nil {
			epTarget := *b.EnPassantSquare
			if epTarget.GetRank() == Rank3 {
				if (pawns & (epTarget.GetFile() << 7 & ^FileH & Rank4)) != 0 {
					moves |= (1 << epTarget)
				}
				if (pawns & (epTarget.GetFile() >> 7 & ^FileA & Rank4)) != 0 {
					moves |= (1 << epTarget)
				}
			}
		}
	}
	return moves
}

func (b *Board) GetKnightMoves(color Color) Bitboard {
	// Creates a bitboard of every legal move that a knight
	// of the coresponding color and square could make.
	knights := b.PieceBB[color][Knights]
	var moves Bitboard
	for knights != 0 {
		loc := Square(bits.TrailingZeros64(uint64(knights)))
		knights &= knights - 1
		moves |= b.allKnightMoves[loc] & ^b.ColorBB[color]
	}
	return moves
}

func (b *Board) GetBishopMoves(color Color) Bitboard {
	// creates a bitboard of every legal attack that a bishop of the
	// coresponding color could make.
	bishops := b.PieceBB[color][Bishops]
	var moves Bitboard
	for bishops != 0 {
		loc := Square(bits.TrailingZeros64(uint64(bishops)))
		bishops &= bishops - 1
		moves |= GetBishopMoves(loc, b.FullBB, b.ColorBB[color])
	}
	return moves
}

func (b *Board) GetRookMoves(color Color) Bitboard {
	rooks := b.PieceBB[color][Rooks]
	var moves Bitboard
	for rooks != 0 {
		loc := Square(bits.TrailingZeros64(uint64(rooks)))
		rooks &= rooks - 1
		moves |= GetRookMoves(loc, b.FullBB, b.ColorBB[color])
	}
	return moves
}

func (b *Board) GetQueenMoves(color Color) Bitboard {
	queens := b.PieceBB[color][Queens]
	var moves Bitboard
	for queens != 0 {
		loc := Square(bits.TrailingZeros64(uint64(queens)))
		queens &= queens - 1
		moves |= GetQueenMoves(loc, b.FullBB, b.ColorBB[color])
	}
	return moves
}

func (b *Board) GetKingMoves(color Color, attacks bool) Bitboard {
  var opBB Bitboard
  if attacks {
    opBB = 0
  } else {
    opBB = b.AllAttacks(b.Turn.Other())
  }
	king := b.PieceBB[color][Kings]
	var moves Bitboard
	loc := Square(bits.TrailingZeros64(uint64(king)))
	// treats opBB as empty here
	moves |= GetKingMoves(loc, b.FullBB, color, b.RKRmoved[color], opBB, b.ColorBB[color])
	return moves
}

func (board *Board) ToTensor() [8][8][19]float32 {
  // converts board to tensor to be used as input for CNN.
  // not yet fully implemented

	var tensor [8][8][19]float32

	for row := 0; row < 8; row++ {
		for col := 0; col < 8; col++ {
			square := Square(row*8 + col)
			color := board.Turn
			piece := board.GetPieceAt(square, color)

			if color == White {
				switch piece {
				case Pawns:
					tensor[row][col][0] = 1.0
				case Knights:
					tensor[row][col][1] = 1.0
				case Bishops:
					tensor[row][col][2] = 1.0
				case Rooks:
					tensor[row][col][3] = 1.0
				case Queens:
					tensor[row][col][4] = 1.0
				case Kings:
					tensor[row][col][5] = 1.0
				}
			} else if color == Black {
				switch piece {
				case Pawns:
					tensor[row][col][6] = 1.0
				case Knights:
					tensor[row][col][7] = 1.0
				case Bishops:
					tensor[row][col][8] = 1.0
				case Rooks:
					tensor[row][col][9] = 1.0
				case Queens:
					tensor[row][col][10] = 1.0
				case Kings:
					tensor[row][col][11] = 1.0
				}
			}

			// Channel 12: Side to move (1 if white, 0 if black)
			switch color {
			case White:
				tensor[row][col][12] = 1.0
			case Black:
				tensor[row][col][12] = 0.0
			}
			// Channel 13-16: Castling rights
			opBB := board.AllAttacks(color.Other())
			whiteCastles := GetCastles(color, board.FullBB, board.RKRmoved[White], opBB)
			blackCastles := GetCastles(color, board.FullBB, board.RKRmoved[Black], opBB)
			if whiteCastles&FileG != 0 {
				tensor[row][col][13] = 1.0
			} else {
				tensor[row][col][13] = 1.0
			}
			if whiteCastles&FileC != 0 {
				tensor[row][col][14] = 1.0
			} else {
				tensor[row][col][14] = 1.0
			}
			if blackCastles&FileG != 0 {
				tensor[row][col][15] = 1.0
			} else {
				tensor[row][col][15] = 1.0
			}
			if blackCastles&FileC != 0 {
				tensor[row][col][16] = 1.0
			} else {
				tensor[row][col][16] = 1.0
			}

			// Channel 17: En passant square
			if board.EnPassantSquare != nil &&
				board.EnPassantSquare.GetRank() == Square(row).GetRank() &&
				board.EnPassantSquare.GetFile() == Square(col).GetFile() {
				tensor[row][col][17] = 1.0
			}

			// Channel 18: Move counter (normalized)
			tensor[row][col][18] = float32(board.MoveCounter) / 100.0
		}
	}

	return tensor
}

func (board *Board) ExportTensorAsJSON() []byte {
  // exports Tensor for model
	tensor := board.ToTensor()
	jsonData, err := json.Marshal(tensor)
	if err != nil {
		return nil
	}
	return jsonData
}

