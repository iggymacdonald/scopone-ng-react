package main

import (
	"fmt"
	"go-tarabish/src/game-logic/tarabish"
	"go-tarabish/src/server/srvgorilla"
	// "go-tarabish/src/game-logic/tarabish"
	// "go-tarabish/src/server/srvgorilla"
)

func main() {
	fmt.Println("Tarabish in memory (no database) started")

	srvgorilla.Start(&tarabish.DoNothingStore{}, &tarabish.DoNothingStore{})
}
