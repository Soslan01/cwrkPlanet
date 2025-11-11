package queries

const (
	QueryCreateSession = `
		INSERT INTO auth_sessions (
			user_id, token_hash, expires_at, created_at, updated_at, user_agent, ip
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	QueryGetSessionByTokenHash = `
		SELECT
			id, user_id, token_hash, expires_at, created_at, updated_at,
			user_agent,
			CASE WHEN ip IS NULL THEN NULL ELSE ip::text END AS ip_text
		FROM auth_sessions
		WHERE token_hash = $1
		LIMIT 1;
	`
	QueryDeleteSessionByID           = `DELETE FROM auth_sessions WHERE id = $1;`
	QueryDeleteSessionByUser         = `DELETE FROM auth_sessions WHERE user_id = $1;`
	QueryDeleteSessionsExpiredByTime = `DELETE FROM auth_sessions WHERE expires_at <= $1;`
)
