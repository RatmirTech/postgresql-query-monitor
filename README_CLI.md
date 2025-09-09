# PostgreSQL Query Monitor CLI

A command-line tool for monitoring PostgreSQL databases, collecting metrics, parsing logs, and reviewing SQL queries using an external review API.

## Features

- **Multi-collector support**: Collect different types of data using subcommands
- **PostgreSQL Logs**: Parse PostgreSQL log files and extract SQL queries for review
- **PostgreSQL Metrics**: Collect database-specific metrics and statistics
- **System Metrics**: Monitor system resources (CPU, RAM, Disk)
- **SQL Review Integration**: Automatically send SQL queries for review and store results
- **Vault Integration**: Secure database credential management via HashiCorp Vault
- **Airflow Ready**: Designed to work with Apache Airflow for scheduling

## Installation

1. Clone the repository
2. Build the CLI:
   ```bash
   go build -o pgmon ./cmd/cli
   ```

## Configuration

Copy the example environment file and configure:
```bash
cp .env.sample .env
```

Edit `.env` with your settings:
```
VAULT_ADDR=http://localhost:8200
VAULT_TOKEN=your_vault_token
VAULT_DB_PATH=database/config
VAULT_TARGET_DB_PATH=database/target
REVIEW_API_URL=http://185.159.111.235:8000/review/
PG_LOG_PATH=/var/log/postgresql
ENVIRONMENT=production
```

## Database Setup

Run the migration script to create required tables:
```sql
-- Connect to your results database and run:
\i migrations/001_create_review_tables.sql
```

## Usage

### Basic Command Structure
```bash
./pgmon collect [collector_types...] [flags]
```

### Collector Types

#### 1. PostgreSQL Logs (`pglogs`)
Parses PostgreSQL log files and sends SQL queries for review:
```bash
./pgmon collect pglogs --log-path /var/log/postgresql --target-db database/target
```

#### 2. PostgreSQL Metrics (`pgmetrics`) 
Collects database performance metrics:
```bash
./pgmon collect pgmetrics --target-db database/target
```

#### 3. System Metrics (`sysmetrics`)
Monitors system resources:
```bash
./pgmon collect sysmetrics
```

### Combined Collection
Run multiple collectors together:
```bash
./pgmon collect pglogs pgmetrics sysmetrics --target-db database/target --env production
```

### Command Line Flags

- `--vault-path`: Vault path for our database credentials (where reviews are stored)
- `--target-db`: Vault path for target database to monitor  
- `--log-path`: Path to PostgreSQL log files (default: /var/log/postgresql)
- `--env`: Environment name (default: production)

## Airflow Integration

The CLI is designed to work with Apache Airflow. Example DAG task:

```python
from airflow.operators.bash_operator import BashOperator

collect_task = BashOperator(
    task_id='collect_pg_data',
    bash_command='/path/to/pgmon collect pglogs pgmetrics sysmetrics --target-db database/{{ params.db_name }}',
    params={'db_name': 'production_db'},
    dag=dag
)
```

## Architecture

### Components

1. **CLI Layer** (`cmd/cli/main.go`): Command-line interface using Cobra
2. **Collectors** (`internal/collectors/`): Data collection modules
   - `pglogs.go`: PostgreSQL log parser
   - `pgmetrics.go`: Database metrics collector  
   - `sysmetrics.go`: System metrics collector
3. **Review Client** (`internal/review/`): HTTP client for SQL review API
4. **Storage Service** (`internal/storage/`): Database operations for storing results
5. **Configuration** (`internal/config/`): Environment and CLI configuration
6. **Vault Integration** (`pkg/vault/`): HashiCorp Vault client

### Data Flow

1. CLI parses command and initializes collectors
2. Collectors gather data from various sources
3. SQL queries are sent to review API
4. Results are stored in PostgreSQL database via Vault credentials
5. Metrics are logged/exported for monitoring

## SQL Review API

The tool integrates with an external SQL review service:

**Endpoint**: `http://185.159.111.235:8000/review/`

**Request Format**:
```json
{
  "sql": "SELECT * FROM orders WHERE customer_id = 123",
  "query_plan": "EXPLAIN output",
  "tables": [
    {
      "name": "orders", 
      "columns": [{"name": "customer_id", "type": "int"}]
    }
  ],
  "server_info": {"version": "15.0"},
  "environment": "production"
}
```

**Response Format**:
```json
{
  "errors": [],
  "overall_score": 100,
  "notes": "Query meets requirements",
  "thread_id": "040d285a-197e-4b94-9c9e-7972771facf0"
}
```

## Database Schema

Review results are stored in the following tables:

- `query_reviews`: Main table for review requests/responses
- `table_structures`: Table structure information for queries

See `migrations/001_create_review_tables.sql` for complete schema.

## Examples

### Monitoring Production Database
```bash
# Set up environment
export VAULT_TOKEN=your_token
export VAULT_DB_PATH=database/results
export VAULT_TARGET_DB_PATH=database/production

# Collect all data types
./pgmon collect pglogs pgmetrics sysmetrics --env production
```

### Development/Testing
```bash
./pgmon collect pglogs \
  --target-db database/staging \
  --log-path /opt/postgresql/logs \
  --env staging
```

### Custom Vault Paths
```bash
./pgmon collect pgmetrics \
  --vault-path database/results/dev \
  --target-db database/app1/prod
```

## Troubleshooting

### Common Issues

1. **Vault Connection Error**: Ensure VAULT_ADDR and VAULT_TOKEN are set correctly
2. **Log File Access**: Check that log path exists and is readable
3. **Database Connection**: Verify Vault paths contain valid database credentials
4. **Review API Timeout**: Check network connectivity to review service

### Logs and Debugging

The CLI outputs detailed logs for debugging:
- Connection status to databases and services
- Number of records collected/processed
- API response summaries
- Error details with context

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable  
5. Submit a pull request

## License

[Add your license information here]
