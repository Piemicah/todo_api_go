set -a
source ./.env
set +a

migrate -path migrations -database $DATABASE_URL up