# Secure data storage manager - Gophkeeper

## Env example for .env file
```
# server
SERVER_HOST=localhost
SERVER_PORT=8099
JWT_SECRET=077e4fb1c8fd41ba8a99a480a8e0ee52

# database
DB_NAME=gophkeeper_db
DB_USER=gophkeeper
DB_PASSWORD=password
DB_PORT=5433
DB_HOST=localhost
DB_SSL_MODE=disable

# other
ENV=local
```

## Commands for launching and configuring applications
```
$ make deps                 # Install dependencies
$ make db-up                # Start test database container
$ make db-down              # Stop test database container
$ make proto                # Generate gRPC code from proto file
$ make build-server         # Build server binary
$ make build-client         # Build client binary
$ make test                 # Run tests(without cache)
$ make test-coverage        # Run tests with coverage
$ make coverage-percent     # See output coverage percent
$ make help                 # Read help message
```