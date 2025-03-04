package client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func (c *RegistrarClient) signRequest(timestamp int64) (authHeader string, err error) {
	challenge := []byte(fmt.Sprintf("%d:%v", timestamp, c.twinID))
	signature, err := c.keyPair.Sign(challenge)
	if err != nil {
		return "", err
	}

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
