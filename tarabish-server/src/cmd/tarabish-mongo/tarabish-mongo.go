package main

import (
	"context"
	"fmt"
	"go-tarabish/src/server/srvgorilla"
	"go-tarabish/src/store/storemongo"
	"time"
	// "go-tarabish/src/server/srvgorilla"
	// "go-tarabish/src/store/storemongo"
)

func main() {
	fmt.Println("Scopone with Mongo store started")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store := storemongo.Connect(ctx)

	srvgorilla.Start(store, store)
}
