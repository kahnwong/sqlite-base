# Examples

## Usage

After changing `db/schema.sql` or `db/queries.sql`:

```bash
go tool sqlc generate -f sqlc.yaml
```

Output is in `store`.

Then create goose migration file by hand.

## Run

```bash
go run .
```
