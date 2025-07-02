package chess

import (
  "errors"
)

type (
  Input uint8
  )

type GameConfig struct {
  WhiteAI bool
  BlackAI bool
  WhiteInput Input
  BlackInput Input
}

type GameResult struct {
  Draw bool
  Winner Color
  Reason string
}

const ZobristKeysFilePath = "zobrist_keys.gob"

var ErrResign  = errors.New("Player resigned")

func CoreGameplayLoop(board *Board, config GameConfig, input1 InputProvider, input2 InputProvider, output1 OutputHandler, output2 OutputHandler) (GameResult, error) {
  _ = LoadZobristKeys()
  for {
    output1.DisplayBoard(board)
    output2.DisplayBoard(board)
    if board.IsCheckmate() {
      return GameResult{false, board.Turn.Other(), "Checkmate"}, nil
    }
    if board.IsStalemate() {
      return GameResult{Draw : true, Reason : "Stalemate"}, nil
    }
    if board.Is50Moves() {
      return GameResult{Draw : true, Reason : "50 move draw"}, nil
    }
    if board.IsCheck(board.Turn) {
      if board.Turn == White {
        output1.DisplayCheck()
      } else {
        output2.DisplayCheck()
      }
    }
    var move Move
    var err error
    if board.Turn == White {
      move, err = input1.GetMove(board)
    } else {
      move, err = input2.GetMove(board)
    }
    if err != nil {
      if errors.Is(err, ErrResign) {
        return GameResult{false, board.Turn.Other(), "Resignation"}, nil
      }
      continue
    }
    if !board.IsLegal(move) {
      continue
    }
    piece := board.GetPieceAt(move.Start, board.Turn)
    if piece == Empty {
      continue
    }
    if board.MovePiece(piece, move) {
      board.MoveCounter = 0
    }

    board.MoveCounter++
    board.TotalMoves++
    board.History[board.GetZobristHash()]++
    if board.IsThreefold() {
      return GameResult{Draw : true, Reason : "Threefold repetition"}, nil
    }
    board.Turn = board.Turn.Other()

  }
}

type InputProvider interface {
  GetMove(board *Board) (Move, error)
}

type OutputHandler interface {
  DisplayBoard(board *Board)
  DisplayCheck()
}


