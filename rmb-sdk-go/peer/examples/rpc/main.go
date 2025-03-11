package main

import (
	"context"
	"fmt"
	"time"

	"log"

	"github.com/threefoldtech/tfgrid-sdk-go/rmb-sdk-go/peer"
)

func app() error {
	mnemonic := "<mnemonics goes here>"

	client, err := peer.NewRpcClient(
		context.Background(),
		mnemonic,
		peer.WithRegistrarUrl("https://registrar.dev4.grid.tf"),
		peer.WithKeyType(peer.KeyTypeSr25519),
		peer.WithRelay("wss://relay.dev.grid.tf"),
		peer.WithSession("test-client"),
	)
	if err != nil {
		return fmt.Errorf("failed to create direct client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// NOTE: we calling the service 'calculator' session
	// as per the router_server example
	service := "calculator"
	const dst = 7 // <- replace this with the twin id of where the service is running
	// it's okay to run both the server and the client behind the same rmb-peer
	var output float64
	for i := 0; i < 20; i++ {
		// uncomment it  you are using peer router example
		// routerSession := "test-router"
		// session = &routerSession

		if err := client.CallWithSession(ctx, dst, &service, "calculator.add", []float64{output, float64(i)}, &output); err != nil {
			return err
		}
	}

	fmt.Printf("output: %f\n", output)

	return nil
}

func main() {
	if err := app(); err != nil {
		log.Fatal(err)
	}
}
