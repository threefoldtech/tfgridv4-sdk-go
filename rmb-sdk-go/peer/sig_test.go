package peer

import (
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/threefoldtech/tfgridv4-sdk-go/rmb-sdk-go/peer/types"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

const sigVerifyAccMnemonics = "garage dad improve reunion girl saddle theory know label reason fantasy deputy"
const sigVerifyAccTwinID = uint32(1171)

func TestSignature(t *testing.T) {

	identity, err := subkey.DeriveKeyPair(sr25519.Scheme{}, sigVerifyAccMnemonics)
	if err != nil {
		t.Fatalf("could not init new identity: %s", err)
	}

	env := types.Envelope{
		Uid:         uuid.NewString(),
		Timestamp:   uint64(time.Now().Unix()),
		Expiration:  10000,
		Destination: &types.Address{Twin: 10},
	}

	env.Message = &types.Envelope_Request{
		Request: &types.Request{
			Command: "cmd",
		},
	}
	env.Payload = &types.Envelope_Plain{
		Plain: []byte("my data"),
	}

	t.Run("valid signature", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		env.Source = &types.Address{
			Twin: sigVerifyAccTwinID,
		}

		toSign, err := Challenge(&env)
		assert.NoError(t, err)

		env.Signature, err = Sign(identity, toSign)
		assert.NoError(t, err)

		twinDB := NewMockTwinDB(ctrl)
		twinDB.EXPECT().Get(sigVerifyAccTwinID).Return(Twin{
			ID:        sigVerifyAccTwinID,
			PublicKey: identity.Public(),
		}, nil)

		err = VerifySignature(twinDB, &env)
		assert.NoError(t, err)
	})

	t.Run("invalid source", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		env.Source = &types.Address{
			Twin: 2,
		}

		toSign, err := Challenge(&env)
		assert.NoError(t, err)

		env.Signature, err = Sign(identity, toSign)
		assert.NoError(t, err)

		twinDB := NewMockTwinDB(ctrl)
		twinDB.EXPECT().Get(uint32(2)).Return(Twin{
			ID:        2,
			PublicKey: []byte("gibberish"),
		}, nil)

		err = VerifySignature(twinDB, &env)
		assert.Error(t, err)
	})

	t.Run("invalid signature", func(t *testing.T) {
		ctrl := gomock.NewController(t)

		env.Source = &types.Address{
			Twin: sigVerifyAccTwinID,
		}

		env.Signature = []byte("s13p49fnaskdjnv")

		twinDB := NewMockTwinDB(ctrl)
		twinDB.EXPECT().Get(sigVerifyAccTwinID).Return(Twin{
			ID:        sigVerifyAccTwinID,
			PublicKey: identity.Public(),
		}, nil)

		err = VerifySignature(twinDB, &env)
		assert.Error(t, err)
	})
}
