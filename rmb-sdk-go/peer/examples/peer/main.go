package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/rmb-sdk-go/peer"
	"github.com/threefoldtech/tfgrid-sdk-go/rmb-sdk-go/peer/types"
)

var resultsChan = make(chan bool)

func app() error {
	mnemonic := "<mnemonics goes here>"
	ctx := context.Background()

	peer, err := peer.NewPeer(
		ctx,
		mnemonic,
		relayCallback,
		peer.WithRegistrarUrl("https://registrar.dev4.grid.tf"),
		peer.WithRelay("wss://relay.dev.grid.tf"),
		peer.WithSession("test-client"),
		peer.WithInMemoryExpiration(6*60*60), // six hours
		peer.WithKeyType(peer.KeyTypeEd25519),
	)

	if err != nil {
		return fmt.Errorf("failed to create direct client: %w", err)
	}

	const dst = 7 // <- replace this with the twin id of where the service is running

	for i := 0; i < 20; i++ {
		data := []float64{rand.Float64(), rand.Float64()}
		var session *string
		// uncomment if you are using peer router example
		// routerSession := "test-router"
		// session = &routerSession

		if err := peer.SendRequest(ctx, uuid.NewString(), dst, session, "calculator.add", data); err != nil {
			return err
		}
	}
	for i := 0; i < 20; i++ {
		<-resultsChan
	}

	return nil
}

func relayCallback(ctx context.Context, p *peer.Peer, response *types.Envelope, callBackErr error) {
	output, err := peer.Json(response, callBackErr)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	fmt.Printf("output: %s\n", string(output))
	resultsChan <- true
}

func main() {
	if err := app(); err != nil {
		log.Fatal().Msg(err.Error())
	}
}
