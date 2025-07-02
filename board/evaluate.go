package chess

import (
  "math/bits"
  "math"
)
func  Evaluate(board *Board) int {
  pawnValue := 100
  knightValue := 300
  bishopValue := 300
  rookValue := 500
  queenValue := 900
  score := 0
  totalPieces := bits.OnesCount64(uint64(board.FullBB))
	lateGame := totalPieces < 10
  if (board.Is50Moves() || board.IsThreefold() || board.IsStalemate()){

    return 0
  }
  if board.IsCheckmate() {
    if board.Turn == Black {
      return math.MaxInt32
    }
    return math.MinInt32
  }

  for c := White; c <= Black; c++ {
        multiplier := 1
        if c == Black {
            multiplier = -1
        }
    bishops := bits.OnesCount64(uint64(board.PieceBB[c][Bishops]))
    score += bits.OnesCount64(uint64(board.PieceBB[c][Pawns])) * pawnValue * multiplier
    score += bits.OnesCount64(uint64(board.PieceBB[c][Knights])) * knightValue * multiplier
    score += bishops * bishopValue * multiplier
    if bishops == 2 {
      score += 75 * multiplier
    }
    score += bits.OnesCount64(uint64(board.PieceBB[c][Rooks])) * rookValue * multiplier
    score += bits.OnesCount64(uint64(board.PieceBB[c][Queens])) * queenValue * multiplier
    //PST
		for p := Pawns; p <= Kings; p++ {
			pieceBB := board.PieceBB[c][p]
			for pieceBB != 0 {
				sq := Square(bits.TrailingZeros64(uint64(pieceBB)))
				pieceBB.ZeroBit(sq) // Clear the bit to process the next piece
				pstValue := PieceSquareTable(p, sq, c, lateGame)
				score += pstValue * multiplier
			}
		}
    // value of attacks and moves
    numLegalMoves := len(board.GetAllLegalMoves(c))
		mobilityWeight := 2
		score += numLegalMoves * mobilityWeight * multiplier
    numAttackedSquares := bits.OnesCount64(uint64(board.AllAttacks(c)))
		attackWeight := 1
		score += numAttackedSquares * attackWeight * multiplier
    }
  if board.Turn == Black {
    if score > 10000 {
    }
    return -score
  }
  return score
}
func PieceSquareTable(piece Piece, square Square, color Color, late bool) int {

var kingPSTMiddle = [64]int{
    20,  30,  10,   0,   0,  10,  30,  20, 
    20,  20,   0,   0,   0,   0,  20,  20,
    -10, -20, -20, -20, -20, -20, -20, -10,
    -20, -30, -30, -40, -40, -30, -30, -20,
    -30, -40, -40, -50, -50, -40, -40, -30,
    -30, -40, -40, -50, -50, -40, -40, -30,
    -30, -40, -40, -50, -50, -40, -40, -30,
    -30, -40, -40, -50, -50, -40, -40, -30,
}
var kingPSTLate = [64]int{
    -10, -10, -10, -10, -10, -10, -10, -10,
    -10,   0,   5,   5,   5,   5,   0, -10,
    -10,   5,  10,  10,  10,  10,   5, -10,
    -10,   5,  15,  15,  15,  15,   5, -10,
    -10,   5,  15,  15,  15,  15,   5, -10,
    -10,   5,  10,  10,  10,  10,   5, -10,
    -10,   0,   5,   5,   5,   5,   0, -10,
    -10, -10, -10, -10, -10, -10, -10, -10,
}
var bishopPST = [64]int{
    20, 10,  0,  0,  0,  0, 10, 20,
    10, 20, 10,  0,  0, 10, 20, 10,
     0, 10, 20, 10, 10, 20, 10,  0,
    0,  0, 10, 20, 20, 10,  0,  0,
     0,  0, 10, 20, 20, 10,  0,  0,
     0, 10, 20, 10, 10, 20, 10,  0, 
    10, 20, 10,  0,  0, 10, 20, 10, 
    20, 10,  0,  0,  0,  0, 10, 20,
}
var pawnPST = [64]int{
    0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0,  0,  0,  0,
    0,  0,  0,  0,  0, -5,  0,  0,
    0,  0,  5, 10, 10,  0,  0,  0,
    10, 10, 15, 20, 20, 10, 10, 10,
    20, 20, 25, 30, 30, 20, 20, 20,
    30, 30, 35, 40, 40, 30, 30, 30,
    0,  0,  0,  0,  0,  0,  0,  0,
}
	var knightPST = [64]int{
		-50, -40, -30, -30, -30, -30, -40, -50,
		-40, -20, 0, 0, 0, 0, -20, -40,
		-30, 0, 10, 15, 15, 10, 0, -30,
		-30, 5, 15, 20, 20, 15, 5, -30,
		-30, 0, 15, 20, 20, 15, 0, -30,
		-30, 5, 10, 15, 15, 10, 5, -30,
		-40, -20, 0, 5, 5, 0, -20, -40,
		-50, -40, -30, -30, -30, -30, -40, -50,
	}
var rookPST = [64]int{
    0,  0,  0,  10,  10,  0,  0,  0,
    0, 0, 0, 0, 10, 10, 0,  0,
    -5,  0,  0,  10,  10,  0,  0, -5,
    -5,  0,  0,  10,  10,  0,  0, -5,
    -5,  0,  0,  10,  10,  0,  0, -5,
    -5,  0,  0,  10,  10,  0,  0, -5, 
    10, 20, 20, 20, 20, 20, 20, 10, 
    10, 20, 20, 30, 30, 20, 20, 10,
}
var queenPST = [64]int{ 
    -20,-10,-10, -5, -5,-10,-10,-20,
    -10,  0,  0,  0,  0,  0,  0,-10,
    -10,  0,  5,  5,  5,  5,  0,-10,
     -5,  0,  5,  5,  5,  5,  0, -5,
      0,  0,  5,  5,  5,  5,  0, -5,
    -10,  5,  5,  5,  5,  5,  0,-10,
    -10,  0,  5,  0,  0,  0,  0,-10,
    -20,-10,-10, -5, -5,-10,-10,-20,
}
if color == Black {
  square = 63 - square
}
switch piece {
  case Pawns:
    return pawnPST[square]
  case Knights:
    return knightPST[square]
  case Bishops:
    return bishopPST[square]
  case Rooks:
    return rookPST[square]
  case Queens:
    return queenPST[square]
  case Kings:
    if late {
      return kingPSTLate[square]
    } else{
      return kingPSTMiddle[square]
    }
  }
  return 0
}
