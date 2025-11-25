# Pulumi Turso Provider

A native Pulumi provider for managing [Turso](https://turso.tech/) databases and groups. Turso is a distributed SQLite database platform built on libSQL.

## Features

This provider allows you to manage the following Turso resources:

- **Group** - Database groups with primary and replica locations
- **Database** - SQLite databases within groups
- **GroupToken** - Authentication tokens for database groups
- **DatabaseToken** - Authentication tokens for individual databases

## Installation

### Go SDK

```bash
go get github.com/celest-dev/pulumi-turso/sdk/go/turso
```

### Node.js SDK

```bash
npm install @celest-dev/pulumi-turso
# or
yarn add @celest-dev/pulumi-turso
```

## Configuration

The provider requires the following configuration:

| Name | Description | Environment Variable |
|------|-------------|---------------------|
| `organization` | Your Turso organization name (required) | `TURSO_ORGANIZATION` |
| `apiToken` | Your Turso API token (optional if using Turso CLI auth) | `TURSO_API_TOKEN` |

### Setting Configuration

You can configure the provider using environment variables:

```bash
export TURSO_ORGANIZATION="your-org"
export TURSO_API_TOKEN="your-api-token"
```

Or via Pulumi configuration:

```bash
pulumi config set turso:organization your-org
pulumi config set turso:apiToken your-api-token --secret
```

### Authentication

The provider supports two authentication methods:

1. **API Token**: Set the `TURSO_API_TOKEN` environment variable or configure `apiToken`
2. **Turso CLI**: If no API token is provided, the provider will use your Turso CLI authentication (run `turso auth login` first)

## Resources

### Group

Creates a database group in a specific location with optional replicas.

#### Properties

| Input | Type | Description |
|-------|------|-------------|
| `name` | string | Name of the group |
| `primaryLocation` | string | Primary location code (e.g., "ord", "fra", "syd") |
| `replicaLocations` | string[] | Optional list of replica location codes |
| `extensions` | string | Optional SQLite extensions ("all" or "none") |

| Output | Type | Description |
|--------|------|-------------|
| `uuid` | string | Unique identifier of the group |
| `locations` | string[] | All locations where the group exists |
| `primary` | string | Primary location of the group |

### Database

Creates a database within a group.

#### Properties

| Input | Type | Description |
|-------|------|-------------|
| `name` | string | Name of the database |
| `group` | string | Name of the group to create the database in |
| `sizeLimit` | string | Optional size limit (e.g., "500mb") |
| `blockReads` | bool | Optional flag to block read operations |
| `blockWrites` | bool | Optional flag to block write operations |
| `seed` | object | Optional seed configuration for database initialization |

| Output | Type | Description |
|--------|------|-------------|
| `dbId` | string | Unique database identifier |
| `hostname` | string | Database hostname for connections |
| `instances` | array | List of database instances |

#### Database Seeding

You can seed a new database from an existing database or a dump file:

```yaml
seed:
  type: "database"        # or "dump"
  name: "source-db-name"  # for type: database
  url: "https://..."      # for type: dump
  timestamp: "2024-01-01T00:00:00Z"  # optional point-in-time recovery
```

### GroupToken

Creates an authentication token for a database group.

#### Properties

| Input | Type | Description |
|-------|------|-------------|
| `group` | string | Name of the group |
| `expiration` | string | Optional expiration (e.g., "2w", "30d", "never") |
| `authorization` | string | Authorization level: "full-access" or "read-only" |

| Output | Type | Description |
|--------|------|-------------|
| `token` | string (secret) | The generated JWT token |
| `expiresAt` | string | Token expiration timestamp |

### DatabaseToken

Creates an authentication token for a specific database.

#### Properties

| Input | Type | Description |
|-------|------|-------------|
| `database` | string | Name of the database |
| `expiration` | string | Optional expiration (e.g., "2w", "30d", "never") |
| `authorization` | string | Authorization level: "full-access" or "read-only" |

| Output | Type | Description |
|--------|------|-------------|
| `token` | string (secret) | The generated JWT token |
| `expiresAt` | string | Token expiration timestamp |

## Examples

### Go

```go
package main

import (
	"github.com/celest-dev/pulumi-turso/sdk/go/turso"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a database group
		group, err := turso.NewGroup(ctx, "my-group", &turso.GroupArgs{
			Name:            pulumi.String("my-group"),
			PrimaryLocation: pulumi.String("ord"),
		})
		if err != nil {
			return err
		}

		// Create a database in the group
		db, err := turso.NewDatabase(ctx, "my-db", &turso.DatabaseArgs{
			Name:  pulumi.String("my-db"),
			Group: group.Name,
		})
		if err != nil {
			return err
		}

		// Create a read-only token for the database
		token, err := turso.NewDatabaseToken(ctx, "my-db-token", &turso.DatabaseTokenArgs{
			Database:      db.Name,
			Expiration:    pulumi.String("2w"),
			Authorization: turso.DatabaseTokenAuthorization_Read_Only,
		})
		if err != nil {
			return err
		}

		// Export outputs
		ctx.Export("hostname", db.Hostname)
		ctx.Export("token", token.Token)

		return nil
	})
}
```

### Node.js / TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as turso from "@celest-dev/pulumi-turso";

// Create a database group
const group = new turso.Group("my-group", {
    name: "my-group",
    primaryLocation: "ord",
});

// Create a database in the group
const db = new turso.Database("my-db", {
    name: "my-db",
    group: group.name,
});

// Create a read-only token for the database
const token = new turso.DatabaseToken("my-db-token", {
    database: db.name,
    expiration: "2w",
    authorization: "read-only",
});

// Export outputs
export const hostname = db.hostname;
export const dbToken = token.token;
```

### YAML

```yaml
name: turso-example
runtime: yaml
resources:
  my-group:
    type: turso:Group
    properties:
      name: my-group
      primaryLocation: ord
  my-db:
    type: turso:Database
    properties:
      name: my-db
      group: ${my-group.name}
  my-db-token:
    type: turso:DatabaseToken
    properties:
      database: ${my-db.name}
      expiration: 2w
      authorization: read-only
outputs:
  hostname: ${my-db.hostname}
  token: ${my-db-token.token}
```

## Development

### Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)
- [Node.js 20+](https://nodejs.org/) (for Node.js SDK)

### Building

```bash
# Build the provider and SDKs
make build

# Install the provider locally
make install

# Run tests
make test
```

### Running Examples

```bash
# Run the YAML example
cd examples/yaml
pulumi up

# Run the Go example
cd examples/go
go mod tidy
pulumi up
```

## License

Apache 2.0 - See [LICENSE](./LICENSE) for details.
