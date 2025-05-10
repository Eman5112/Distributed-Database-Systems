# Distributed Database System

A fault-tolerant distributed database system with master-slave architecture, built in Go. This system provides a RESTful API for database operations with automatic replication across multiple nodes and master election when the primary node fails.

## Features

- **Master-Slave Architecture**: One master node handles write operations and replicates changes to slave nodes
- **Automatic Failover**: If the master node goes down, a slave node will automatically be elected as the new master
- **RESTful API**: Simple HTTP API for database operations
- **Cross-Origin Resource Sharing (CORS)**: Supports cross-origin requests
- **Fault Tolerance**: Implements retry mechanisms for replication with exponential backoff
- **Database Operations**: Support for creating/dropping databases, creating tables, and CRUD operations on records

## Prerequisites

- Go 1.16 or higher
- MySQL server installed and running
- Git (for cloning the repository)

## Setup Instructions

### 1. Clone the Repository

```bash
git clone <repository-url>
cd distributed-database-system
```

### 2. Configure MySQL

Ensure MySQL is running and create user accounts with appropriate permissions:

For master node:
```sql
CREATE USER 'root'@'localhost' IDENTIFIED BY 'root';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
```

For slave nodes:
```sql
CREATE USER 'root'@'localhost' IDENTIFIED BY 'rootroot';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'localhost' WITH GRANT OPTION;
FLUSH PRIVILEGES;
```

### 3. Configure Node Addresses

Edit the master server configuration in `main.go` to update slave addresses if needed:

```go
var slaveAddresses = []string{
    "http://192.168.1.4:8002", 
    "http://192.168.1.8:8005", 
    "http://192.168.1.6:8004", 
    "http://192.168.1.7:8006"
}
```

### 4. Build and Run the Master Node

```bash
go build -o master
./master
```

The master server will be running on port 8001.

### 5. Build and Run Slave Nodes

Update the slave configuration (port and credentials) in each slave's main.go file if needed, then:

```bash
go build -o slave
./slave
```

Repeat for each slave node on different servers, changing the port number and MySQL credentials as necessary.

## API Usage Examples

### Check Node Status

```bash
curl http://localhost:8001/ping
```

Expected response:
```
pong
```

### Check if Node is Master

```bash
curl http://localhost:8001/is-master
```

Expected response:
```json
{"isMaster":true}
```

### Get Node Information
```bash
curl http://localhost:8001/nodes
```

Expected response:
```json
{
  "master": "http://localhost:8001",
  "slaves": [
    "http://192.168.1.4:8002",
    "http://192.168.1.8:8005",
    "http://192.168.1.6:8004",
    "http://192.168.1.7:8006"
  ],
  "isMaster": true
}
```

### Create a Database

```bash
curl "http://localhost:8001/createdb?name=testdb"
```

Expected response:
```json
{"message":"Database created successfully"}
```

### Create a Table

#### Method 1: Using schema parameter

```bash
curl -X POST "http://localhost:8001/createtable?dbname=testdb&table=users&schema=id%20INT%20AUTO_INCREMENT%20PRIMARY%20KEY,%20name%20VARCHAR(50),%20email%20VARCHAR(100)"
```

#### Method 2: Using JSON body with columns

```bash
curl -X POST "http://localhost:8001/createtable?dbname=testdb&table=users" \
  -H "Content-Type: application/json" \
  -d '{
    "columns": [
      {"Name": "name", "DataType": "VARCHAR(50)"},
      {"Name": "email", "DataType": "VARCHAR(100)"}
    ]
  }'
```

Expected response:
```json
{"message":"Table created successfully"}
```

### Insert a Record

#### Method 1: Using values string

```bash
curl -X POST "http://localhost:8001/insert" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb",
    "table": "users",
    "values": "NULL, \"John Doe\", \"john@example.com\""
  }'
```

#### Method 2: Using records object

```bash
curl -X POST "http://localhost:8001/insert" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb",
    "table": "users",
    "records": {
      "name": "Jane Doe",
      "email": "jane@example.com"
    }
  }'
```

Expected response:
```json
{"message":"Record inserted successfully"}
```

### Query Records

```bash
curl "http://localhost:8001/select?dbname=testdb&table=users"
```

Expected response:
```json
[
  {
    "Id": 1,
    "email": "john@example.com",
    "name": "John Doe"
  },
  {
    "Id": 2,
    "email": "jane@example.com",
    "name": "Jane Doe"
  }
]
```

### Update a Record

```bash
curl -X POST "http://localhost:8001/update" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb",
    "table": "users",
    "set": "name = \"John Smith\"",
    "where": "Id = 1"
  }'
```

Expected response:
```json
{"message":"Record updated successfully"}
```

### Delete a Record

```bash
curl -X POST "http://localhost:8001/delete" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb",
    "table": "users",
    "where": "Id = 1"
  }'
```

Expected response:
```json
{"message":"Record deleted successfully"}
```

### Drop a Database

```bash
curl "http://localhost:8001/dropdb?name=testdb"
```

Expected response:
```json
{"message":"Database dropped successfully"}
```

## Failover Process

The system automatically checks the health of the master node every 10-15 seconds. If the master becomes unavailable, an election process starts among the slave nodes and a new master is elected.

To test failover:
1. Start the master and at least one slave node
2. Terminate the master process
3. Wait approximately 15 seconds
4. Observe that one of the slave nodes becomes the new master

You can check by running:
```bash
curl http://localhost:8006/is-master
```

## Architecture Overview

### Master Node Responsibilities
- Accept all client requests (read/write operations)
- Execute operations on its local database
- Replicate changes to all slave nodes
- Monitor its own health

### Slave Node Responsibilities
- Maintain a replica of the master's data
- Accept replication requests from the master
- Monitor the master's health
- Participate in leader election when the master fails
- Take over as the new master when elected

## Implementation Details

- The system uses HTTP for communication between nodes
- MySQL is used as the underlying database engine
- Replication is done through the master pushing changes to slaves
- Leader election uses a simple algorithm with randomized delays to prevent conflicts
- Error handling includes retries with exponential backoff

## Limitations and Future Improvements

- Current implementation doesn't handle network partitions correctly
- Authentication and security features are minimal
- No built-in data partitioning/sharding
- No automatic synchronization of a slave that was down and comes back online
- Could add support for read replicas to distribute read operations

## License

[License Information]
