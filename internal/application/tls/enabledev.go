package tls

import (
	"github.com/fanky5g/ponzu/internal/config"
	"log"
	"net/http"
	"path/filepath"
)

// EnableDev generates self-signed SSL certificates to use HTTPS & HTTP/2 while
// working in a development environment. The certs are saved in a different
// directory than the production certs (from Let's Encrypt), so that the
// acme/autocert package doesn't mistake them for it's own.
// Additionally, a TLS server is started using the default http mux.
func (s *service) EnableDev() {
	s.setupDev()

	vendorPath := config.TlsDir()
	cert := filepath.Join(vendorPath, "devcerts", "cert.pem")
	key := filepath.Join(vendorPath, "devcerts", "key.pem")

	log.Fatalln(http.ListenAndServeTLS(":10443", cert, key, nil))
}
