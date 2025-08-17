package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListNodes(t *testing.T) {
	BeforeEach(t)

	t.Run("List All Nodes", func(t *testing.T) {
		nodes, err := testDB.ListNodes(NodeFilter{}, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, nodes, 3)
	})

	t.Run("List Nodes with Node ID Filter", func(t *testing.T) {
		nodeID := uint64(1)
		filter := NodeFilter{NodeID: &nodeID}
		nodes, err := testDB.ListNodes(filter, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, nodes, 1)
		assert.Equal(t, nodeID, nodes[0].NodeID)
		assert.Equal(t, "SN001", nodes[0].SerialNumber)
	})

	t.Run("List Nodes with Non-matching Filter", func(t *testing.T) {
		nodeID := uint64(999)
		filter := NodeFilter{NodeID: &nodeID}
		nodes, err := testDB.ListNodes(filter, DefaultLimit())
		require.NoError(t, err)
		assert.Len(t, nodes, 0)
	})
}

func TestGetNode(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Existing Node", func(t *testing.T) {
		originalNode := testData.Nodes[0]
		node, err := testDB.GetNode(1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), node.NodeID)
		assert.Equal(t, originalNode.FarmID, node.FarmID)
		assert.Equal(t, originalNode.TwinID, node.TwinID)
		assert.Equal(t, originalNode.SerialNumber, node.SerialNumber)
		assert.Equal(t, originalNode.Location.Country, node.Location.Country)
		assert.Equal(t, originalNode.Resources.CRU, node.Resources.CRU)
		assert.Len(t, node.Interfaces, 1)
		assert.Equal(t, originalNode.Interfaces[0].Name, node.Interfaces[0].Name)

		var nodeGorm Node
		err = testDB.gormDB.Where("node_id = ?", 1).First(&nodeGorm).Error
		require.NoError(t, err)
		assert.Equal(t, node.NodeID, nodeGorm.NodeID)
	})

	t.Run("Get Non-existent Node", func(t *testing.T) {
		_, err := testDB.GetNode(999)
		require.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}

func TestRegisterNode(t *testing.T) {
	BeforeEach(t)

	t.Run("Register Valid Node", func(t *testing.T) {
		newNode := Node{
			FarmID: 1,
			TwinID: 4,
			Location: Location{
				Country:   "CA",
				City:      "Toronto",
				Longitude: "-79.3470",
				Latitude:  "43.6532",
			},
			Resources: Resources{
				HRU: 4000,
				SRU: 2000,
				CRU: 64,
				MRU: 128,
			},
			Interfaces: []Interface{
				{
					Name: "eth0",
					Mac:  "00:11:22:33:44:99",
					IPs:  []string{"192.168.1.200"},
				},
			},
			SecureBoot:   true,
			Virtualized:  false,
			SerialNumber: "SN999",
			Approved:     true,
		}

		nodeID, err := testDB.RegisterNode(newNode)
		require.NoError(t, err)
		assert.Greater(t, nodeID, uint64(0))

		node, err := testDB.GetNode(nodeID)
		require.NoError(t, err)
		assert.Equal(t, newNode.FarmID, node.FarmID)
		assert.Equal(t, newNode.TwinID, node.TwinID)
		assert.Equal(t, newNode.SerialNumber, node.SerialNumber)
		assert.Equal(t, newNode.Location.Country, node.Location.Country)
	})

	t.Run("Register Node with Duplicate TwinID", func(t *testing.T) {
		newNode := Node{
			FarmID: 1,
			TwinID: 1,
			Location: Location{
				Country:   "FR",
				City:      "Paris",
				Longitude: "2.3522",
				Latitude:  "48.8566",
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
					Mac:  "00:11:22:33:44:AA",
					IPs:  []string{"192.168.1.201"},
				},
			},
			SerialNumber: "SN888",
		}

		_, err := testDB.RegisterNode(newNode)
		require.Error(t, err)
	})

	t.Run("Register Node with Empty Interfaces", func(t *testing.T) {
		newNode := Node{
			FarmID: 1,
			TwinID: 4,
			Location: Location{
				Country:   "UK",
				City:      "London",
				Longitude: "-0.1276",
				Latitude:  "51.5074",
			},
			Resources: Resources{
				HRU: 1000,
				SRU: 500,
				CRU: 8,
				MRU: 16,
			},
			Interfaces:   []Interface{},
			SerialNumber: "SN777",
		}

		_, err := testDB.RegisterNode(newNode)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "interfaces must not be empty")
	})

	t.Run("Register Node with Invalid FarmID", func(t *testing.T) {
		newNode := Node{
			FarmID: 999,
			TwinID: 4,
			Location: Location{
				Country:   "ES",
				City:      "Madrid",
				Longitude: "-3.7038",
				Latitude:  "40.4168",
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
					Mac:  "00:11:22:33:44:BB",
					IPs:  []string{"192.168.1.202"},
				},
			},
			SerialNumber: "SN666",
		}

		_, err := testDB.RegisterNode(newNode)
		require.Error(t, err)
	})
}

