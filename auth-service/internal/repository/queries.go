package repository

const (
	QueryCreateSession = `INSERT INTO sessions (user_id, token_hash, refresh_token_hash, expires_at, created_at, updated_at, user_agent, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`
	QueryGetByTokenHash = `SELECT id, user_id, token_hash, refresh_token_hash, expires_at, created_at, updated_at, user_agent, ip_address
		FROM sessions
		WHERE token_hash = $1`
	QueryGetSessionByRefreshToken = `SELECT id, user_id, token_hash, refresh_token_hash, expires_at, created_at, updated_at, user_agent, ip_address
		FROM sessions
		WHERE refresh_token_hash = $1`
	QueryDeleteSession        = `DELETE FROM sessions WHERE id = $1`
	QueryDeleteExpiredSession = `DELETE FROM sessions WHERE expires_at < NOW()`

	QueryCreateUser = `INSERT INTO users (email, username, password_hash, display_name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	QueryGetUserByID = `SELECT id, email, username, password_hash, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1`
	QueryGetUserByEmail = `SELECT id, email, username, password_hash, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1`
	QueryGetUserByUsername = `SELECT id, email, username, password_hash, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE username = $1`
	QueryUpdateUser = `UPDATE users
		SET email = $2, username = $3, display_name = $4, avatar_url = $5, updated_at = $6
		WHERE id = $1`
)
