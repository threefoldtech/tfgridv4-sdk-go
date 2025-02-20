package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

var urls = map[string]string{
	"dev":  "https://registrar.dev4.grid.tf/v1",
	"qa":   "https://registrar.qa4.grid.tf/v1",
	"test": "https://registrar.test4.grid.tf/v1",
	"main": "https://registrar.prod4.grid.tf/v1",
}

func main() {
	var seed string
	var network string

	flag.StringVar(&seed, "seed", "", "seed")
	flag.StringVar(&network, "network", "", "network (dev, qa, test, main)")
	flag.Parse()

	u, ok := urls[network]
	if !ok {
		log.Fatal().Msgf("invalid network %s", network)
	}

	var publicKey ed25519.PublicKey
	var privateKey ed25519.PrivateKey
	var err error

	if len(seed) == 0 {
		s := make([]byte, 32)
		_, err := rand.Read(s)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		seed = hex.EncodeToString(s)
		fmt.Println("New Seed (Hex):", seed)
	}

	publicKey, privateKey, err = getKeyPair(seed)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	twinID, err := createAccount(privateKey, publicKey, u)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	fmt.Println(twinID)
}

func createAccount(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, u string) (twinID uint64, err error) {
	url, err := url.JoinPath(u, "accounts")
	if err != nil {
		return
	}

	timestamp := time.Now().Unix()
	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, publicKeyBase64))
	signature := ed25519.Sign(privateKey, challenge)

	data := map[string]any{
		"public_key": publicKey,
		"signature":  signature,
		"timestamp":  timestamp,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Post(url, "application/json", &body)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusCreated {
		return twinID, fmt.Errorf("account not created successfully")
	}

	defer resp.Body.Close()

	result := struct {
		TwinID uint64 `json:"twin_id"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.TwinID, err
}

func getKeyPair(seed string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, nil, err
	}

	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	return publicKey, privateKey, nil
}
