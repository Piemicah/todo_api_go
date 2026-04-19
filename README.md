# TODO API

### COMMANDS

1. gin

```bash
go get -u github.com/gin-gonic/gin

```

2. migrate tool

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

3. air

```bash
go install github.com/air-verse/air@latest

```

3. dotenv

```bash
go get -u github.com/joho/godotenv

```

4. bcrypt

```bash
go get -u golang.org/x/crypto/bcrypt

```

5. JWT

```bash
go get -u github.com/golang-jwt/jwt/v5

```

6. Postgres driver

```bash
go get -u github.com/jackc/pgx/v5

```

7. Postgres connection pool

```bash
go get -u github.com/jackc/pgx/v5/pgxpool

```

8. migration command

```bash
migrate create -ext sql -dir migrations -seq create_todos_table

```
