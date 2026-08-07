package databases

func createIndexOnTable() {
	usersIndex := `CREATE INDEX IF NOT EXISTS users_table_index
		ON users (username, email)`
	_, err := DB.Exec(usersIndex)
	if err != nil {
		panic("Failed to create index on users table: " + err.Error())
	}

	friendshipsIndex := `CREATE INDEX IF NOT EXISTS friendships_table_index
		ON friendships (status)`
	_, err = DB.Exec(friendshipsIndex)
	if err != nil {
		panic("Failed to create index on friendships table: " + err.Error())
	}

	postsIndex := `CREATE INDEX IF NOT EXISTS post_table_index 
		ON posts (user_id, created_at)`
	_, err = DB.Exec(postsIndex)
	if err != nil {
		panic("Failed to create index on posts table: " + err.Error())
	}

	likesIndex := `CREATE INDEX IF NOT EXISTS likes_table_index
		ON likes (post_id)`
	_, err = DB.Exec(likesIndex)
	if err != nil {
		panic("Failed to create index on likes table: " + err.Error())
	}

	commentsIndex := `CREATE INDEX IF NOT EXISTS comments_table_index
		ON comments (post_id, user_id)`
	_, err = DB.Exec(commentsIndex)
	if err != nil {
		panic("Failed to create index on comments table: " + err.Error())
	}

	notificationsIndex := `CREATE INDEX IF NOT EXISTS notifications_table_index
		ON notifications (reference_id, is_read)`
	_, err = DB.Exec(notificationsIndex)
	if err != nil {
		panic("Failed to create index on notifications table: " + err.Error())
	}
}
