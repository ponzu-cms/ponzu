package application

import (
	"log"
)

type ServiceToken string
type Services map[ServiceToken]interface{}

func (services Services) Get(token ServiceToken) interface{} {
	if service, ok := services[token]; ok {
		return service
	}

	log.Fatalf("Service %s is not implemented", token)
	return nil
}
