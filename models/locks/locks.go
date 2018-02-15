package locks

import (
	"errors"
	"sync"
	"time"

	"github.com/edvakf/go-pploy/models"
	"github.com/edvakf/go-pploy/models/hook"
)

type lock models.Lock

func (l *lock) valid(now time.Time) bool {
	return l.EndTime.After(now)
}

func (l *lock) by(user string) bool {
	return l.User == user
}

// map of project name to lock
var locks = make(map[string]lock)

var mu sync.Mutex

var lockDuration = 20 * time.Minute

// Check returns lock of a project
func Check(project string, now time.Time) *models.Lock {
	mu.Lock()
	defer mu.Unlock()

	l, ok := locks[project]
	if ok && l.valid(now) {
		return (*models.Lock)(&l)
	}
	return nil
}

// Gain lets a user to gain lock for a project
// returns error when lock is taken by others
// if the user already has gained lock for the project, then re-set the expiration time
func Gain(project string, user string, now time.Time) (*models.Lock, error) {
	mu.Lock()
	defer mu.Unlock()

	l, ok := locks[project]
	if ok && l.valid(now) && !l.by(user) {
		return nil, errors.New("lock is already taken by someone else")
	}
	l = lock{User: user, EndTime: now.Add(lockDuration)}
	locks[project] = l
	hook.LockGained(project, user)
	return (*models.Lock)(&l), nil
}

// Extend adds the duration to the lock
// returns error when the user does not have lock for the project
func Extend(project string, user string, now time.Time) (*models.Lock, error) {
	mu.Lock()
	defer mu.Unlock()

	l, ok := locks[project]
	if !ok || !l.valid(now) || !l.by(user) {
		return nil, errors.New("user does not have lock for the project")
	}
	// l.EndTime = l.EndTime.Add(lockDuration)
	l = lock{User: user, EndTime: l.EndTime.Add(lockDuration)}
	locks[project] = l
	hook.LockExtended(project, user)
	return (*models.Lock)(&l), nil
}

// Release unsets lock for a project
// returns error when the user does not have lock for the project
func Release(project string, user string, now time.Time) error {
	mu.Lock()
	defer mu.Unlock()

	l, ok := locks[project]
	if !ok || !l.valid(now) || !l.by(user) {
		return errors.New("user does not have lock for the project")
	}
	delete(locks, project)
	hook.LockReleased(project, user)
	return nil
}

// SetDuration overrides the duration to take lock
func SetDuration(dur time.Duration) {
	lockDuration = dur
}
