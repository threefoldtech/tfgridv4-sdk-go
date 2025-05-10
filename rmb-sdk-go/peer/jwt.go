package peer

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vedhavyas/go-subkey/v2"
)

const CustomSigning = "RMB"

var (
	_ jwt.SigningMethod = (*RmbSigner)(nil)
)

type RmbSigner struct{}

func (s *RmbSigner) Verify(signingString, signature string, key interface{}) error {
	panic("unimplemented")
}

func (s *RmbSigner) Sign(signingString string, key interface{}) (string, error) {
	identity, ok := key.(subkey.KeyPair)
	if !ok {
		return "", fmt.Errorf("invalid key expecting subkey keypair")
	}

	signature, err := Sign(identity, []byte(signingString))
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(signature), nil
}

func (s *RmbSigner) Alg() string {
	return "RS512"
}

func NewJWT(identity subkey.KeyPair, id uint32, session string, ttl uint32) (string, error) {
	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"sub": id,
		"iat": now,
		"exp": now + int64(ttl),
	}
	if session != "" {
		claims["sid"] = session
	}
	token := jwt.NewWithClaims(&RmbSigner{}, claims)

	return token.SignedString(identity)
}

func Sign(signer subkey.KeyPair, input []byte) ([]byte, error) {
	signature, err := signer.Sign(input)
	if err != nil {
		return nil, err
	}
	withType := make([]byte, len(signature)+1)

	keyType, err := getKeyPairType(signer)
	if err != nil {
		return nil, err
	}

	if keyType == KeyTypeSr25519 {
		withType[0] = []byte("s")[0] // edIdentity will return e, while sr will be s
	}

	if keyType == KeyTypeEd25519 {
		withType[0] = []byte("e")[0] // edIdentity will return e, while sr will be s
	}

	copy(withType[1:], signature)
	return withType, nil
}

func getKeyPairType(pair subkey.KeyPair) (string, error) {
	switch reflect.TypeOf(pair).String() {
	case "*sr25519.keyRing":
		return KeyTypeSr25519, nil
	case "ed25519.keyRing":
		return KeyTypeEd25519, nil
	default:
		return "", errors.New("unknown key type")
	}
}
