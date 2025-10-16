-- name: CreateFeed :one
INSERT INTO feeds(id, created_at, updated_at, name, url, user_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
       )
    RETURNING *;

-- name: GetFeeds :many
SELECT feeds.name as feed_name, feeds.url, users.name as user_name FROM feeds
INNER JOIN users
ON users.id = feeds.user_id;

-- name: GetFeedByURL :one
SELECT * FROM feeds WHERE url = $1;

-- name: MarkFeedFetched :one
UPDATE feeds
SET updated_at = CURRENT_TIMESTAMP, last_fetched_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT id, name, url,last_fetched_at FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST, id ASC
LIMIT 1;

