# Distributed Master-Slave Database System

A robust, fault-tolerant distributed database management system with master-slave architecture built with Go and MySQL.
## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [System Requirements](#system-requirements)
- [Installation & Setup](#installation--setup)
  - [Prerequisites](#prerequisites)
  - [Database Setup](#database-setup)
  - [System Configuration](#system-configuration)
  - [Running the System](#running-the-system)
- [Usage Examples](#usage-examples)
  - [Database Operations](#database-operations)
  - [Table Operations](#table-operations)
  - [Data Operations](#data-operations)
  - [Search Operations](#search-operations)
- [API Reference](#api-reference)
- [System Architecture](#system-architecture)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Overview

This distributed database system implements a master-slave architecture for high availability and fault tolerance. It allows for automatic failover in case the master node becomes unavailable, with slaves capable of being promoted to master roles. The system handles database operations via a RESTful API and maintains data consistency through asynchronous replication.
## Architecture Overview

```
+-------------------------+
|      Master Node        |
|-------------------------|
| - DB Write Access       |
| - Broadcast to Slaves   |
+-------------------------+
            |
      +-----+-----+
      |           |
      ▼           ▼
+-------------+  +-------------+
| Slave Node  |  | Slave Node  |
|-------------|  |-------------|
| - Read-only |  | - Read-only |
|   DB        |  |   DB        |
| - Listen to |  | - Listen to |
|   MQ        |  |   MQ        |
+-------------+  +-------------+
```

## Features

- **Master-slave architecture** with automatic failover
- **Database operations** - creation, modification, and deletion
- **Table management** with custom schemas
- **CRUD operations** for data management
- **Asynchronous replication** with retry mechanisms
- **Search functionality** across databases
- **Health monitoring** and leader election
- **REST API** interface with CORS support

## System Requirements

- Go 1.15+
- MySQL 5.7+ or 8.0+
- Network connectivity between nodes

## Installation & Setup

### Prerequisites

1. Install Go:
   ```
   # For Windows
   # Download and install from https://golang.org/dl/
   ```

2. Install MySQL:
   ```
   # For Windows
   # Download and install from https://dev.mysql.com/downloads/installer/
   ```

3. Install required Go packages:
   ```
   go get github.com/go-sql-driver/mysql
   ```

### System Configuration

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/distributed-db.git
   cd distributed-db
   ```

2. Configure master node:
   
   Edit the `master.go` file to update slave addresses if needed:
   ```go
   var slaveAddresses = []string{
     "http://localhost:8009", 
     "http://172.31.243.32:8002",
     // Add more slaves as needed
   }
   ```

3. Configure slave nodes:
   
   Edit each `slave.go` file to set the master address:
   ```go
   var masterAddress string = "http://192.168.137.33:8001"
   ```

### Running the System

1. Start the master node:
   ```
   go run master.go
   ```
   You should see: `Master server running on port 8001...`

2. Start each slave node:
   ```
   go run slave.go
   ```
   You should see: `Slave server running on port 8009...`

3. Verify the setup by checking master status:
   ```
   curl http://localhost:8001/is-master
   ```
   Expected response: `{"isMaster":true}`

## Usage Examples

### Database Operations

#### Create a Database
```bash
curl -X GET "http://localhost:8001/createdb?name=testdb"
```

#### Drop a Database
```bash
curl -X GET "http://localhost:8001/dropdb?name=testdb"
```

### Table Operations

#### Create a Table
```bash
curl -X GET "http://localhost:8001/createtable?dbname=testdb&table=users&schema=name VARCHAR(100), email VARCHAR(100), age INT"
```

### Data Operations

#### Insert Data
```bash
curl -X POST "http://localhost:8001/insert" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "users", 
    "values": "NULL, \"John Doe\", \"john@example.com\", 30"
  }'
```

#### Select Data
```bash
curl -X GET "http://localhost:8001/select?dbname=testdb&table=users"
```

#### Update Data
```bash
curl -X POST "http://localhost:8001/update" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "users", 
    "set": "name = \"Jane Doe\", age = 31", 
    "where": "Id = 1"
  }'
```

#### Delete Data
```bash
curl -X POST "http://localhost:8001/delete" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "users", 
    "where": "Id = 1"
  }'
```

## API Reference

### Master Node Endpoints

| Endpoint | Method | Description | Parameters |
|----------|--------|-------------|------------|
| `/ping` | GET | Health check | None |
| `/nodes` | GET | List all nodes | None |
| `/is-master` | GET | Check if node is master | None |
| `/createdb` | GET | Create a database | `name` (query param) |
| `/dropdb` | GET | Drop a database | `name` (query param) |
| `/createtable` | GET/POST | Create a table | `dbname`, `table`, `schema` (query params) or JSON body with columns |
| `/insert` | POST | Insert record | JSON body with insertion details |
| `/select` | GET | Select records | `dbname`, `table` (query params) |
| `/search` | GET | Search records | `dbname`, `table`, `column`, `value` (query params) |
| `/update` | POST | Update records | JSON body with update details |
| `/delete` | POST | Delete records | JSON body with deletion details |

### Slave Node Endpoints

| Endpoint | Method | Description | Parameters |
|----------|--------|-------------|------------|
| `/ping` | GET | Health check | None |
| `/is-master` | GET | Check if node is master | None |
| `/replicate/db` | GET | Replicate database | `name` (query param) |
| `/replicate/dropdb` | GET | Replicate database drop | `name` (query param) |
| `/replicate/table` | GET | Replicate table creation | `dbname`, `table`, `schema` (query params) |
| `/replicate/insert` | POST | Replicate insert | JSON body with insertion details |
| `/replicate/update` | POST | Replicate update | JSON body with update details |
| `/replicate/delete` | POST | Replicate delete | JSON body with deletion details |
| `/search` | GET | Search records | `dbname`, `table`, `column`, `value` (query params) |

## System Architecture Details

The distributed database system uses a master-slave architecture for high availability and fault tolerance:

1. **Master Node**
   - Handles all write operations (INSERT, UPDATE, DELETE)
   - Coordinates data replication to slave nodes
   - Manages database and table creation/deletion
   - Automatically broadcasts changes to all registered slaves

2. **Slave Nodes**
   - Process read operations (SELECT, SEARCH)
   - Maintain synchronized copies of master data
   - Can be promoted to master in case of master failure
   - Listen for updates from master node

3. **Replication Flow**
   - Client sends write request to master
   - Master processes request on its database
   - Master broadcasts change to all slave nodes
   - Slaves apply change to their local databases
   - Acknowledgement sent back to master

4. **Failover Mechanism**
   - Health monitoring between nodes
   - Automatic detection of failed master
   - Election process to promote slave to new master
   - System reconfiguration for new topology


