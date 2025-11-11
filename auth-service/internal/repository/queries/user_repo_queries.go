package queries

const (
	QueryCreateUser = `
		INSERT INTO users (email, email_verified, password_hash, display_name, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id;
	`
	QueryGetUserByID = `
		SELECT id, email, email_verified, password_hash, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE id = $1;
	`
	QueryGetUserByEmail = `
		SELECT id, email, email_verified, password_hash, display_name, avatar_url, created_at, updated_at
		FROM users
		WHERE email = $1;
	`
	QueryExistsUserByEmail  = `SELECT 1 FROM users WHERE email = $1;`
	QueryUpdatePasswordHash = `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1;
	`
	QueryUpdateEmailVerified = `
		UPDATE users
		SET email_verified = TRUE, updated_at = $2
		WHERE id = $1;
	`
)
