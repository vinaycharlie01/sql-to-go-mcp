# sql-repository-mcp

A production-grade **Model Context Protocol (MCP) server** that converts SQL `CREATE TABLE` schemas into complete, compilable Go repository code following **Hexagonal / Clean Architecture**.

Works as a local MCP tool with **GitHub Copilot**, **Claude**, and **ChatGPT**.

---

## What It Generates

Give it a SQL schema and it produces:

| Artifact | Description |
|---|---|
| **Domain entity** | Typed Go struct with GORM + JSON tags |
| **Repository interface** | Context-aware CRUD + List contract |
| **GORM implementation** | Full CRUD with slog instrumentation |
| **sqlc config + queries** | `sqlc.yaml` + annotated `.sql` file |
| **sqlc adapter** | Repository adapter over generated sqlc code |
| **Unit tests** | Table-driven tests with testify mocks |
| **Integration tests** | Tests against a real database |
| **Testify mocks** | Drop-in mocks for your test doubles |
| **Benchmarks** | `BenchmarkCreate/GetByID/Update/Delete/List` |
| **Migration** | Up/down SQL migration file |

---

## MCP Tools

| Tool | Purpose |
|---|---|
| `generate_repository` | Full artifact set from a `CREATE TABLE` statement |
| `sql_to_gorm` | GORM model + repository + tests |
| `sql_to_sqlc` | `sqlc.yaml` + `queries.sql` + adapter + tests |
| `benchmark_query` | Anti-pattern detection + Go benchmark stub |
| `generate_tests` | Unit tests, integration tests, testify mocks |

---

## Prerequisites

