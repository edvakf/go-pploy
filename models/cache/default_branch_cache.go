package cache

var defaultBranchCache = make(map[string]string)

// Get cached default branch
func GetDefaultBranch(project string) string {
	branch, ok := defaultBranchCache[project]

	if !ok {
		return ""
	}

	return branch
}

// SetDefaultBranch caches the default branch name.
func SetDefaultBranch(project string, branch string) {
	defaultBranchCache[project] = branch
}

// DeleteDefaultBranch deletes the default branch cache entry.
func DeleteDefaultBranch(project string) {
	delete(defaultBranchCache, project)
}
