package databases

func createTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) UNIQUE NOT NULL,
		email VARCHAR(200) UNIQUE NOT NULL,
		password_hash VARCHAR(225) NOT NULL,
		full_name VARCHAR(225) NOT NULL,
		bio TEXT NULL,
		role ENUM('user', 'admin') DEFAULT 'user',
		avatar_url VARCHAR(225) NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	)`
	_, err := DB.Exec(usersTable)
	if err != nil {
		panic("Failed to create users table: " + err.Error())
	}

	createRefreshTokenTable := `CREATE TABLE IF NOT EXISTS refresh_tokens ( 
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, 
		user_id BIGINT UNSIGNED NOT NULL, 
		device_name VARCHAR(100) NULL, 
		token_hash VARCHAR(225) NOT NULL, 
		expires_at DATETIME NOT NULL, 
		revoked_at DATETIME NULL, 
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, 
		
		CONSTRAINT token_user_id_fk 
			FOREIGN KEY(user_id) 
			REFERENCES users (id) 
			ON DELETE CASCADE 
			ON UPDATE CASCADE
	)`
	_, err = DB.Exec(createRefreshTokenTable)
	if err != nil {
		panic("Failed to create refresh_tokens table: " + err.Error())
	}

	friendshipsTable := `CREATE TABLE IF NOT EXISTS friendships (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		requester_id BIGINT UNSIGNED,
		addressee_id BIGINT UNSIGNED,
		status ENUM('pending', 'accepted', 'blocked') DEFAULT 'pending',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

		CONSTRAINT fr_req_id_fk
			FOREIGN KEY(requester_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT fr_add_id_fk
			FOREIGN KEY(addressee_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT fr_req_add_uq UNIQUE (requester_id, addressee_id)
	)`
	_, err = DB.Exec(friendshipsTable)
	if err != nil {
		panic("Failed to create friendships table: " + err.Error())
	}

	postsTable := `CREATE TABLE IF NOT EXISTS posts (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT UNSIGNED,
		caption TEXT,
		content_url VARCHAR(225),
		visibility ENUM('public', 'friends') DEFAULT 'friends',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

		CONSTRAINT ps_usr_id_fk
			FOREIGN KEY(user_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE
	)`
	_, err = DB.Exec(postsTable)
	if err != nil {
		panic("Failed to create posts table: " + err.Error())
	}

	likesTable := `CREATE TABLE IF NOT EXISTS likes (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		post_id BIGINT UNSIGNED,
		user_id BIGINT UNSIGNED,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		CONSTRAINT lk_pst_id_fk
			FOREIGN KEY(post_id)
			REFERENCES posts (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT lk_usr_id_fk
			FOREIGN KEY(user_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT lk_post_user_uq UNIQUE (post_id, user_id)
	)`
	_, err = DB.Exec(likesTable)
	if err != nil {
		panic("Failed to create likes table: " + err.Error())
	}

	commentsTable := `CREATE TABLE IF NOT EXISTS comments (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		post_id BIGINT UNSIGNED,
		user_id BIGINT UNSIGNED,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		CONSTRAINT cm_pst_id_fk
			FOREIGN KEY(post_id)
			REFERENCES posts (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT cm_usr_id_fk
			FOREIGN KEY(user_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE
	)`
	_, err = DB.Exec(commentsTable)
	if err != nil {
		panic("Failed to create comments table: " + err.Error())
	}

	notificationsTable := `CREATE TABLE IF NOT EXISTS notifications (
		id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
		recipient_id BIGINT UNSIGNED,
		actor_id BIGINT UNSIGNED,
		type ENUM('like','comment','friend_request','friend_accept') NOT NULL,
		reference_id BIGINT UNSIGNED,
		is_read BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

		CONSTRAINT nt_rec_id_fk
			FOREIGN KEY(recipient_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE,

		CONSTRAINT nt_act_id_fk
			FOREIGN KEY(actor_id)
			REFERENCES users (id)
			ON DELETE CASCADE
			ON UPDATE CASCADE
	)`
	_, err = DB.Exec(notificationsTable)
	if err != nil {
		panic("Failed to create notifications table: " + err.Error())
	}
}
