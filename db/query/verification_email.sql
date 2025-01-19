-- name: CreateVerificationEmail :one
INSERT INTO verification_emails (
    username,
    email,
    secret_code
) VALUES (
    $1, $2, $3
) RETURNING *;