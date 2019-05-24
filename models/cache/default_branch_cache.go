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

// Set default branch to cache
func SetDefaultBranch(project string, branch string) {
	defaultBranchCache[project] = branch
}
