# New Farm Tool

## Overview

This tool allows users to create a farm account on a specified network.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/threefoldtech/tfgrid4-sdk-go
   ```

2. Navigate to the project directory:

   ```sh
   cd tfgrid4-sdk-go/node-registrar/tools/farm
   ```

3. Build the application:

   ```sh
   go build -o new-farm main.go
   ```

## Usage

```sh
./new-farm -seed <seed> -network <network> -farm-name <farm-name>
```

### Parameters

- `-seed`  (required): A hexadecimal string used as a private key seed.
- `-network` (required): Specifies the network (`dev`, `qa`, `test`, `main`).
- `-farm-name` (required): The name of the farm to create.

### Example Usage

```sh
./new-farm -seed aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899 -network dev -farm-name MyFarm
```

The `FarmID` (e.g., `11`) is returned upon successful farm creation.

## Next Step

Once this is done, you can create a V4 bootstrap image using the V4 Zero-OS Boot Generator: <https://v4.bootstrap.grid.tf/>.
