# Getting Started

## Prerequisites

Ensure you have the following installed on your machine:

- Go 1.23
- Docker and Docker Compose
- Goose (for database migrations)

## Starting the Server

To start the server on your local machine, run the following command:

```
make server
```

## Running Database Migrations

To apply all pending migrations, use the following command:
```
make goose-up
```

For additional migration commands, you can use:
- Check migration status: 
```
make goose-status
```
- Roll back the last migration:
```
make goose-down
```