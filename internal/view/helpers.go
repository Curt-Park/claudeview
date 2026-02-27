package view

// ShortID returns the first n characters of id, or the full id if shorter.
func ShortID(id string, n int) string {
	if len(id) > n {
		return id[:n]
	}
	return id
}

func truncateHash(hash string) string {
	// Project hashes can be long path-encoded strings; truncate for display
	if len(hash) > 50 {
		return "…" + hash[len(hash)-49:]
	}
	return hash
}
