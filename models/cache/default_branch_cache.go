package cache

var defaultBranchCache = make(map[string]string)

// GetDefaultBranch returns the default branch name in cache.
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
