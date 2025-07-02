package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	chess "chess/board"
)

type InputProvider struct {}

func playTUI() {
	var board *chess.Board = chess.NewBoard()
  provider1 := InputProvider{}
  provider2 := InputProvider{}
  handler1 := OutputHandler{}
  handler2 := OutputHandler{}
  config := chess.GameConfig{}
  result, err := chess.CoreGameplayLoop(board, config, provider1, provider2, handler1, handler2)
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

var pieceMap = map[string]chess.Piece{
	"Pawn": chess.Pawns, "Pawns": chess.Pawns, "pawn": chess.Pawns, "pawns": chess.Pawns,
	"Knight": chess.Knights, "Knights": chess.Knights, "knight": chess.Knights, "knights": chess.Knights,
	"Bishop": chess.Bishops, "Bishops": chess.Bishops, "bishop": chess.Bishops, "bishops": chess.Bishops,
	"Rook": chess.Rooks, "Rooks": chess.Rooks, "rook": chess.Rooks, "rooks": chess.Rooks,
	"Queen": chess.Queens, "Queens": chess.Queens, "queen": chess.Queens, "queens": chess.Queens,
	"King": chess.Kings, "Kings": chess.Kings, "king": chess.Kings, "kings": chess.Kings,
}

func (t InputProvider) GetMove(board *chess.Board) (chess.Move, error) {
	fmt.Println("Please input move.")
  println(board.Turn)
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	var start chess.Square
	var end chess.Square
	var promotion chess.Piece
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return chess.Move{start, end, promotion}, nil
	}
	input = strings.TrimSpace(input)
	move := strings.Split(input, " ")
	if input == "resign" || input == "Resign" {
		return chess.Move{start, end, promotion}, chess.ErrResign
	}
	if len(move) < 2 || len(move) > 3 {
		fmt.Println("Please provide at two words if not promoting, and three words if promoting.")
		return chess.Move{start, end, promotion}, nil
	}
	if len(move) == 3 {
		promotion = pieceMap[move[2]]
	} else {
		promotion = chess.Empty
	}
	start = chess.NotationToIndex(move[0])
	end = chess.NotationToIndex(move[1])
	return chess.Move{start, end, promotion}, nil
}

type OutputHandler struct {}

func (handler OutputHandler) DisplayBoard(board *chess.Board) {
  println(board.PrintBoard())
}

func (handler OutputHandler) DisplayCheck() {
  println("Check!")
}

