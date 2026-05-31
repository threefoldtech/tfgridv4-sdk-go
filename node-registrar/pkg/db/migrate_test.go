package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateNodeLastSeen(t *testing.T) {
	BeforeEach(t)

	t.Run("Migrate Nodes with Null LastSeen", func(t *testing.T) {
		testDB.gormDB.Model(&Node{}).Where("node_id IN (?)", []uint64{1, 2}).Update("last_seen", nil)

		err := testDB.MigrateNodeLastSeen()
		require.NoError(t, err)

		var updatedNullCount int64
		testDB.gormDB.Model(&Node{}).Where("last_seen IS NULL").Count(&updatedNullCount)
		assert.Equal(t, int64(1), updatedNullCount)

		// have uptime report should be updated
		node1, err := testDB.GetNode(1)
		require.NoError(t, err)
		assert.False(t, node1.LastSeen.IsZero())

		// no uptime report should not be updated
		node2, err := testDB.GetNode(2)
		require.NoError(t, err)
		assert.True(t, node2.LastSeen.IsZero())
	})

	t.Run("Migrate Nodes with Zero Time LastSeen", func(t *testing.T) {
		zeroTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
		testDB.gormDB.Model(&Node{}).Where("node_id = ?", 3).Update("last_seen", zeroTime)

		err := testDB.MigrateNodeLastSeen()
		require.NoError(t, err)

		node3, err := testDB.GetNode(3)
		require.NoError(t, err)
		assert.False(t, node3.LastSeen.Equal(zeroTime))
		assert.False(t, node3.LastSeen.IsZero())
	})

	t.Run("Skip Nodes with Valid LastSeen", func(t *testing.T) {
		originalTime := time.Now().Add(-1 * time.Hour)
		testDB.gormDB.Model(&Node{}).Where("node_id = ?", 1).Update("last_seen", originalTime)

		err := testDB.MigrateNodeLastSeen()
		require.NoError(t, err)

		node1, err := testDB.GetNode(1)
		require.NoError(t, err)
		assert.WithinDuration(t, originalTime, node1.LastSeen, time.Second)
	})

	t.Run("Skip Nodes without Uptime Reports", func(t *testing.T) {
		newNode := Node{
			FarmID: 1,
			TwinID: 4,
			Location: Location{
				Country:   "IT",
				City:      "Rome",
				Longitude: "12.4964",
				Latitude:  "41.9028",
			},
			Resources: Resources{
				HRU: 1000,
				SRU: 500,
				CRU: 8,
				MRU: 16,
			},
			Interfaces: []Interface{
				{
					Name: "eth0",
					Mac:  "00:11:22:33:44:CC",
					IPs:  []string{"192.168.1.203"},
				},
			},
			SerialNumber: "SN555",
		}

		nodeID, err := testDB.RegisterNode(newNode)
		require.NoError(t, err)

		err = testDB.MigrateNodeLastSeen()
		require.NoError(t, err)

		migratedNode, err := testDB.GetNode(nodeID)
		require.NoError(t, err)
		assert.True(t, migratedNode.LastSeen.IsZero())
	})
}
