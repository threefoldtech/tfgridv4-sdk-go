# New Account Tool

## Overview
This tool is a command-line application for creating an account on a specified network on the tfgrid4 registrar using an Ed25519 key pair.
It generates a public-private key pair from a seed and sends a signed request to register an account.

## Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/threefoldtech/tfgrid4-sdk-go
   ```
2. Navigate to the project directory:
   ```sh
   cd node-registrar
   ```
3. Build the application:
   ```sh
   go build -o registrar-tool main.go
   ```

## Usage
Run the tool using the following command:
```sh
./registrar-tool -seed <seed> -network <network>
```

### Parameters
- `-seed` (optional): A 64-character hexadecimal string representing the private key seed. **If omitted**, a new random seed is generated.
- `-network` (required): Specifies the target network (`dev`, `qa`, `test`, `main`).

### Example Usage
#### 1. Using an Existing Seed
```sh
./registrar-tool -seed aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899 -network dev
```

#### 2. Generating a New Seed
```sh
./registrar-tool -network qa
```
Output:
```
New Seed (Hex): abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
12345678
```
The `TwinID` (e.g., `12345678`) is returned upon successful account creation.
