# Distributed Database Slave GUI

A web-based GUI for interacting with the distributed database master node from a slave node. This interface allows users to perform database operations like selecting, inserting, updating, and deleting records.

## Features

- Connect to master node with status monitoring
- Select data from tables
- Insert new records
- Update existing records
- Delete records
- Responsive user interface with tabbed navigation

## Setup Instructions

### Prerequisites

- A running distributed database system with:
  - Master node (default: http://192.168.1.5:8001)
  - Slave node(s) (running on port 8002 or others)
- MySQL database installed and configured on all nodes
- Go environment set up for the backend

### Installation

1. *Save the HTML file:*
   
   Save the slave-gui.html file to your local machine.

2. *Run the file in a web browser:*
   
   You can open the HTML file directly in a web browser. No web server is required as it's a client-side application.

3. *Configure the master address:*
   
   If your master node is running on a different IP or port than the default (http://192.168.1.5:8001), update it in the GUI using the input field at the top of the page.

## Usage Guide

### Connecting to Master

1. The GUI automatically attempts to connect to the master node at startup
2. The status indicator shows green when connected, red when disconnected
3. You can update the master address manually if needed

### Select Records

1. Click on the "Select" tab
2. Enter the database name
3. Enter the table name
4. Click "Select Records" to fetch and display the data

### Insert Records

1. Click on the "Insert" tab
2. Enter the database name
3. Enter the table name
4. Enter the values as a comma-separated list (e.g., 'value1', 2, 'value3')
5. Click "Insert Record" to add the new record

### Update Records

1. Click on the "Update" tab
2. Enter the database name
3. Enter the table name
4. Enter the SET clause (e.g., column1='new value', column2=42)
5. Enter the WHERE clause to specify which records to update (e.g., id=1)
6. Click "Update Record" to apply the changes

### Delete Records

1. Click on the "Delete" tab
2. Enter the database name
3. Enter the table name
4. Enter the WHERE clause to specify which records to delete (e.g., id=1)
5. Click "Delete Record" to remove the records

## Troubleshooting

- If the master status shows offline, check that:
  - The master node is running
  - The IP address and port are correct
  - There are no network issues or firewalls blocking the connection
- If operations fail, check the error message displayed in the status area for details

## Architecture Notes

This GUI interacts with the distributed database system by:
1. Sending HTTP requests to the master node
2. The master node processes these requests and applies changes to the database
3. The master node replicates these changes to all slave nodes
4. Results are returned to the GUI and displayed to the user

## Security Considerations

- This GUI is designed for internal use within a trusted network
- No authentication is implemented in this basic version
- Consider implementing authentication and HTTPS for production use
