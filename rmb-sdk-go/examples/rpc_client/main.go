package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/threefoldtech/tfgridv4-sdk-go/rmb-sdk-go/peer"
)

type version struct {
	ZOS   string `json:"zos"`
	ZInit string `json:"zinit"`
}

func app() error {
	mnemonic := "<mnemonics goes here>"

	client, err := peer.NewRpcClient(
		context.Background(),
		mnemonic,
		peer.WithRegistrarUrl("https://registrar.dev4.grid.tf"),
		peer.WithRelay("wss://relay.dev.grid.tf"),
		peer.WithSession("test-client"),
	)
	if err != nil {
		return fmt.Errorf("failed to create direct client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	const dstTwin uint32 = 11 // <- replace this with any node i
	var ver version
	if err := client.Call(ctx, dstTwin, "zos.system.version", nil, &ver); err != nil {
		return err
	}

	fmt.Printf("output: %s\n", ver)
	return nil
}

func main() {
	if err := app(); err != nil {
		log.Fatal(err)
	}
}
