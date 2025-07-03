package chess

import (
	"fmt"
	"math/bits"
	"strings"
  "strconv"
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

func NewBoardFromFEN(fen string) (*Board, error) {
	b := NewBoard()
	b.ClearBoard()

	parts := strings.Split(fen, " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid FEN string")
	}

	ranks := strings.Split(parts[0], "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("incorrect number of ranks")
	}

	for rIdx, rankStr := range ranks {
		fileIdx := 0
		for _, char := range rankStr {
			if char >= '1' && char <= '8' {
				fileIdx += int(char - '0')
			} else {
				sq := Square((7-rIdx)*8 + fileIdx)
				piece, color, err := charToPiece(char)
				if err != nil {
					return nil, err
				}
				b.PieceBB[color][piece].SetBit(sq)
				fileIdx++
			}
		}
	}

	// color
	if len(parts) > 1 {
		if parts[1] == "w" {
			b.Turn = White
		} else if parts[1] == "b" {
			b.Turn = Black
		} else {
			return nil, fmt.Errorf("invalid color")
		}
	}

	// Castling
	if len(parts) > 2 {
		castlingRights := parts[2]
		b.RKRmoved = [2][3]bool{} 
		
		if castlingRights != "-" {
			if !strings.ContainsRune(castlingRights, 'K') { b.RKRmoved[White][2] = true } // White King side rook (h1) moved
			if !strings.ContainsRune(castlingRights, 'Q') { b.RKRmoved[White][0] = true } // White Queen side rook (a1) moved
			if !strings.ContainsRune(castlingRights, 'k') { b.RKRmoved[Black][2] = true } // Black King side rook (h8) moved
			if !strings.ContainsRune(castlingRights, 'q') { b.RKRmoved[Black][0] = true } // Black Queen side rook (a8) moved
		}
		if !strings.ContainsRune(castlingRights, 'K') && !strings.ContainsRune(castlingRights, 'Q') {
			b.RKRmoved[White][1] = true
		}
		if !strings.ContainsRune(castlingRights, 'k') && !strings.ContainsRune(castlingRights, 'q') {
			b.RKRmoved[Black][1] = true
		}
	}


	// En Passant
	if len(parts) > 3 {
		epSquareStr := parts[3]
		if epSquareStr != "-" {
			sq := NotationToIndex(epSquareStr)
			b.EnPassantSquare = &sq
		} else {
			b.EnPassantSquare = nil
		}
	}

	// 50 moes
	if len(parts) > 4 {
		halfMoveClock, err := strconv.Atoi(parts[4])
		if err != nil {
			return nil, fmt.Errorf("invalid FEN string")
		}
		b.MoveCounter = uint8(halfMoveClock)
	}

  if len(parts) > 5 {
		clock, err := strconv.Atoi(parts[5])
		if err != nil {
			return nil, fmt.Errorf("invalid FEN string")
		}
		b.TotalMoves = uint8(clock)
	}
	b.CombineBB()
	b.History[b.GetZobristHash()] = 1

	return b, nil
}

func (b *Board) ClearBoard() {
	for c := White; c <= Black; c++ {
		for p := Pawns; p <= Kings; p++ {
			b.PieceBB[c][p] = 0
		}
		b.ColorBB[c] = 0
	}
	b.FullBB = 0
	b.Turn = White
	b.EnPassantSquare = nil
	b.MoveCounter = 0
	b.TotalMoves = 0
	b.History = make(map[uint64]int)
	b.RKRmoved = [2][3]bool{}
}

func charToPiece(char rune) (Piece, Color, error) {
	switch char {
	case 'P': return Pawns, White, nil
	case 'N': return Knights, White, nil
	case 'B': return Bishops, White, nil
	case 'R': return Rooks, White, nil
	case 'Q': return Queens, White, nil
	case 'K': return Kings, White, nil
	case 'p': return Pawns, Black, nil
	case 'n': return Knights, Black, nil
	case 'b': return Bishops, Black, nil
	case 'r': return Rooks, Black, nil
	case 'q': return Queens, Black, nil
	case 'k': return Kings, Black, nil
	default: return Empty, White, fmt.Errorf("invalid FEN piece character: %c", char)
	}
}



func IndexToNotation(sq Square) string {
  var indexToNotationMap = map[Square]string{
	0: "a1", 1: "b1", 2: "c1", 3: "d1", 4: "e1", 5: "f1", 6: "g1", 7: "h1",
	8: "a2", 9: "b2", 10: "c2", 11: "d2", 12: "e2", 13: "d2", 14: "g2", 15: "h2",
	16: "a3", 17: "b3", 18: "c3", 19: "d3", 20: "e3", 21: "f3", 22: "g3", 23: "h3",
	24: "a4", 25: "b4", 26: "c4", 27: "d4", 28: "e4", 29: "f4", 30: "g4", 31: "h4",
	32: "a5", 33: "b5", 34: "c5", 35: "d5", 36: "e5", 37: "f5", 38: "g5", 39: "h5",
	40: "a6", 41: "b6", 42: "c6", 43: "d6", 44: "e6", 45: "f6", 46: "g6", 47: "h6",
	48: "a7", 49: "b7", 50: "c7", 51: "d7", 52: "e7", 53: "f7", 54: "g7", 55: "h7",
	56: "a8", 57: "b8", 58: "c8", 59: "d8", 60: "e8", 61: "f8", 62: "g8", 63: "h8",
}
	return indexToNotationMap[sq]
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

func PieceToChar(p Piece, c Color) string {
	char := ""
	switch p {
	case Pawns: char = "p"
	case Knights: char = "n"
	case Bishops: char = "b"
	case Rooks: char = "r"
	case Queens: char = "q"
	case Kings: char = "k"
	default: return ""
	}
	return char
}
