
# Node Registrar Service

[![Go Report Card](https://goreportcard.com/badge/github.com/threefoldtech/tfgrid-sdk-go/node-registrar)](https://goreportcard.com/report/github.com/threefoldtech/tfgrid-sdk-go/node-registrar)
[![GoDoc](https://godoc.org/github.com/threefoldtech/tfgrid-sdk-go/node-registrar?status.svg)](https://godoc.org/github.com/threefoldtech/tfgrid-sdk-go/node-registrar)

## Overview

This project provides an API for registring and managing zos nodes on ThreeFold GridV4. Built with the Go Gin framework and PostgreSQL database,
It offers operations like registring, listing, and updating farms and nodes, as well as reporting uptime and consumption data for nodes.

## Features

- **Farm Management**
  - Create/update farms with owner authorization
  - List farms with filtering/pagination
  - Automatic twin ID association via authentication
  
- **Node Registration**
  - Create/update nodes with owner authorization
  - Uptime reporting
  - Node metadata management (location, interfaces, specs)
  
- **Account System**
  - ED25519/SR25519 authentication
  - Relay management for RMB communication

- **Security**
  - Challenge-response authentication middleware
  - Ownership verification for mutations
  - Timestamp replay protection

## Endpoint Descriptions

### Farms Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/farms/` | List all farms with optional filtering |
| GET | `/farms/:farm_id` | Get a specific farm by ID |
| POST | `/farms/` | Create a new farm |
| PATCH | `/farms/` | Update an existing farm |

### Nodes Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/nodes/` | List all nodes with optional filtering |
| GET | `/nodes/:node_id` | Get a specific node by ID |
| POST | `/nodes/` | Register a new node |
| POST | `/nodes/:node_id/uptime` | Report uptime for a specific node |
| POST | `/nodes/:node_id/consumption` | Report consumption for a specific node |

## Setup Instructions

1. **Start PostgreSQL:**

   ```bash
   make start-postgres
   ```

2. **Run the Server:**

   ```bash
   make run
   ```

3. **Stop PostgreSQL:**

   ```bash
   make stop-postgres
   ```

## Swagger Documentation

Once the server is running, Swagger documentation can be accessed at:

```bash
http://<domain>:<port>/swagger/index.html
```

Replace `<domain>` and `<port>` with the appropriate values.

## How to Use the Server

1. Use a tool like Postman or cURL to interact with the API.
2. Refer to the Swagger documentation for detailed information about request parameters and response structures.

## How to run the server with docker

1. To use the docker file to build the docker image, run this command in the root directory of the sdk

```bash
docker build -t myserver:latest -f node-registrar/Dockerfile .
```

2. run the image

   ```bash
   docker run -d \
   -p 8080:8080 \
   --name registrar \
   registrar:latest \
   ./server
   --postgres-host=<your-postgres-host> \
   --postgres-port=5432 \
   --postgres-db=<your-db-name> \
   --postgres-user=<your-db-user> \
   --postgres-password=<your-db-password> \
   --ssl-mode=disable \
   --sql-log-level=2 \
   --max-open-conn=10 \
   --max-idle-conn=5 \
   --server-port=8080 \
   --<domain=your-domain> \
   --network=main\
   --admin_twin_id=1
   --debug
   ```

## Authentication

Requests requiring authorization must include:

```http
X-Auth: Base64(Challenge):Base64(Signature)
```

**Challenge Format:**
`<unix_timestamp>:<twin_id>`

**Signature:**
ED25519/SR25519 signature of challenge bytes

The service uses a PostgreSQL database with the following key tables:

| Table | Description |
|-------|-------------|
| `accounts` | Authentication credentials and relay configurations |
| `farms` | Farm metadata with owner relationship |
| `nodes` | Node hardware specifications and resources |
| `uptime_reports` | Historical node availability data |

## Development

### Generating Swagger Docs

```bash
swag init -g pkg/server/handlers.go --output docs --parseDependency --parseDepth 2
```

## Client Library

A Go client library is available to interact with the Node Registrar Service. See the [client documentation](./client/README.md) for details on installation and usage.

Example usage:

```go
import "github.com/threefoldtech/tfgrid-sdk-go/node-registrar/client"

// Initialize client
cli, err := client.NewRegistrarClient("https://registrar.dev.grid.tf", mnemonic)

// Register a node
nodeID, err := cli.RegisterNode(farmID, twinID, interfaces, location, resources, serialNumber, secureBoot, virtualized)
```

### Generating Swagger Documentation

```bash
swag init -g pkg/server/handlers.go --output docs --parseDependency --parseDepth 2
```

### Running Tests

```bash
make test
```
