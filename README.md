# Chirpy

Running migrations:
- `goose postgres <connection_url> up`
- `goose postgres <connection_url> down`

Compile and run:
- `go build -o out && ./out`

Generate sql code:
- `sqlc generate`