package view

func truncateHash(hash string) string {
	// Project hashes can be long path-encoded strings; truncate for display
	if len(hash) > 50 {
		return "…" + hash[len(hash)-49:]
	}
	return hash
}
