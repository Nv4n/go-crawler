package store

import "sync"

type UrlStore struct {
	rmu  sync.RWMutex
	urls map[string]struct{}
}

func InitUrlStore() *UrlStore {
	return &UrlStore{
		rmu:  sync.RWMutex{},
		urls: make(map[string]struct{}),
	}
}

func (us *UrlStore) Contains(url string) bool {
	us.rmu.RLock()
	defer us.rmu.RUnlock()
	_, ok := us.urls[url]
	return ok
}

func (us *UrlStore) Add(url string) {
	us.rmu.Lock()
	us.urls[url] = struct{}{}
	us.rmu.Unlock()
}
