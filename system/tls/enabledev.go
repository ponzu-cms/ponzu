package tls

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// EnableDev generates self-signed SSL certificates to use HTTPS & HTTP/2 while
// working in a development environment. The certs are saved in a different
// directory than the production certs (from Let's Encrypt), so that the
// acme/autocert package doesn't mistake them for it's own.
// Additionally, a TLS server is started using the default http mux.
func EnableDev() {
	setupDev()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Couldn't find working directory to activate dev certificates:", err)
	}

	vendorPath := filepath.Join(pwd, "cmd", "ponzu", "vendor", "github.com", "ponzu-cms", "ponzu", "system", "tls")

	cert := filepath.Join(vendorPath, "devcerts", "cert.pem")
	key := filepath.Join(vendorPath, "devcerts", "key.pem")

	go log.Fatalln(http.ListenAndServeTLS(":10443", cert, key, nil))
	fmt.Println("Server listening on https://localhost:10443 for requests... [DEV]")
	fmt.Println("----")
	fmt.Println("If your browser rejects HTTPS requests, try allowing insecure connections on localhost.")
	fmt.Println("on Chrome, visit chrome://flags/#allow-insecure-localhost")
}
