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

## Account Command

Account command represents events on account on Threefold grid4

### Create New Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--mnemonic`  (optional):  create an account of a mnemonic.
- `--relays` (optional): relays urls.
- `--rmb-enc-key` (optional): rmb encryption key.

**Example Usage**:

```sh
➜ ./registrar-cli account create --network local
1:07PM INF new account is created with mnemonic mnemonic="pyramid cattle mutual brush green east slam lava source stereo rigid able"
1:07PM INF account is created successfully twinID=4
```

### Get Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--twin-id` (optional): twin id of the account needed to be loaded.
- `--public-key` (optional): public key of the account needed to be loaded.

**Example Usage**:

```sh
➜  ./registrar-cli account get --network dev --twin-id 33
5:00PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":[],"rmb_enc_key":"","twin_id":33}
➜
➜  ./registrar-cli account get --network dev --public-key <public-key>
5:01PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":[],"rmb_enc_key":"","twin_id":33}
```

### Update Account

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--mnemonic`  (required): Account mnemonic.
- `--relays` (optional): new relays urls.
- `--rmb-enc-key` (optional): new rmb encryption key.

**Example Usage**:

```sh
➜  ./registrar-cli account update --network dev --mnemonic <mnemonic> --relays wss://relay.dev.grid.tf
5:02PM INF account is updated successfully
➜
➜  ./registrar-cli account get --network dev --twin-id 33
5:02PM INF account={"public_key":"w5TYTeB/rCskd1iNrOKaFlpGn+Cp28gFZobTNABUv50=","relays":["wss://relay.dev.grid.tf"],"rmb_enc_key":"","twin_id":33}
```

## Farm Command

Farm command represents events on farms on Threefold grid4

### Create New Farm

**Flags**:

- `--farm-name` (required): The name of the farm to create.
- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--mnemonic`  (required): Account mnemonic.
- `--dedicated` (optional default: false): is the farm dedicated.

**Example Usage**:

```sh
➜  ./registrar-cli farm create --farm-name testFarm1 --mnemonic <mnemonic> --network dev
5:03PM INF farm is created successfully farmID=12
```

### Get Farm

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--farm-id` (optional): id of the farm needed to be loaded.

**Example Usage**:

```sh
➜  ./registrar-cli farm get --farm-id 12 --network dev
5:03PM INF farm={"dedicated":false,"farm_id":12,"farm_name":"testFarm1","twin_id":33}
```

### Update Farm

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--mnemonic`  (required): Account mnemonic.
- `--farm-id` (optional): id of the farm needed to be loaded.
- `--farm-name` (optional): The new name of the farm.
- `--dedicated` (optional): update the farm to dedicated.

**Example Usage**:

```sh
➜  ./registrar-cli farm update --farm-name notTestFarm1 --mnemonic <mnemonic> --network dev --farm-id 12
5:04PM INF farm is updated successfully
➜
➜  ./registrar-cli farm get --farm-id 12 --network dev
5:04PM INF farm={"dedicated":false,"farm_id":12,"farm_name":"notTestFarm1","twin_id":33}
```

## Node Command

Node command represents events on nodes on Threefold grid4

### Get Node

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--node-id` (optional): id of the node needed to be loaded.
- `--twin-id` (optional): twin id of the node needed to be loaded.

**Example Usage**:

```sh
➜  ./registrar-cli node get --network dev --node-id 1
12:36PM INF node={"Approved":false,"farm_id":4,"interfaces":[{"ips":"192.168.123.22","mac":"54:fe:9a:b0:73:61","name":"zos"}],"location":{"city":"Cairo","country":"Egypt","latitude":"30.0588","longitude":"31.2268"},"node_id":1,"resources":{"cru":4,"hru":1073741824000,"mru":6230032384,"sru":1610612736000},"secure_boot":false,"serial_number":"Not Specified","twin_id":5,"uptime":null,"virtualized":true}
➜
➜  ./registrar-cli node get --network dev --twin-id 5
12:36PM INF node={"Approved":false,"farm_id":4,"interfaces":[{"ips":"192.168.123.22","mac":"54:fe:9a:b0:73:61","name":"zos"}],"location":{"city":"Cairo","country":"Egypt","latitude":"30.0588","longitude":"31.2268"},"node_id":1,"resources":{"cru":4,"hru":1073741824000,"mru":6230032384,"sru":1610612736000},"secure_boot":false,"serial_number":"Not Specified","twin_id":5,"uptime":null,"virtualized":true}
```

## Zos Version Command

Zos Version command represents events on zos version on Threefold grid4

### Get Zos Version

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).

**Example Usage**:

```sh
➜  ./registrar-cli zos-version get --network dev
12:40PM INF zosVersion={"safe_to_upgrade":true,"version":"v0.1.8"}
```

### Update Zos Version

**Flags**:

- `--network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `--version` (required): new zos version to be set on specific network (`v0.1.x`)
- `--safe-to-upgrade` (required): if this version is safe to upgrade

**Example Usage**:

```sh
➜  ./registrar-cli zos-version update --network dev --version v0.1.8 --safe-to-upgrade --mnemonic <mnemonic>
5:07PM INF zos version is updated successfully
➜
➜  ./registrar-cli zos-version get --network dev
5:07PM INF zosVersion={"safe_to_upgrade":true,"version":"v0.1.8"}
```
