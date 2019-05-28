package cache

import "sync"

type defaultBranchCache struct {
	sm sync.Map
}

// DefaultBranch represents cache of default branch
var DefaultBranch = &defaultBranchCache{}

// Load returns the default branch name in cache.
func (c *defaultBranchCache) Load(project string) string {
	branch, ok := c.sm.Load(project)

	if !ok {
		return ""
	}

	return branch.(string)
}

// Store caches the default branch name.
func (c *defaultBranchCache) Store(project string, branch string) {
	c.sm.Store(project, branch)
}

// Delete deletes the default branch cache entry.
func (c *defaultBranchCache) Delete(project string) {
	c.sm.Delete(project)
}
