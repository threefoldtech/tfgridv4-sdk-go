package client

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/pkg/errors"
)

func (c RegistrarClient) signRequest() (authHeader string) {
	timestamp := time.Now().Unix()
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, c.twinID))

	signature := ed25519.Sign(c.privateKey, challenge)

	authHeader = fmt.Sprintf(
		"%s:%s",
		base64.StdEncoding.EncodeToString(challenge),
		base64.StdEncoding.EncodeToString(signature),
	)
	return
}

func parseResponseError(body io.Reader) (err error) {
	errResp := struct {
		Error string `json:"error"`
	}{}

	err = json.NewDecoder(body).Decode(&errResp)
	if err != nil {
		return errors.Wrap(err, "failed to parse response error")
	}

	return errors.New(errResp.Error)
}
