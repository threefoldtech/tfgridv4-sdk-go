# ThreeFold Grid Node Registrar Client

A Go client for interacting with the ThreeFold Grid Node Registrar service. Facilitates node registration and management on the ThreeFold Grid.

## Overview

The Node Registrar Client enables communication with the ThreeFold Grid's node registration service. It provides methods to:

* Register nodes
* Manage node metadata
* Retrieve node information
* Delete node registrations

## Features

### Version

* **Get Zos Version**: Loads zos version for the current network
* **Set Zos Version**: Set zos version to specific version (can only be done by the network admin)

### Accounts

* **Create Account**: Create new account on the registrar with uniqe key.
* **Update Account**: Update the account configuration (relays & rmbEncKey).
* **Ensure Account**: Ensures that an account is created with specific seed.
* **Get Account**: Get an account using either its twin\_id or its public\_key.

### Farms

* **Create Farm**: Create new farm on the registrar with uniqe name.
* **update Farm**: update farm configuration (farm\_id, dedicated).
* **Get Farm**: Get a farm using either its farm\_id.

### Node

* **Register Node**: Register physical/virtual nodes with the TFGrid.
* **Update Node**: Update node configuration (farm\_id, interfaces, resources, location, secure\_boot, virtualized).
* **Get Node**: Fetch registered node details using (node\_id, twin\_id, farm\_id).
* **Update Node Uptime**: Update node Uptime.

### API Methods

#### Version Operations

| Method          | Description                      | Parameters                 | Returns             |
|-----------------|----------------------------------|----------------------------|---------------------|
| GetZosVersion   | Get current zos version          | None                       | (ZosVersion, error) |
| SetZosVersion   | Update zos version (admin-only)  | version string, force bool | error               |

#### Account Management

| Method         | Description               | Parameters                        | Returns          |
|----------------|---------------------------|-----------------------------------|------------------|
| CreateAccount  | Create new account        | relays []string, rmbEncKey string | (Account, error) |
| EnsureAccount  | Create account if missing | relays []string, rmbEncKey string | (Account, error) |
| GetAccount     | Get account by twin ID    | twinID uint64                     | (Account, error) |
| GetAccountByPK | Get account by public key | publicKey string                  | (Account, error) |
| UpdateAccount  | Modify account config     | ...UpdateOption                   | error            |

#### Farm Operations

| Method      | Description              | Parameters                               | Returns         |
|-------------|------------------------|--------------------------------------------|-----------------|
| CreateFarm  | Register new farm      | name string, twinID uint64, dedicated bool | (uint64, error) |
| UpdateFarm  | Modify farm properties | farmID uint64, ...UpdateOption             | error           |
| GetFarm     | Get farm by ID         | farmID uint64                              | (Farm, error)   |
| ListFarms   | List farms             | ...ListOption                              | ([]Farm, error) |

#### Node Operations

| Method          | Description           | Parameters                                     | Returns         |
|-----------------|-----------------------|------------------------------------------------|-----------------|
| RegisterNode    | Register new node     | farmID uint64, twinID uint64, interfaces []Interface, location Location, resources Resources, serial string, secureBoot bool, virtualized bool | (uint64, error) |
| UpdateNode      | Modify node config    | ...UpdateOption                                | error           |
| GetNode         | Get node by node ID   | nodeID uint64                                  | (Node, error)   |
| GetNodeByTwinID | Get node by twin ID   | twinID uint64                                  | (Node, error)   |
| ListNodes       | List nodes            | ...ListOption                                  | ([]Node, error) |
| ReportUptime    | Submit uptime metrics | report UptimeReport                            | error           |

## Installation

```bash
go get github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client
```

## Usage

### Initialize Client

```go
import (
  "github.com/threefoldtech/tfgrid4-sdk-go/node-registrar/client"
)

func main() {
  registrarURL := "https://registrar.dev4.grid.tf/v1"

  s := make([]byte, 32)
  _, err := rand.Read(s)
  if err != nil {
    log.Fatal().Err(err).Send()
  }
  seed = hex.EncodeToString(s)
  fmt.Println("New Seed (Hex):", seed)
  
  cli, err := client.NewRegistrarClient(registrarURL, s)
  if err != nil {
    panic(err)
  }
}
```

