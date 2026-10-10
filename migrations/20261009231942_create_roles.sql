-- +goose Up

CREATE TABLE roles (
    id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    role_id SMALLINT NOT NULL
        REFERENCES roles(id) ON DELETE CASCADE,

    PRIMARY KEY (user_id, role_id)
);

INSERT INTO roles (name)
VALUES
    ('buyer'),
    ('admin'),
    ('super_admin');

-- +goose Down

DROP TABLE user_roles;
DROP TABLE roles;