func TestUpdateNode(t *testing.T) {
	BeforeEach(t)

	t.Run("Update Node Successfully", func(t *testing.T) {
		originalNode, err := testDB.GetNode(1)
		require.NoError(t, err)

		updateNode := Node{
			Location: Location{
				Country:   "AU",
				City:      "Sydney",
				Longitude: "151.2093",
				Latitude:  "-33.8688",
			},
			Resources: Resources{
				HRU: 5000,
				SRU: 2500,
				CRU: 128,
				MRU: 256,
			},
			SecureBoot:   false,
			Virtualized:  true,
			SerialNumber: "SN_UPDATED",
			Approved:     false,
		}

		err = testDB.UpdateNode(1, updateNode)
		require.NoError(t, err)

		updatedNode, err := testDB.GetNode(1)
		require.NoError(t, err)
		assert.Equal(t, updateNode.Location.Country, updatedNode.Location.Country)
		assert.Equal(t, updateNode.Resources.CRU, updatedNode.Resources.CRU)
		assert.Equal(t, updateNode.SerialNumber, updatedNode.SerialNumber)
		assert.Equal(t, originalNode.FarmID, updatedNode.FarmID)
		assert.Equal(t, originalNode.TwinID, updatedNode.TwinID)

		var nodeGorm Node
		err = testDB.gormDB.Where("node_id = ?", 1).First(&nodeGorm).Error
		require.NoError(t, err)
		assert.Equal(t, updateNode.SerialNumber, nodeGorm.SerialNumber)
	})

	t.Run("Update Non-existent Node", func(t *testing.T) {
		updateNode := Node{
			SerialNumber: "NON_EXISTENT",
		}

		err := testDB.UpdateNode(999, updateNode)
		require.Error(t, err)
		assert.Equal(t, ErrRecordNotFound, err)
	})
}

func TestGetUptimeReports(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Uptime Reports for Node", func(t *testing.T) {
		start := time.Now().Add(-2 * time.Hour)
		end := time.Now()

		reports, err := testDB.GetUptimeReports(1, start, end)
		require.NoError(t, err)
		assert.Len(t, reports, 2)

		for _, report := range reports {
			assert.Equal(t, uint64(1), report.NodeID)
			assert.True(t, report.Timestamp.After(start) || report.Timestamp.Equal(start))
			assert.True(t, report.Timestamp.Before(end) || report.Timestamp.Equal(end))
		}
	})

	t.Run("Get Uptime Reports for Node with No Reports", func(t *testing.T) {
		start := time.Now().Add(-2 * time.Hour)
		end := time.Now()

		reports, err := testDB.GetUptimeReports(2, start, end)
		require.NoError(t, err)
		assert.Len(t, reports, 0)
	})
}

func TestCreateUptimeReport(t *testing.T) {
	BeforeEach(t)

	t.Run("Create Uptime Report and Updated LastSeen", func(t *testing.T) {
		originalNode, err := testDB.GetNode(1)
		require.NoError(t, err)

		newTimestamp := time.Now()
		report := &UptimeReport{
			NodeID:     1,
			Duration:   72 * time.Hour,
			Timestamp:  newTimestamp,
			WasRestart: false,
		}

		err = testDB.CreateUptimeReport(report)
		require.NoError(t, err)

		reports, err := testDB.GetUptimeReports(1, newTimestamp.Add(-time.Minute), newTimestamp.Add(time.Minute))
		require.NoError(t, err)
		require.NotEmpty(t, reports)
		assert.Equal(t, report.Duration, reports[len(reports)-1].Duration)

		updatedNode, err := testDB.GetNode(1)
		require.NoError(t, err)
		assert.True(t, updatedNode.LastSeen.After(originalNode.LastSeen))
		assert.WithinDuration(t, newTimestamp, updatedNode.LastSeen, time.Second)

	})

	t.Run("Create Uptime Report for Non-existent Node", func(t *testing.T) {
		report := &UptimeReport{
			NodeID:     999,
			Duration:   24 * time.Hour,
			Timestamp:  time.Now(),
			WasRestart: false,
		}

		err := testDB.CreateUptimeReport(report)
		require.Error(t, err)
	})
}

func TestSetZOSVersion(t *testing.T) {
	BeforeEach(t)

	t.Run("Set New ZOS Version", func(t *testing.T) {
		err := testDB.SetZOSVersion("4.1.0")
		require.NoError(t, err)

		version, err := testDB.GetZOSVersion()
		require.NoError(t, err)
		assert.Equal(t, "4.1.0", version)

		var versionGorm ZosVersion
		err = testDB.gormDB.Where("key = ?", ZOS4VersionKey).First(&versionGorm).Error
		require.NoError(t, err)
		assert.Equal(t, "4.1.0", versionGorm.Version)
	})

	t.Run("Update Existing ZOS Version", func(t *testing.T) {
		err := testDB.SetZOSVersion("4.2.0")
		require.NoError(t, err)

		version, err := testDB.GetZOSVersion()
		require.NoError(t, err)
		assert.Equal(t, "4.2.0", version)
	})

	t.Run("Set Same Version Again", func(t *testing.T) {
		currentVersion, err := testDB.GetZOSVersion()
		require.NoError(t, err)

		err = testDB.SetZOSVersion(currentVersion)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "version already set")
	})
}

func TestGetZOSVersion(t *testing.T) {
	BeforeEach(t)

	t.Run("Get Existing ZOS Version", func(t *testing.T) {
		version, err := testDB.GetZOSVersion()
		require.NoError(t, err)
		assert.Equal(t, "4.0.0", version)
	})
}
