-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3
    )
    RETURNING *
)
SELECT inserted_feed_follow.*,
       feeds.name AS feed_name,
       users.name AS user_name
       FROM inserted_feed_follow
INNER JOIN users ON users.id = inserted_feed_follow.user_id
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT feeds.name AS feed_name FROM feed_follows AS ff
INNER JOIN feeds ON ff.feed_id = feeds.id
WHERE ff.user_id = $1;