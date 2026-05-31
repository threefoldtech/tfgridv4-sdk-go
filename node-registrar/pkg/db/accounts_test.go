package db

import (
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAccount(t *testing.T) {
	BeforeEach(t)

	t.Run("Create Account", func(t *testing.T) {
		account := Account{
			PublicKey: "test_public_key_new_1",
			Relays:    pq.StringArray{"relay4.example.com"},
			RMBEncKey: "test_rmb_key_4",
		}

		err := testDB.CreateAccount(&account)
		require.NoError(t, err)
		require.NotZero(t, account.TwinID)
		require.NotZero(t, account.CreatedAt)
		require.NotZero(t, account.UpdatedAt)

		createdAccount, err := testDB.GetAccount(account.TwinID)
		require.NoError(t, err)
		require.Equal(t, account.PublicKey, createdAccount.PublicKey)
		require.Equal(t, account.Relays, createdAccount.Relays)
		require.Equal(t, account.RMBEncKey, createdAccount.RMBEncKey)
	})

	t.Run("Create Account with Empty PublicKey", func(t *testing.T) {
		account := Account{
			PublicKey: "",
			Relays:    pq.StringArray{"relay5.example.com"},
			RMBEncKey: "test_rmb_key_5",
		}

		err := testDB.CreateAccount(&account)
		require.Error(t, err)
		require.Contains(t, err.Error(), "public key cannot be empty")
	})

	t.Run("Create Account with Missing Optional Fields", func(t *testing.T) {
		account := Account{
			PublicKey: "test_public_key_optional",
		}

		err := testDB.CreateAccount(&account)
		require.NoError(t, err)
		require.NotZero(t, account.TwinID)

		createdAccount, err := testDB.GetAccount(account.TwinID)
		require.NoError(t, err)
		require.Equal(t, account.PublicKey, createdAccount.PublicKey)
		require.Equal(t, pq.StringArray{}, createdAccount.Relays)
		require.Equal(t, "", createdAccount.RMBEncKey)
	})

	t.Run("Create Duplicate Account", func(t *testing.T) {
		duplicatePublicKey := "test_public_key_duplicate"

		account1 := Account{
			PublicKey: duplicatePublicKey,
		}

		err := testDB.CreateAccount(&account1)
		require.NoError(t, err)
		require.NotZero(t, account1.TwinID)

		createdAccount, err := testDB.GetAccount(account1.TwinID)
		require.NoError(t, err)
		require.Equal(t, duplicatePublicKey, createdAccount.PublicKey)

		account2 := Account{
			PublicKey: duplicatePublicKey,
		}

		err = testDB.CreateAccount(&account2)
		require.Error(t, err)
	})
}

func TestGetAccount(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Existing Account", func(t *testing.T) {
		account, err := testDB.GetAccount(1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), account.TwinID)
		assert.Equal(t, "test_public_key_1", account.PublicKey)
		assert.Equal(t, pq.StringArray{"relay1.example.com", "relay2.example.com"}, account.Relays)
		assert.Equal(t, "test_rmb_key_1", account.RMBEncKey)
	})

	t.Run("Get Non-existent Account", func(t *testing.T) {
		account, err := testDB.GetAccount(999)
		assert.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
		assert.Equal(t, Account{}, account)
	})
}

func TestGetAccountByPublicKey(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Account by Existing Public Key", func(t *testing.T) {
		account, err := testDB.GetAccountByPublicKey("test_public_key_2")
		assert.NoError(t, err)
		assert.Equal(t, uint64(2), account.TwinID)
		assert.Equal(t, "test_public_key_2", account.PublicKey)
	})

	t.Run("Get Account by Non-existent Public Key", func(t *testing.T) {
		account, err := testDB.GetAccountByPublicKey("non_existent_key")
		assert.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
		assert.Equal(t, Account{}, account)
	})
}

func TestUpdateAccount(t *testing.T) {
	BeforeEach(t)

	t.Run("Update Existing Account", func(t *testing.T) {
		newRelays := pq.StringArray{"updated_relay1.com", "updated_relay2.com"}
		newRMBEncKey := "updated_rmb_key"

		err := testDB.UpdateAccount(1, newRelays, newRMBEncKey)
		assert.NoError(t, err)

		account, err := testDB.GetAccount(1)
		assert.NoError(t, err)
		assert.Equal(t, newRelays, account.Relays)
		assert.Equal(t, newRMBEncKey, account.RMBEncKey)
	})

	t.Run("Update Non-existent Account", func(t *testing.T) {
		newRelays := pq.StringArray{"some_relay.com"}
		newRMBEncKey := "some_key"

		err := testDB.UpdateAccount(999, newRelays, newRMBEncKey)
		assert.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}
