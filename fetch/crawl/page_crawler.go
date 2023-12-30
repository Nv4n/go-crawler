package crawl

import "github.com/nv4n/go-crawler/fetch/store"

var PageStore *store.UrlStore

func InitPageStore() {
	PageStore := store.InitUrlStore()
}
