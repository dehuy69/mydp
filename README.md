# mydp
MyDP is an open-source database platform that supports multi-region deployments, providing flexibility and efficiency for managing both OLTP and OLAP workloads.

## Features

- **Storage Engine**: Utilizes Badger for OLTP tables and Parquet for OLAP tables for efficient data management.
- **Authentication**: Supports login authentication using admin user password.
- **User Management**: Admin can create other users, create small workspaces, and assign permissions to users for each workspace.
- **Workspace Structure**: Each workspace contains smaller units called collections and tables.
  - **Collections (OLTP Tables)**: Support common index types such as persistence, hash, and inverted indexes.
  - **Tables (OLAP Tables)**: Support OLAP workloads using Parquet as the storage engine.
- **API Communication**: Clients can connect and interact with the database through a well-defined API.
- **Pipeline Management**: Create pipelines using Jupyter Notebook and schedule them through the admin interface.
- **Catalog Management**: Uses SQLite to store information about workspaces, collections, tables, indexes, and pipelines.
- **Written in Go**: The entire platform is developed using the Go programming language for performance and concurrency.

## Getting Started

### Prerequisites

- Go 1.16 or higher
- BadgerDB
- DuckDB, Parquet
- SQLite3

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/dehuy69/mydp.git
   cd mydp