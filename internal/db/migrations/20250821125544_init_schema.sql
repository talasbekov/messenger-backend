-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE contact_state AS ENUM ('pending', 'accepted', 'blocked');
CREATE TYPE contact_request_state AS ENUM ('pending', 'accepted', 'rejected', 'blocked');

CREATE TABLE users (
    id              TEXT PRIMARY KEY,
    username        TEXT UNIQUE NOT NULL,
    phone           TEXT UNIQUE,
    email           TEXT UNIQUE,
    hashed_password TEXT NOT NULL,
    status          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE devices (
    id          TEXT PRIMARY KEY,
    user_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    platform    TEXT,
    push_token  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at  TIMESTAMPTZ
);

CREATE TABLE auth_sessions (
    id              TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id       TEXT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    refresh_token   TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contacts (
    id          TEXT PRIMARY KEY,
    owner_id    TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    peer_id     TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    alias       TEXT,
    state       contact_state NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(owner_id, peer_id)
);

CREATE TABLE contact_requests (
    id          TEXT PRIMARY KEY,
    from_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id   TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    state       contact_request_state NOT NULL DEFAULT 'pending',
    message     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(from_user_id, to_user_id)
);

CREATE TABLE blocks (
    owner_id        TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id  TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (owner_id, target_user_id)
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_contacts_owner_id_state ON contacts(owner_id, state);
CREATE INDEX idx_devices_user_id ON devices(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS contact_requests;
DROP TABLE IF EXISTS contacts;
DROP TABLE IF EXISTS auth_sessions;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS contact_request_state;
DROP TYPE IF EXISTS contact_state;
-- +goose StatementEnd
