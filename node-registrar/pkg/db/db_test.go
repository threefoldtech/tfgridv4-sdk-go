package db

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/logger"
)

var testDB Database
var testData *TestFixtures

type TestFixtures struct {
	Accounts      []Account
	Farms         []Farm
	Nodes         []Node
	UptimeReports []UptimeReport
	ZosVersions   []ZosVersion
}

func TestMain(m *testing.M) {

	if err := setupTestDB(); err != nil {
		log.Printf("Failed to setup test database: %v", err)
		os.Exit(1)
	}

	testData = createTestFixtures()

	code := m.Run()

	if testDB.gormDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

func setupTestDB() error {
	testConfig := Config{
		PostgresHost:     "localhost",
		PostgresPort:     5432,
		DBName:           "test_noderegistrar",
		PostgresUser:     "postgres",
		PostgresPassword: "postgres",
		SSLMode:          "disable",
		SqlLogLevel:      logger.Silent,
		MaxOpenConns:     5,
		MaxIdleConns:     2,
	}

	db, err := NewDB(testConfig)
	if err != nil {
		log.Printf("Cannot connect to test database, skipping tests: %v", err)
		return nil
	}

	testDB = db
	return nil
}

func BeforeEach(t *testing.T) {
	if testDB.gormDB == nil {
		t.Skip("Test database not available")
	}

	cleanDatabase()

	loadTestData(t)
}

func cleanDatabase() {
	if testDB.gormDB == nil {
		return
	}

	testDB.gormDB.Exec("DELETE FROM uptime_reports")
	testDB.gormDB.Exec("DELETE FROM nodes")
	testDB.gormDB.Exec("DELETE FROM farms")
	testDB.gormDB.Exec("DELETE FROM accounts")
	testDB.gormDB.Exec("DELETE FROM zos_versions")

	// Reset sequences to avoid primary key conflicts
	testDB.gormDB.Exec("ALTER SEQUENCE accounts_twin_id_seq RESTART WITH 1")
	testDB.gormDB.Exec("ALTER SEQUENCE farms_farm_id_seq RESTART WITH 1")
	testDB.gormDB.Exec("ALTER SEQUENCE nodes_node_id_seq RESTART WITH 1")
	testDB.gormDB.Exec("ALTER SEQUENCE uptime_reports_id_seq RESTART WITH 1")
}

func loadTestData(t *testing.T) {

	for _, account := range testData.Accounts {
		err := testDB.gormDB.Create(&account).Error
		require.NoError(t, err)
	}

	for _, farm := range testData.Farms {
		err := testDB.gormDB.Create(&farm).Error
		require.NoError(t, err)
	}

	for _, node := range testData.Nodes {
		err := testDB.gormDB.Create(&node).Error
		require.NoError(t, err)
	}

	for _, report := range testData.UptimeReports {
		err := testDB.gormDB.Create(&report).Error
		require.NoError(t, err)
	}

	for _, version := range testData.ZosVersions {
		err := testDB.gormDB.Create(&version).Error
		require.NoError(t, err)
	}
}

func createTestFixtures() *TestFixtures {
	now := time.Now()
	thirtyMinsAgo := now.Add(-30 * time.Minute)
	oneHourAgo := now.Add(-1 * time.Hour)

	return &TestFixtures{
		Accounts: []Account{
			{
				PublicKey: "test_public_key_1",
				Relays:    pq.StringArray{"relay1.example.com", "relay2.example.com"},
				RMBEncKey: "test_rmb_key_1",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				PublicKey: "test_public_key_2",
				Relays:    pq.StringArray{"relay3.example.com"},
				RMBEncKey: "test_rmb_key_2",
				CreatedAt: now,
				UpdatedAt: now,
			},
			{
				PublicKey: "test_public_key_3",
				Relays:    pq.StringArray{},
				RMBEncKey: "test_rmb_key_3",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		Farms: []Farm{
			{
				FarmName:       "TestFarm1",
				TwinID:         1,
				StellarAddress: "G" + strings.Repeat("A", 55),
				Dedicated:      false,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			{
				FarmName:       "TestFarm2",
				TwinID:         2,
				StellarAddress: "G" + strings.Repeat("B", 55),
				Dedicated:      true,
				CreatedAt:      now,
				UpdatedAt:      now,
			},
		},
		Nodes: []Node{
			{
				FarmID: 1,
				TwinID: 1,
				Location: Location{
					Country:   "US",
					City:      "New York",
					Longitude: "-74.0060",
					Latitude:  "40.7128",
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
						Mac:  "00:11:22:33:44:55",
						IPs:  []string{"192.168.1.100", "2001:db8::1"},
					},
				},
				SecureBoot:   true,
				Virtualized:  false,
				SerialNumber: "SN001",
				LastSeen:     thirtyMinsAgo,
				CreatedAt:    now,
				UpdatedAt:    now,
				Approved:     true,
			},
			{
				FarmID: 1,
				TwinID: 2,
				Location: Location{
					Country:   "DE",
					City:      "Berlin",
					Longitude: "13.4050",
					Latitude:  "52.5200",
				},
				Resources: Resources{
					HRU: 2000,
					SRU: 1000,
					CRU: 16,
					MRU: 32,
				},
				Interfaces: []Interface{
					{
						Name: "eth0",
						Mac:  "00:11:22:33:44:66",
						IPs:  []string{"192.168.1.101"},
					},
				},
				SecureBoot:   false,
				Virtualized:  true,
				SerialNumber: "SN002",
				LastSeen:     oneHourAgo,
				CreatedAt:    now,
				UpdatedAt:    now,
				Approved:     false,
			},
			{
				FarmID: 2,
				TwinID: 3,
				Location: Location{
					Country:   "JP",
					City:      "Tokyo",
					Longitude: "139.6917",
					Latitude:  "35.6895",
				},
				Resources: Resources{
					HRU: 3000,
					SRU: 1500,
					CRU: 32,
					MRU: 64,
				},
				Interfaces: []Interface{
					{
						Name: "eth0",
						Mac:  "00:11:22:33:44:77",
						IPs:  []string{"192.168.1.102", "192.168.1.103"},
					},
					{
						Name: "eth1",
						Mac:  "00:11:22:33:44:88",
						IPs:  []string{"10.0.0.1"},
					},
				},
				SecureBoot:   true,
				Virtualized:  false,
				SerialNumber: "SN003",
				LastSeen:     now,
				CreatedAt:    now,
				UpdatedAt:    now,
				Approved:     true,
			},
		},
		UptimeReports: []UptimeReport{
			{
				NodeID:     1,
				Duration:   24 * time.Hour,
				Timestamp:  thirtyMinsAgo,
				WasRestart: false,
				CreatedAt:  thirtyMinsAgo,
			},
			{
				NodeID:     1,
				Duration:   48 * time.Hour,
				Timestamp:  oneHourAgo,
				WasRestart: true,
				CreatedAt:  oneHourAgo,
			},
			{
				NodeID:     3,
				Duration:   12 * time.Hour,
				Timestamp:  now,
				WasRestart: false,
				CreatedAt:  now,
			},
		},
		ZosVersions: []ZosVersion{
			{
				Key:     ZOS4VersionKey,
				Version: "4.0.0",
			},
			{
				Key:     "zos_3",
				Version: "3.14.1",
			},
		},
	}
}
