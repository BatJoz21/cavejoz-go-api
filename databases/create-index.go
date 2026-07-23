package databases

func createIndexOnTable() {
	postIndex := `CREATE INDEX IF NOT EXISTS post_table_index 
		ON posts (user_id, created_at)`
	_, err := DB.Exec(postIndex)
	if err != nil {
		panic("Failed to create index on posts table: " + err.Error())
	}
}