- **Go 1.22+** — [install](https://go.dev/dl/)
- **Git**

---

## Build

```bash
git clone https://github.com/vinaycharlie01/sql-to-go-mcp.git
cd sql-to-go-mcp
go build -o bin/sql-repository-mcp ./cmd/server
```

The binary at `bin/sql-repository-mcp` is the MCP server. It communicates over **stdio** (standard input/output), which is the transport all three clients below expect.

---

## Setup — GitHub Copilot (VS Code)

GitHub Copilot in VS Code supports MCP servers via a workspace or user-level config file.

### 1. Install the extension

Make sure you have the **GitHub Copilot Chat** extension (`GitHub.copilot-chat`) version **≥ 0.22** installed.

### 2. Create the MCP config

**Workspace-level** (checked into your project — recommended):

```bash
mkdir -p .vscode
```

`.vscode/mcp.json`:

```json
{
  "servers": {
    "sql-repository-mcp": {
      "type": "stdio",
      "command": "/absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp",
      "args": [],
      "env": {
        "LOG_LEVEL": "warn"
      }
    }
  }
}
```

**User-level** (available in every workspace):

Open VS Code settings (`Ctrl+,` / `Cmd+,`), search for `mcp`, click **Edit in settings.json**, and add:

```json
{
  "mcp": {
    "servers": {
      "sql-repository-mcp": {
        "type": "stdio",
        "command": "/absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp",
        "args": [],
        "env": {
          "LOG_LEVEL": "warn"
        }
      }
    }
  }
}
```

### 3. Use it in Copilot Chat

Open Copilot Chat (`Ctrl+Alt+I` / `Cmd+Alt+I`) and switch to **Agent mode** (`@` → `mcp`).

```
@mcp Use generate_repository with this schema:

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    stock INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Setup — Claude (Desktop & Claude Code)

### Claude Desktop (macOS / Windows)

**macOS** — edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "sql-repository-mcp": {
      "command": "/absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp",
      "args": [],
      "env": {
        "LOG_LEVEL": "warn"
      }
    }
  }
}
```

**Windows** — edit `%APPDATA%\Claude\claude_desktop_config.json` with the same structure. Use forward slashes or escaped backslashes in the path:

```json
{
  "mcpServers": {
    "sql-repository-mcp": {
      "command": "C:/Users/you/sql-to-go-mcp/bin/sql-repository-mcp.exe",
      "args": []
    }
  }
}
```

Restart Claude Desktop after saving. A hammer icon (🔨) appears in the chat input when the tool server is connected.

### Claude Code (CLI)

Add to your project's `.claude/mcp.json` (or run the command below):

```bash
claude mcp add sql-repository-mcp \
  /absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp
```

Or create `.claude/mcp.json` manually:

```json
{
  "mcpServers": {
    "sql-repository-mcp": {
      "command": "/absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp",
      "args": []
    }
  }
}
```

Verify the server is recognised:

```bash
claude mcp list
```

### Usage in Claude

```
Use the sql_to_gorm tool with this schema:

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

Module path: github.com/myorg/myapp
```

---

## Setup — ChatGPT (Desktop App)

The **ChatGPT desktop app** (macOS / Windows) supports MCP servers natively.

### macOS

Edit `~/Library/Application Support/ChatGPT/mcp.json` (create it if it does not exist):

```json
{
  "mcpServers": {
    "sql-repository-mcp": {
      "command": "/absolute/path/to/sql-to-go-mcp/bin/sql-repository-mcp",
      "args": [],
      "env": {
        "LOG_LEVEL": "warn"
      }
    }
  }
}
```

### Windows

Edit `%APPDATA%\ChatGPT\mcp.json`:

```json
{
  "mcpServers": {
    "sql-repository-mcp": {
      "command": "C:/Users/you/sql-to-go-mcp/bin/sql-repository-mcp.exe",
      "args": []
    }
  }
}
```

Restart the ChatGPT app after saving. A plug icon (🔌) in the toolbar indicates active MCP connections.

### Usage in ChatGPT

```
Use the generate_tests tool with this table definition:

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

Module path: github.com/myorg/myapp
```

---

## Tool Reference

### `generate_repository`

Generates the full repository layer in one call.

| Parameter | Required | Default | Description |
|---|---|---|---|
| `table_definition` | ✅ | — | SQL `CREATE TABLE` statement |
| `module_path` | | `github.com/example/app` | Go module path |
| `orm` | | `gorm` | `gorm` or `sqlc` |
| `driver` | | `postgres` | `postgres`, `mysql`, `sqlite`, `sqlserver` |
| `package` | | `repository` | Go package name for generated files |

### `sql_to_gorm`

| Parameter | Required | Default | Description |
|---|---|---|---|
| `schema` | ✅ | — | SQL `CREATE TABLE` statement |
| `module_path` | | `github.com/example/app` | Go module path |
| `package` | | `repository` | Go package name |

### `sql_to_sqlc`

| Parameter | Required | Default | Description |
|---|---|---|---|
| `schema` | ✅ | — | SQL `CREATE TABLE` statement |
| `query` | | — | Optional SQL query to analyse |
| `driver` | | `postgres` | `postgres`, `mysql`, `sqlite` |
| `module_path` | | `github.com/example/app` | Go module path |

### `benchmark_query`

| Parameter | Required | Description |
|---|---|---|
| `query` | ✅ | SQL query to analyse and benchmark |

Detects: `SELECT *`, missing `WHERE`, cartesian joins, leading-wildcard `LIKE`, missing `LIMIT`, `NOT IN (SELECT ...)`, functions on indexed columns, multiple `OR` conditions.

### `generate_tests`

| Parameter | Required | Default | Description |
|---|---|---|---|
| `table_definition` | ✅ | — | SQL `CREATE TABLE` statement |
| `module_path` | | `github.com/example/app` | Go module path |
| `package` | | `repository` | Go package name |

---

## Configuration

### `configs/config.yaml`

```yaml
server:
  port: 8080

database:
  driver: postgres
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: app

logging:
  level: info     # debug | info | warn | error
  format: json    # json | text

generation:
  orm: gorm       # gorm | sqlc
  testing: true
  benchmark: true
```

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `8080` | Server port |
| `DB_DRIVER` | `postgres` | Database driver |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `` | Database password |
| `DB_NAME` | `app` | Database name |
| `LOG_LEVEL` | `info` | Log level |
| `LOG_FORMAT` | `json` | Log format |

Priority: **CLI flags > environment variables > config.yaml**

### CLI Flags

```bash
./bin/sql-repository-mcp \
  --config configs/config.yaml \
  --log-level debug \
  --log-format text \
  --port 9090
```

---

## Supported Databases

| Driver value | Database |
|---|---|
| `postgres` | PostgreSQL ≥ 12 |
| `mysql` / `mariadb` | MySQL ≥ 5.7, MariaDB ≥ 10.3 |
| `sqlite` | SQLite 3 |
| `sqlserver` | SQL Server 2017+ |

---

## Supported SQL Types

| SQL Type | Go Type |
|---|---|
| `BIGSERIAL`, `BIGINT` | `int64` |
| `SERIAL`, `INTEGER`, `INT` | `int32` |
| `SMALLINT` | `int16` |
| `BOOLEAN` | `bool` |
| `FLOAT`, `DOUBLE`, `DECIMAL`, `NUMERIC` | `float64` |
| `VARCHAR`, `TEXT`, `CHAR`, `UUID` | `string` |
| `TIMESTAMP`, `TIMESTAMPTZ`, `DATETIME` | `time.Time` |
| `JSONB`, `JSON` | `json.RawMessage` |
| `BYTEA`, `BLOB` | `[]byte` |

Nullable columns (without `NOT NULL`) are automatically converted to pointer types (`*string`, `*time.Time`, etc.).

---

## Development

```bash
# Run tests
make test

# Run tests with verbose output
make test-v

# Generate coverage report
make test-cover

# Run benchmarks
make test-bench

# Build binary
make build

# Vet
make vet
```

---

## Project Structure

```
.
├── cmd/server/                  # Entry point
├── internal/
│   ├── config/                  # Configuration loader
│   ├── domain/
│   │   ├── entity/              # Core domain types (no framework deps)
│   │   └── port/                # Interfaces (ports)
│   ├── application/service/     # Use-case orchestration
│   └── adapters/
│       ├── mcp/                 # MCP server + tool handlers
│       └── db/repository/       # Database adapters
└── pkg/
    ├── sqlparser/               # SQL DDL parser
    ├── analyzer/                # Query anti-pattern detector
    ├── generator/               # Code generators
    ├── logger/                  # slog JSON logger
    └── database/                # Database connection utilities
```

---

## License

[Apache 2.0](LICENSE)
