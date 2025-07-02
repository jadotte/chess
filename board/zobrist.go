package chess

import (
  "math/rand"
  "time"
  "os"
  "fmt"
  "encoding/gob"
)

var zobristKeys struct {
  Pieces [2][7][64]uint64
  Castling [16]uint64
  EnPassant [8]uint64
  Turn uint64
}

func InitZobristKeys() {
  r := rand.New(rand.NewSource(time.Now().UnixNano()))
  for c := 0; c < 2; c ++ {
    for p := 0; p< 7; p++ {
      for s:= 0; s < 64; s++ {
        zobristKeys.Pieces[c][p][s] = r.Uint64()
      }
    }
  }
  for i := 0; i < 16; i++ {
    zobristKeys.Castling[i] = r.Uint64()
  }
  for i := 0; i < 8; i++ {
    zobristKeys.EnPassant[i] = r.Uint64()
  }
  zobristKeys.Turn = r.Uint64()
}

func SaveZobristKeys() error {
  file, err := os.Create(ZobristKeysFilePath)
  if err != nil {
    return fmt.Errorf("failed to create zobrist keys file: %w", err)
  }
  defer file.Close()

  encoder := gob.NewEncoder(file)
  if err := encoder.Encode(zobristKeys); err != nil {
    return fmt.Errorf("failed to encode zobrist keys: %w", err)
  }
  return nil
}
func LoadZobristKeys() error {
  file, err := os.Open(ZobristKeysFilePath)
  if err != nil {
    if os.IsNotExist(err) {
      InitZobristKeys()
        return SaveZobristKeys()
      }
    return fmt.Errorf("failed to open zobrist keys file: %w", err)
  }
  defer file.Close()

  decoder := gob.NewDecoder(file)
  if err := decoder.Decode(&zobristKeys); err != nil {
    return fmt.Errorf("failed to decode zobrist keys: %w", err)
  }
  return nil
}
