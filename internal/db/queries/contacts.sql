-- name: FindUserByIdentifier :one
SELECT *
FROM users
WHERE username = sqlc.narg(username) OR email = sqlc.narg(email) OR phone = sqlc.narg(phone);

-- name: CreateContactRequest :one
INSERT INTO contact_requests (id, from_user_id, to_user_id, message)
VALUES ($1, $2, $3, sqlc.narg(message))
RETURNING *;

-- name: GetContactRequest :one
SELECT * FROM contact_requests
WHERE id = $1;

-- name: UpdateContactRequestState :exec
UPDATE contact_requests
SET state = $2
WHERE id = $1;

-- name: CreateContact :one
INSERT INTO contacts (id, owner_id, peer_id, state)
VALUES ($1, $2, $3, 'accepted')
RETURNING *;

-- name: ListContacts :many
SELECT * FROM contacts
WHERE owner_id = $1 AND state = $2;

-- name: DeleteContact :exec
DELETE FROM contacts
WHERE owner_id = $1 AND peer_id = $2;

-- name: CreateBlock :exec
INSERT INTO blocks (owner_id, target_user_id)
VALUES ($1, $2)
ON CONFLICT (owner_id, target_user_id) DO NOTHING;

-- name: DeleteBlock :exec
DELETE FROM blocks
WHERE owner_id = $1 AND target_user_id = $2;

-- name: IsBlocked :one
SELECT EXISTS(
    SELECT 1 FROM blocks
    WHERE (owner_id = $1 AND target_user_id = $2) OR (owner_id = $2 AND target_user_id = $1)
);
