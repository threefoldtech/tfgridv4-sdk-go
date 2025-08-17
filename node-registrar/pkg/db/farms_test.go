package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateFarm(t *testing.T) {
	BeforeEach(t)

	t.Run("Create Farm", func(t *testing.T) {
		farm := Farm{
			FarmName:       "NewTestFarm",
			TwinID:         1,
			StellarAddress: "G" + strings.Repeat("C", 55),
			Dedicated:      true,
		}

		farmID, err := testDB.CreateFarm(farm)
		require.NoError(t, err)
		require.NotZero(t, farmID)

		createdFarm, err := testDB.GetFarm(farmID)
		require.NoError(t, err)
		assert.Equal(t, farm.FarmName, createdFarm.FarmName)
		assert.Equal(t, farm.TwinID, createdFarm.TwinID)
		assert.Equal(t, farm.StellarAddress, createdFarm.StellarAddress)
		assert.Equal(t, farm.Dedicated, createdFarm.Dedicated)
		assert.NotZero(t, createdFarm.CreatedAt)
		assert.NotZero(t, createdFarm.UpdatedAt)
	})

	t.Run("Create Farm with Duplicate Name", func(t *testing.T) {
		farm := Farm{
			FarmName:       "TestFarm1",
			TwinID:         2,
			StellarAddress: "G" + strings.Repeat("D", 55),
			Dedicated:      false,
		}

		farmID, err := testDB.CreateFarm(farm)
		require.Error(t, err)
		assert.Equal(t, ErrRecordAlreadyExists, err)
		assert.Zero(t, farmID)
	})

	t.Run("Create Farm with Invalid Twin ID", func(t *testing.T) {
		farm := Farm{
			FarmName:       "InvalidTwinIDFarm",
			TwinID:         999,
			StellarAddress: "G" + strings.Repeat("E", 55),
			Dedicated:      false,
		}

		farmID, err := testDB.CreateFarm(farm)
		require.Error(t, err)
		assert.Zero(t, farmID)
	})

	t.Run("Create Farm with Missing Required Fields", func(t *testing.T) {
		farm := Farm{
			FarmName:  "MissingFieldsFarm",
			Dedicated: false,
		}

		farmID, err := testDB.CreateFarm(farm)
		require.Error(t, err)
		assert.Zero(t, farmID)
	})
}

func TestGetFarm(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Existing Farm", func(t *testing.T) {
		farm, err := testDB.GetFarm(1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), farm.FarmID)
		assert.Equal(t, "TestFarm1", farm.FarmName)
		assert.Equal(t, uint64(1), farm.TwinID)
		assert.Equal(t, "G"+strings.Repeat("A", 55), farm.StellarAddress)
		assert.False(t, farm.Dedicated)

		farmVerify, err := testDB.GetFarm(1)
		require.NoError(t, err)
		assert.Equal(t, farm.FarmName, farmVerify.FarmName)
		assert.Equal(t, farm.TwinID, farmVerify.TwinID)
	})

	t.Run("Get Non-existent Farm", func(t *testing.T) {
		farm, err := testDB.GetFarm(999)
		assert.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
		assert.Equal(t, Farm{}, farm)
	})
}

func TestListFarms(t *testing.T) {
	BeforeEach(t)

	t.Run("List All Farms", func(t *testing.T) {
		farms, err := testDB.ListFarms(FarmFilter{}, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, farms, 2)
	})

	t.Run("List Farms with Farm Name Filter", func(t *testing.T) {
		farmName := "TestFarm1"
		filter := FarmFilter{FarmName: &farmName}
		farms, err := testDB.ListFarms(filter, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, farms, 1)
		assert.Equal(t, farmName, farms[0].FarmName)
	})

	t.Run("List Farms with Farm ID Filter", func(t *testing.T) {
		farmID := uint64(2)
		filter := FarmFilter{FarmID: &farmID}
		farms, err := testDB.ListFarms(filter, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, farms, 1)
		assert.Equal(t, farmID, farms[0].FarmID)
		assert.Equal(t, "TestFarm2", farms[0].FarmName)
	})

	t.Run("List Farms with Twin ID Filter", func(t *testing.T) {
		twinID := uint64(1)
		filter := FarmFilter{TwinID: &twinID}
		farms, err := testDB.ListFarms(filter, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, farms, 1)
		assert.Equal(t, twinID, farms[0].TwinID)
	})
}

func TestUpdateFarm(t *testing.T) {
	BeforeEach(t)

	t.Run("Update Both Name and Stellar Address", func(t *testing.T) {
		originalFarm, err := testDB.GetFarm(1)
		require.NoError(t, err)

		newName := "UpdatedTestFarm1"
		newStellarAddr := "G" + strings.Repeat("U", 55)
		err = testDB.UpdateFarm(1, newName, newStellarAddr)
		require.NoError(t, err)

		farm, err := testDB.GetFarm(1)
		require.NoError(t, err)
		assert.Equal(t, newName, farm.FarmName)
		assert.Equal(t, newStellarAddr, farm.StellarAddress)
		assert.Equal(t, originalFarm.TwinID, farm.TwinID)
		assert.Equal(t, originalFarm.Dedicated, farm.Dedicated)
	})

	t.Run("Update Non-existent Farm", func(t *testing.T) {
		err := testDB.UpdateFarm(999, "NonExistentFarm", "G"+strings.Repeat("X", 55))
		require.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
	})

	t.Run("Update with Empty Fields", func(t *testing.T) {
		originalFarm, err := testDB.GetFarm(1)
		require.NoError(t, err)

		err = testDB.UpdateFarm(originalFarm.FarmID, "", "")
		require.NoError(t, err)

		updatedFarm, err := testDB.GetFarm(originalFarm.FarmID)
		require.NoError(t, err)
		assert.Equal(t, originalFarm.FarmName, updatedFarm.FarmName)
		assert.Equal(t, originalFarm.StellarAddress, updatedFarm.StellarAddress)
		assert.Equal(t, originalFarm.TwinID, updatedFarm.TwinID)
		assert.Equal(t, originalFarm.Dedicated, updatedFarm.Dedicated)
	})

	t.Run("Update Farm Name to Duplicate Should Fail", func(t *testing.T) {
		farm1, err := testDB.GetFarm(1)
		require.NoError(t, err)
		farm2, err := testDB.GetFarm(2)
		require.NoError(t, err)

		err = testDB.UpdateFarm(farm1.FarmID, farm2.FarmName, "")
		require.Error(t, err)

		unchangedFarm, err := testDB.GetFarm(farm1.FarmID)
		require.NoError(t, err)
		assert.Equal(t, farm1.FarmName, unchangedFarm.FarmName)
		assert.NotEqual(t, farm2.FarmName, unchangedFarm.FarmName)
	})
}
