package main

import (
	"crypto/ed25519"
	"os"
	"time"

	subkeyEd25519 "github.com/vedhavyas/go-subkey/v2/ed25519"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
	"github.com/vedhavyas/go-subkey/v2"
)

var (
	twinID   uint64 = 4
	farmID   uint64 = 1
	nodeID   uint64 = 2
	farmName        = "freeFarm"
	net             = ""
)

const (
	localHexKey = "acoustic foot tomorrow brown candy cash reject hurt wood roof blossom sausage"
	dev4HexKey  = ""
	qa4HexKey   = ""
	test4HexKey = ""
	prod4hexKey = ""
)

type network struct {
	url string
	key string
}

var networks = map[string]network{
	"local": {url: "http://localhost:8080/api/v1", key: localHexKey},
	// "dev":   {url: "https://registrar.dev4.grid.tf/api/v1", key: dev4HexKey},
	// "qa":    {url: "https://registrar.qa4.grid.tf/api/v1", key: qa4HexKey},
	// "test":  {url: "https://registrar.test4.grid.tf/api/v1", key: test4HexKey},
	// "prod":  {url: "https://registrar.prod4.grid.tf/api/v1", key: prod4hexKey},
}

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	if _, ok := networks[net]; !ok {
		net = "local"
	}

	registrarURL := networks[net].url
	// seedOrMnemonic := networks[net].key

	// publicKey, err := parseSeed(seedOrMnemonic)
	// if err != nil {
	// 	log.Fatal().Err(err).Send()
	// }

	c, err := client.NewRegistrarClient(registrarURL, "tool mirror high clay quit cube affair dirt rely hire joy text")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new registrar client")
	}

	// manageAccount(&c, publicKey)

	// manageFarm(&c)

	manageNode(&c)

	// manageVersion(&c)
}

// The flow:
// - we need to create an account
// - try to get the account by id and by pk
// - try to update the account
// - try to ensure account
func manageAccount(c *client.RegistrarClient, publicKey ed25519.PublicKey) {
	relays := []string{}
	rmbEncKey := ""

	log.Info().Msg("***************************************************************************")
	log.Info().Msg("manage account create/update/get/ensure")

	createAccount(c, relays, rmbEncKey)
	getAccountByPK(c, publicKey)
	relays = []string{"relay1", "relay2"}
	rmbEncKey = "key1&2"
	// updateAccount(c, relays, rmbEncKey)
	enusreAccount(c, relays, rmbEncKey)
	getAccount(c, twinID)
}

// The flow:
// - we need to create a farm with the created twin
// - try to get the farm
// - try to update the farm
func manageFarm(c *client.RegistrarClient) {
	log.Info().Msg("***************************************************************************")
	log.Info().Msg("manage farm create/update/get")
	// createFarm(c, farmName)
	// getFarm(c, farmID)
	// updateFarm(c, "notFreeFarm101")
	// getFarm(c, farmID)
}

// The flow:
// - we need to create an account and use it to register a node in the created farm
// - try to update the node
// - try to get the node
// - try to send the node up time report
// - try to the node by twin id
func manageNode(c *client.RegistrarClient) {
	log.Info().Msg("***************************************************************************")
	log.Info().Msg("manage node register/update/get/list/send uptime report")
	// Try with a nil IPs field to see if that helps
	interface1 := client.Interface{
		Name: "zos",
		Mac:  "mac",
		IPs:  []string{"1.1.1.1"},

		// Not explicitly setting IPs field at all
	}

	registerNode(c, farmID, twinID, []client.Interface{interface1}, client.Location{City: "somewhere"}, client.Resources{CRU: 8, SRU: 5497558138880,
		MRU: 34359738368,
		HRU: 17592186044416}, "serialNumber", false, false)
	node := getNode(c, nodeID)
	// updateNode(c, client.Location{City: "somewhere"})
	sendUptimeReport(c, client.UptimeReport{Uptime: 40 * 60, Timestamp: time.Now().Unix()})
	getNodeWithTwinID(c, node.TwinID)
}

func manageVersion(c *client.RegistrarClient) {
	log.Info().Msg("***************************************************************************")
	log.Info().Msg("manage version set/get")
	err := c.SetZosVersion("v0.1.8", true)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set registrar version")
	}

	version, err := c.GetZosVersion()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set registrar version")
	}

	log.Info().Msg("version is updated successfully")
	log.Info().Msgf("%s version is: %+v", net, version)
}

