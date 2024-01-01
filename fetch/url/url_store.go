package url

import "sync"

type Store struct {
	rmu  sync.RWMutex
	urls map[string]struct{}
}

func InitUrlStore() *Store {
	return &Store{
		rmu:  sync.RWMutex{},
		urls: make(map[string]struct{}),
	}
}

func (us *Store) Contains(url string) bool {
	us.rmu.RLock()
	defer us.rmu.RUnlock()
	_, ok := us.urls[url]
	return ok
}

func (us *Store) Add(url string) {
	us.rmu.Lock()
	us.urls[url] = struct{}{}
	us.rmu.Unlock()
}
