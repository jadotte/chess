package engine

import (
  "math"
  chess "chess/board"
  "fmt"
  "sort"
)

type EngineMove struct {
  Move chess.Move
  Score int
}

func PlayEngine(depth int) {
	var board *chess.Board = chess.NewBoard()
  provider1 := AlphaBetaInputProvider{depth}
  provider2 := AlphaBetaInputProvider{depth}
  handler1 := AlphaBetaOutputHandler{}
  handler2 := AlphaBetaOutputHandler{}
  config := chess.GameConfig{}
  result, err := chess.CoreGameplayLoop(board, config,provider1, provider2, handler1, handler2)
  if err != nil {
    println(err)
    return
  }
  if result.Draw {
    fmt.Printf("Draw by %s", result.Reason)
  } else if result.Winner == 0{
    fmt.Printf("White wins by %s", result.Reason)
  } else {
    fmt.Printf("Black wins by %s", result.Reason)
  }
}

func AlphaBetaSearch(board *chess.Board, alpha, beta, depth int) int {
  if (depth == 0 || board.IsCheckmate() || board.IsStalemate() || board.IsThreefold() || board.Is50Moves()) {
    eval := chess.Evaluate(board)
    return eval
  }
  color := board.Turn
  legalMoves := board.GetAllLegalMoves(color)
  if color == chess.White {
    maxEval := math.MinInt32
    for _, move := range legalMoves {
      simBoard := board.Clone()
      simBoard.MovePiece(simBoard.GetPieceAt(move.Start, simBoard.Turn), move)
      simBoard.Turn = simBoard.Turn.Other()
      simBoard.MoveCounter++
      simBoard.History[simBoard.GetZobristHash()]++
      eval := AlphaBetaSearch(simBoard, alpha, beta, depth - 1)
      maxEval = max(maxEval, eval)
      alpha = max(alpha, eval)
      if eval >= beta {
        break
      }
    }
  return maxEval
  } else {
    minEval := math.MaxInt32
    for _, move := range legalMoves {
      simBoard := board.Clone()
      simBoard.MovePiece(simBoard.GetPieceAt(move.Start, simBoard.Turn), move)
      simBoard.Turn = simBoard.Turn.Other()
      simBoard.MoveCounter++
      simBoard.History[simBoard.GetZobristHash()]++
      eval := AlphaBetaSearch(simBoard, alpha, beta, depth - 1)
      minEval = min(minEval, eval)
      beta = min(beta, eval)
      if eval <= alpha {
        break
      }
    }
  return minEval
  }
}

type AlphaBetaInputProvider struct {
    SearchDepth int // in half moves
}

type AlphaBetaOutputHandler struct {}

func (handler AlphaBetaOutputHandler) DisplayBoard(board *chess.Board) {
  println(board.PrintBoard())
}

func (handler AlphaBetaOutputHandler) DisplayCheck() {
  return
}

func (ab AlphaBetaInputProvider) GetMove(board *chess.Board) (chess.Move, error) {
  var bestMove chess.Move
  var bestEval int
  color := board.Turn
  alpha := math.MinInt32
  beta := math.MaxInt32
  if color == chess.White {
    bestEval = math.MinInt32
  } else {
    bestEval = math.MaxInt32
  }
  legalMoves := ab.sortMoves(board)

  for _, eMove := range legalMoves {
    simBoard := board.Clone()
    move := eMove.Move
    piece := simBoard.GetPieceAt(move.Start, simBoard.Turn)
    simBoard.MovePiece(piece, move)
    simBoard.Turn = simBoard.Turn.Other()
    simBoard.MoveCounter++
    simBoard.History[simBoard.GetZobristHash()]++
    eval := AlphaBetaSearch(simBoard, alpha, beta, ab.SearchDepth-1)
    if color == chess.White {
      if eval >= bestEval {
        bestEval = eval
        bestMove = move
      }
      alpha = max(alpha, eval)
    } else if color == chess.Black {
      if eval <= bestEval{
        bestEval = eval
        bestMove = move
    }
      beta= min(beta, eval)
    }
    if alpha >= beta {
      break}
  }
	fmt.Printf("Engine chose move: %v with evaluation: %d\n", bestMove, bestEval)
  println(board.TotalMoves)

  return bestMove, nil
}

func (ab AlphaBetaInputProvider) sortMoves(board *chess.Board) []EngineMove {
  color := board.Turn
  legalMoves := board.GetAllLegalMoves(color)
  var engineMoves []EngineMove
  for i := range legalMoves {
    move := &legalMoves[i]
    eMove := EngineMove{legalMoves[i], 0}
    capturedPiece := board.GetPieceAt(move.End, color.Other())
    if capturedPiece != chess.Empty {
      switch capturedPiece {
      case chess.Pawns:
        eMove.Score += 100
      case chess.Knights:
        eMove.Score += 300
      case chess.Bishops:
        eMove.Score += 330
      case chess.Rooks:
        eMove.Score += 500
      case chess.Queens:
        eMove.Score += 900
      }
    }
    if move.Promotion != chess.Empty {
      eMove.Score += 900
    }
    simBoard := board.Clone()
		_ = simBoard.MovePiece(simBoard.GetPieceAt(move.Start, simBoard.Turn), *move)
    if simBoard.IsCheck(color.Other()) {
      eMove.Score += 100
    }
    if !board.IsAttacked(color.Other(), move.End) {
      eMove.Score += 500
    }
    // PST

    engineMoves = append(engineMoves, eMove)
  }
  sort.Slice(engineMoves, func(i, j int) bool {
		return engineMoves[i].Score > engineMoves[j].Score
	})
  return engineMoves
}


func max(a, b int) int {
  if a >= b {
    return a
  }
  return b
}

func min(a, b int) int {
  if a <= b {
    return a
  }
  return b
}

