package main

import (
	"bytes"
	"crypto/ed25519"
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
	var name string

	flag.StringVar(&seed, "seed", "", "seed")
	flag.StringVar(&network, "network", "", "network (dev, qa, test, main)")
	flag.StringVar(&name, "farm-name", "", "farm name")
	flag.Parse()

	u, ok := urls[network]
	if !ok {
		log.Fatal().Msgf("invalid network %s", network)
	}

	publicKey, privateKey, err := getKeyPair(seed)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	twinID, err := getAccount(publicKey, u)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	result, err := createFarm(name, twinID, privateKey, u)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	fmt.Println(result)
}

func createFarm(name string, twinID uint64, privateKey ed25519.PrivateKey, u string) (farmID uint64, err error) {
	url, err := url.JoinPath(u, "farms")
	if err != nil {
		return
	}
	data := map[string]any{
		"farm_name": name,
		"twin_id":   twinID,
		"dedicated": false,
	}

	var body bytes.Buffer
	err = json.NewEncoder(&body).Encode(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return
	}

	timestamp := time.Now().Unix()
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, twinID))
	signature := ed25519.Sign(privateKey, challenge)

	authHeader := fmt.Sprintf(
		"%s:%s",
		base64.StdEncoding.EncodeToString(challenge),
		base64.StdEncoding.EncodeToString(signature),
	)
	req.Header.Set("X-Auth", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return farmID, fmt.Errorf("could not create a farm")
	}

	result := struct {
		FarmID uint64 `json:"farm_id"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}
	fmt.Println(result)
	return result.FarmID, nil
}

func getAccount(publicKey []byte, u string) (twinID uint64, err error) {
	url, err := url.JoinPath(u, "accounts")
	if err != nil {
		return
	}

	publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	q := req.URL.Query()
	q.Add("public_key", publicKeyBase64)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return twinID, fmt.Errorf("status code not ok")
	}

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
