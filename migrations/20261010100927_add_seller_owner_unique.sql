-- +goose Up

CREATE UNIQUE INDEX sellers_owner_id_unique
ON sellers(owner_id);
-- +goose Down
DROP INDEX IF EXISTS sellers_owner_id_unique;
