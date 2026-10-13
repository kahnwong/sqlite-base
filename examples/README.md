# Examples

## Init

```bash
go get -tool github.com/sqlc-dev/sqlc/cmd/sqlc
```

## Usage

After changing `sql/schema.sql` or `sql/queries.sql`:

```bash
go tool sqlc generate -f sqlc.yaml
```

Output is in `db`.

Then create goose migration file by hand.

## Run

```bash
go run .
```
