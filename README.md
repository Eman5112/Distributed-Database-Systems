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

## Features

- Master-slave architecture with automatic failover
- Database creation, modification, and deletion
- Table management with custom schemas
- CRUD operations for data management
- Asynchronous replication with retry mechanisms
- Search functionality
- Health monitoring and leader election
- REST API interface
- Cross-Origin Resource Sharing (CORS) support

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

4. Check slave status:
   ```
   curl http://localhost:8009/is-master
   ```
   Expected response: `{"isMaster":false}`

## Usage Examples

### Database Operations

#### Create a Database

**Request:**
```bash
curl -X GET "http://localhost:8001/createdb?name=testdb"
```

**Response:**
```json
{"message":"Database created successfully"}
```

#### Drop a Database

**Request:**
```bash
curl -X GET "http://localhost:8001/dropdb?name=testdb"
```

**Response:**
```json
{"message":"Database dropped successfully"}
```

### Table Operations

#### Create a Table

**Method 1: Using Schema String**

**Request:**
```bash
curl -X GET "http://localhost:8001/createtable?dbname=testdb&table=users&schema=name VARCHAR(100), email VARCHAR(100), age INT"
```

**Response:**
```json
{"message":"Table created successfully"}
```

**Method 2: Using JSON Body**

**Request:**
```bash
curl -X POST "http://localhost:8001/createtable?dbname=testdb&table=products" \
  -H "Content-Type: application/json" \
  -d '{
    "columns": [
      {"Name": "product_name", "DataType": "VARCHAR(100)"},
      {"Name": "price", "DataType": "DECIMAL(10,2)"},
      {"Name": "inventory", "DataType": "INT"}
    ]
  }'
```

**Response:**
```json
{"message":"Table created successfully"}
```

### Data Operations

#### Insert Data

**Method 1: Using Values String**

**Request:**
```bash
curl -X POST "http://localhost:8001/insert" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "users", 
    "values": "NULL, \"John Doe\", \"john@example.com\", 30"
  }'
```

**Method 2: Using Records Object**

**Request:**
```bash
curl -X POST "http://localhost:8001/insert" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "products", 
    "records": {
      "product_name": "Laptop", 
      "price": 999.99, 
      "inventory": 50
    }
  }'
```

**Response:**
```json
{"message":"Record inserted successfully"}
```

#### Select Data

**Request:**
```bash
curl -X GET "http://localhost:8001/select?dbname=testdb&table=users"
```

**Response:**
```json
[
  {"Id": 1, "name": "John Doe", "email": "john@example.com", "age": 30}
]
```

#### Update Data

**Request:**
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

**Response:**
```json
{"message":"Record updated successfully"}
```

#### Delete Data

**Request:**
```bash
curl -X POST "http://localhost:8001/delete" \
  -H "Content-Type: application/json" \
  -d '{
    "dbname": "testdb", 
    "table": "users", 
    "where": "Id = 1"
  }'
```

**Response:**
```json
{"message":"Record deleted successfully"}
```

### Search Operations

**Request:**
```bash
curl -X GET "http://localhost:8001/search?dbname=testdb&table=products&column=product_name&value=Laptop"
```

**Response:**
```json
[
  {"Id": 1, "product_name": "Laptop", "price": "999.99", "inventory": 50}
]
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

## System Architecture

The system uses a master-slave architecture:

1. **Master Node** - Handles all write operations and coordinates replication
2. **Slave Nodes** - Process read operations and maintain replicated copies of data
3. **Failover Mechanism** - Automatic promotion of slaves to master role when needed

For more details, refer to the [architecture document](./ARCHITECTURE.md).

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure MySQL is running: `sudo systemctl status mysql`
   - Verify database credentials in the code

2. **Replication Failures**
   - Check network connectivity between nodes
   - Ensure all nodes are running and accessible

3. **Permission Denied**
   - Verify MySQL user permissions: `SHOW GRANTS FOR 'root'@'localhost';`

### Logs

- Check application logs for detailed error messages
- MySQL logs are typically found at `/var/log/mysql/error.log`

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
