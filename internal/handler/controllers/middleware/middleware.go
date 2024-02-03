package middleware

import (
	"log"
	"net/http"
)

type Token string

type Middleware func(next http.HandlerFunc) http.HandlerFunc
type Middlewares map[Token]Middleware

func (middlewares Middlewares) Get(token Token) Middleware {
	if middleware, ok := middlewares[token]; ok {
		return middleware
	}

	log.Fatalf("Middleware %s is not implemented", token)
	return nil
}
