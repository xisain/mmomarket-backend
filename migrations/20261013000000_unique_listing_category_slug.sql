-- +goose Up
ALTER TABLE listing_categorys
    ADD CONSTRAINT listing_categorys_slug_key UNIQUE (slug);

-- +goose Down
ALTER TABLE listing_categorys DROP CONSTRAINT listing_categorys_slug_key;