func createAccount(c *client.RegistrarClient, relays []string, rmbEncKey string) {
	log.Info().Msg("create account")
	account, mnemonic, err := c.CreateAccount(relays, rmbEncKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new account on registrar")
	}

	log.Info().Uint64("twinID", account.TwinID).Str("mnemonic", mnemonic).Msg("account created successfully")
	twinID = account.TwinID
}

func getAccountByPK(c *client.RegistrarClient, pk []byte) {
	log.Info().Msg("get account by public key")
	account, err := c.GetAccountByPK(pk)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get account from registrar")
	}
	log.Info().Any("account", account).Send()
}

func getAccount(c *client.RegistrarClient, id uint64) {
	log.Info().Msg("get account by twinID")
	account, err := c.GetAccount(id)
	log.Info().Uint64("twinID", id).Send()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get account from registrar")
	}
	log.Info().Any("account", account).Send()
}

func updateAccount(c *client.RegistrarClient, relays []string, rmbEncKey string) {
	log.Info().Msg("update account")
	err := c.UpdateAccount(client.UpdateAccountWithRelays(relays), client.UpdateAccountWithRMBEncKey(rmbEncKey))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get account from registrar")
	}
	log.Info().Msg("account updated successfully")
}

func enusreAccount(c *client.RegistrarClient, relays []string, rmbEncKey string) {
	log.Info().Msg("ensure account")
	account, err := c.EnsureAccount(relays, rmbEncKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ensure account account from registrar")
	}
	log.Info().Any("account", account).Send()
}

func createFarm(c *client.RegistrarClient, farmName string) {
	log.Info().Msg("create farm")
	id, err := c.CreateFarm(farmName, "GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGJJJJJJGGGGGGGGGG", false)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create new farm on registrar")
	}

	log.Info().Uint64("farmID", id).Msg("farm created successfully")
	farmID = id
}

func getFarm(c *client.RegistrarClient, id uint64) {
	log.Info().Msg("get farm by id")
	farm, err := c.GetFarm(id)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get farm from registrar")
	}
	log.Info().Any("farm", farm).Send()
}

func updateFarm(c *client.RegistrarClient, name string) {
	log.Info().Msg("update farm")
	addr := "GGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGJJJJJJGGGGGGGGGG"
	err := c.UpdateFarm(2, client.FarmUpdate{FarmName: &name, StellarAddress: &addr})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to update farm from registrar")
	}
	log.Info().Msg("farm updated successfully")
}

func registerNode(c *client.RegistrarClient, farmID uint64, twinID uint64, interfaces []client.Interface, location client.Location, resources client.Resources, serialNumber string, secureBoot, virtualized bool) {
	log.Info().Msg("register node")
	id, err := c.RegisterNode(client.Node{
		FarmID:       farmID,
		TwinID:       twinID,
		Interfaces:   interfaces,
		Location:     location,
		Resources:    resources,
		SerialNumber: serialNumber,
		SecureBoot:   secureBoot,
		Virtualized:  virtualized,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register a new node on registrar")
	}

	log.Info().Uint64("nodeID", id).Msg("node registered successfully")
	nodeID = id
}

func getNode(c *client.RegistrarClient, id uint64) (node client.Node) {
	log.Info().Msg("get node with node id")
	node, err := c.GetNode(id)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get node from registrar")
	}
	log.Info().Any("node", node).Send()
	return node
}

func getNodeWithTwinID(c *client.RegistrarClient, id uint64) {
	log.Info().Msg("get node with twin id")
	node, err := c.GetNodeByTwinID(id)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get node from registrar")
	}
	log.Info().Any("node", node).Send()
}

func updateNode(c *client.RegistrarClient, location client.Location) {
	log.Info().Msg("update node, update node location")
	err := c.UpdateNode(client.NodeUpdate{Resources: &client.Resources{CRU: 8, MRU: 68719476736, SRU: 4398046511104, HRU: 17592186044416}})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to update node location on the registrar")
	}
	log.Info().Msg("node updated successfully")
}

func sendUptimeReport(c *client.RegistrarClient, report client.UptimeReport) {
	log.Info().Msg("send uptime report")
	err := c.ReportUptime(report)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to update node uptime in the registrar")
	}
	log.Info().Msg("node uptime is updated successfully")
}

func parseSeed(mnemonicOrSeed string) (publicKey ed25519.PublicKey, err error) {
	keypair, err := subkey.DeriveKeyPair(subkeyEd25519.Scheme{}, mnemonicOrSeed)
	if err != nil {
		return publicKey, errors.Wrapf(err, "Failed to derive key pair from seed %s", mnemonicOrSeed)
	}

	return keypair.Public(), err
}
