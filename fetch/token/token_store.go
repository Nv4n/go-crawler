package token

var tokenStore chan struct{}

func InitTokenStore(goroutines uint) {
	tokenStore = make(chan struct{}, goroutines)
}

func GetReadTokenChan() <-chan struct{} {
	return tokenStore
}

func GetWriteTokenChan() chan<- struct{} {
	return tokenStore
}
