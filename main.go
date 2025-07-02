package main

import (
  chess "chess/board"
  tui "chess/tui"
  engine "chess/engine"
  "fmt"
)

func playEngine(depth int) {
	var board *chess.Board = chess.NewBoard()
  provider1 := tui.InputProvider{}
  provider2 := engine.AlphaBetaInputProvider{depth}
  handler1 := tui.OutputHandler{}
  handler2 := engine.AlphaBetaOutputHandler{}
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
func main() {
  playEngine(5)
}
