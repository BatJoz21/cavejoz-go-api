package databases

func createIndexOnTable() {
	usersIndex := `CREATE INDEX IF NOT EXISTS users_email_index
		ON users (email)`
	_, err := DB.Exec(usersIndex)
	if err != nil {
		panic("Failed to create index on users table: " + err.Error())
	}

	postIndex := `CREATE INDEX IF NOT EXISTS post_table_index 
		ON posts (user_id, created_at)`
	_, err = DB.Exec(postIndex)
	if err != nil {
		panic("Failed to create index on posts table: " + err.Error())
	}
}
