-- name: CreateUser :one
INSERT INTO users (id, username, email, phone, hashed_password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
