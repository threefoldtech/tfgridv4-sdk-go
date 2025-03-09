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
   cd node-registrar/tools/registrar-cli
   ```

3. Build the application:

   ```sh
   go build -o registrar-cli main.go
   ```

## Usage

```sh
./registrar-cli -seed <seed> -network <network> -farm_name <farm_name>
```

### Parameters

- `-seed`  (required): A hexadecimal string used as a private key seed.
- `-network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `-farm_name` (required): The name of the farm to create.

### Example Usage

```sh
./registrar-cli -seed aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899 -network dev -farm_name MyFarm
```
