# Go Service Template

## Tech stack

- [Echo](https://echo.labstack.com)
- PostgreSQL with [pgx/v5](https://github.com/jackc/pgx), [sqlc](https://sqlc.dev) and [goose](https://github.com/pressly/goose)
- S3 with [AWS SDK](https://github.com/aws/aws-sdk-go-v2) ([RustFS](https://rustfs.com) included in compose)
- Valkey
- Kafka with [segmentio/kafka-go](https://github.com/segmentio/kafka-go)

## Structure

```
.
├── cmd/
│   └── main.go - entry point for the whole application, loads a config, creates an app and manages its lifecycle
├── internal
│   ├── app/
│   │   └── app.go - initialization of all components
│   ├── config/
│   │   └── config.go - config structs and loading from .env
│   ├── db/
│   │   ├── migrate.go - simple function to migrate db
│   │   ├── migrations/ - there would be your migrations
│   │   └── sqlc/ - sqlc config and queries
│   ├── errorz/ - domain-level errors
│   ├── models/ - domain-level models
│   ├── repository/ - storage, repo level
│   ├── service/ - use-case, business-logic and whatever else it's called
│   └── transport/ - top-level transport communications
│       └── http/ - specifically http transport
├── pkg/
│   └── dberrors/ - some sugar for db errors processing
│       └── is_unique_violation.go
```
