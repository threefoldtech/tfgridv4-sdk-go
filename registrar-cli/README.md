# New Farm Tool

## Overview

This tool allows users to create/get/update  farm/account/node on a specified network.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/threefoldtech/tfgrid4-sdk-go
   ```

2. Navigate to the project directory:

   ```sh
   cd tfgrid4-sdk-go/registrar-cli/
   ```

3. Build the application:

   ```sh
   go build -o registrar-cli main.go
   ```

## Usage

## Create Command

Create command allows users to create an account or a farm

### Create New Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--seed`  (optional):  create an account of a seed.
- `--relays` (optional): relays urls.
- `--rmb-enc-key` (optional): rmb encryption key.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ ./registrar-cli create account --network dev
5:00PM INF New Seed (Hex): 7f40eb52530f1a1c1253873cf17d44bd66d3e5ba71a14d0deba7df5517c9ed12
5:00PM INF public key (Hex): c394d84de07fac2b2477588dace29a165a469fe0a9dbc8056686d3340054bf9d
5:00PM INF account is created successfully twinID=33
```

### Create New Farm

**Flags**:

- `--farm_name` (required): The name of the farm to create.
- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--seed` (required): A hexadecimal  representation of the seed.
- `--dedicated` (optional default: false): is the farm dedicated.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ ./registrar-cli create farm --farm-name testFarm1 --seed <seed> --network dev
5:03PM INF farm is created successfully farmID=12
```

## Get Command

Get command allows users to get account, farm, node or zos version.

### Get Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--twin-id` (optional): twin id of the account needed to be loaded.
- `--public-key` (optional): public key of the account needed to be loaded.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get account --network dev --twin-id 33
5:00PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":[],"rmb_enc_key":"","twin_id":33}
➜
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get account --network dev --public-key <public-key> 
5:01PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":[],"rmb_enc_key":"","twin_id":33}
```

### Get Farm

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--farm-id` (optional): id of the farm needed to be loaded.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get farm --farm-id 12 --network dev
5:03PM INF farm={"dedicated":false,"farm_id":12,"farm_name":"testFarm1","twin_id":33}
```

### Get Node

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--node-id` (optional): id of the node needed to be loaded.
- `--twin-id` (optional): twin id of the node needed to be loaded.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get node --network dev --node-id 1
12:36PM INF node={"Approved":false,"farm_id":4,"interfaces":[{"ips":"192.168.123.22","mac":"54:fe:9a:b0:73:61","name":"zos"}],"location":{"city":"Cairo","country":"Egypt","latitude":"30.0588","longitude":"31.2268"},"node_id":1,"resources":{"cru":4,"hru":1073741824000,"mru":6230032384,"sru":1610612736000},"secure_boot":false,"serial_number":"Not Specified","twin_id":5,"uptime":null,"virtualized":true}
➜
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get node --network dev --twin-id 5
12:36PM INF node={"Approved":false,"farm_id":4,"interfaces":[{"ips":"192.168.123.22","mac":"54:fe:9a:b0:73:61","name":"zos"}],"location":{"city":"Cairo","country":"Egypt","latitude":"30.0588","longitude":"31.2268"},"node_id":1,"resources":{"cru":4,"hru":1073741824000,"mru":6230032384,"sru":1610612736000},"secure_boot":false,"serial_number":"Not Specified","twin_id":5,"uptime":null,"virtualized":true}
```

### Get Zos Version

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get version --network dev
12:40PM INF zosVersion={"safe_to_upgrade":true,"version":"v0.1.8"}
```

## Update Command

Update command allows users to update account, farm or zos version.

### Update Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--seed`  (required):  A hexadecimal  representation of the seed.
- `--relays` (optional): new relays urls.
- `--rmb-enc-key` (optional): new rmb encryption key.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go update account --network dev --seed <seed> --relays wss://relay.dev.grid.tf
5:02PM INF account is updated successfully
➜
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get account --network dev --twin-id 33
5:02PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":["wss://relay.dev.grid.tf"],"rmb_enc_key":"","twin_id":33}
```

### Update Farm

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--seed`  (required):  A hexadecimal  representation of the seed.
- `--farm-id` (optional): id of the farm needed to be loaded.
- `--farm_name` (optional): The new name of the farm.
- `--dedicated` (optional): update the farm to dedicated.

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go update farm --farm-name notTestFarm1 --seed <seed> --network dev --farm-id 12
5:04PM INF farm is updated successfully
➜
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get farm --farm-id 12 --network dev
5:04PM INF farm={"dedicated":false,"farm_id":12,"farm_name":"notTestFarm1","twin_id":33}
```

### Update Zos Version

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--version` (required): new zos version to be set on specific network (`v0.1.x`)
- `--safe-to-upgrade` (required): if this version is safe to upgrade

**Example Usage**:

```sh
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go update version --network dev --version v0.1.8 --safe-to-upgrade --seed <seed>
5:07PM INF farm is updated successfully
➜
➜  registrar-cli git:(add-registrar-cli-tool) ✗ go run main.go get version --network dev
5:07PM INF zosVersion={"safe_to_upgrade":true,"version":"v0.1.8"}
```
