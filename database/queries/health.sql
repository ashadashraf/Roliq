-- name: DatabaseNow :one
SELECT now()::timestamptz;

-- name: FindIdentity :one
SELECT ai.user_id, u.email, u.display_name
FROM auth_identities ai
JOIN users u ON u.id = ai.user_id
WHERE ai.issuer = $1 AND ai.subject = $2 AND u.deleted_at IS NULL;

-- name: FindMembershipsForUser :many
SELECT om.organization_id, om.role, o.name, o.slug, o.type
FROM organization_memberships om
JOIN organizations o ON o.id = om.organization_id
WHERE om.user_id = $1 AND o.deleted_at IS NULL
ORDER BY om.created_at;