### Get Zos Version

```go
 version, err := c.GetZosVersion()
 if err != nil {
   log.Fatal().Err(err).Msg("failed to set registrar version")
 }

 log.Info().Msgf("%s version is: %+v", network, version)
```

### Set Zos Version (ONLY for network admin)

```go
 err := c.SetZosVersion("v0.1.8", true)
 if err != nil {
   log.Fatal().Err(err).Msg("failed to set registrar version")
 }

 log.Info().Msg("version is updated successfully")
```

### Create Account

```go
 account, err := c.CreateAccount(relays, rmbEncKey)
 if err != nil {
   log.Fatal().Err(err).Msg("failed to create new account on registrar")
 }

log.Info().Uint64("twinID", account.TwinID).Msg("account created successfully")

```

### Get Account

#### Get Account By Public Key

```go
 account, err := c.GetAccountByPK(pk)
 if err != nil {
   log.Fatal().Err(err).Msg("failed to get account from registrar")
 }
 log.Info().Any("account", account).Send()
```

#### Get Account By Twin ID

```go
 account, err := c.GetAccount(id)
 if err != nil {
   log.Fatal().Err(err).Msg("failed to get account from registrar")
 }
 log.Info().Any("account", account).Send()

```

### Update Account

```go
 err := c.UpdateAccount(client.UpdateAccountWithRelays(relays), client.UpdateAccountWithRMBEncKey(rmbEncKey))
 if err != nil {
   log.Fatal().Err(err).Msg("failed to get account from registrar")
 }
 log.Info().Msg("account updated successfully")
```

### Ensure Account

```go
 account, err := c.EnsureAccount(relays, rmbEncKey)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to ensure account account from registrar")
 }
 log.Info().Any("account", account).Send()
```

### Create Farm

```go
 id, err := c.CreateFarm(farmName, twinID, false)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to create new farm on registrar")
 }

 log.Info().Uint64("farmID", id).Msg("farm created successfully")
```

### Get Farm

```go
 farm, err := c.GetFarm(id)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to get farm from registrar")
 }
 log.Info().Any("farm", farm).Send()
```

### List Farms

```go
 farms, err := c.ListFarms(ListFarmWithName(name))
 if err != nil {
  log.Fatal().Err(err).Msg("failed to list farms from registrar")
 }
 log.Info().Any("farm", farms[0]).Send()
```

### Update Farm

```go
 err := c.UpdateFarm(farmID, client.UpdateFarmWithName(name))
 if err != nil {
  log.Fatal().Err(err).Msg("failed to get farm from registrar")
 }
 log.Info().Msg("farm updated successfully")
```

### Register a Node

```go
 id, err := c.RegisterNode(farmID, twinID, interfaces, location, resources, serialNumber, secureBoot, virtualized)
 if err != nil {
   log.Fatal().Err(err).Msg("failed to register a new node on registrar")
 }
 log.Info().Uint64("nodeID", id).Msg("node registered successfully")
```

### Get Node

#### Get Node With Node ID

```go
 node, err := c.GetNode(id)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to get node from registrar")
 }
 log.Info().Any("node", node).Send()
```

#### Get Node With Twin ID

```go
 node, err := c.GetNodeByTwinID(id)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to get node from registrar")
 }
 log.Info().Any("node", node).Send()
```

### List Nodes

```go
 nodes, err := c.ListNodes(client.ListNodesWithFarmID(id))
 if err != nil {
  log.Fatal().Err(err).Msg("failed to list nodes from registrar")
 }
 log.Info().Any("node", node[0]).Send()
```

### Update Node

```go
 err := c.UpdateNode(client.UpdateNodesWithLocation(location))
 if err != nil {
  log.Fatal().Err(err).Msg("failed to update node location on the registrar")
 }
 log.Info().Msg("node updated successfully")
```

### Update Node Uptime

```go
 err := c.ReportUptime(report)
 if err != nil {
  log.Fatal().Err(err).Msg("failed to update node uptime in the registrar")
 }
 log.Info().Msg("node uptime is updated successfully")
```
