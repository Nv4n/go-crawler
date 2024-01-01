package crawl

import "github.com/nv4n/go-crawler/fetch/url"

var PageStore *url.Store

func InitPageStore() {
	PageStore = url.InitUrlStore()
}
