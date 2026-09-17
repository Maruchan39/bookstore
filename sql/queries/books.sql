-- name: GetBooks :many
SELECT
    *
FROM
    books
ORDER BY
    created_at ASC;