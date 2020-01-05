package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"path/filepath"

	"github.com/albertodonato/h2static/version"
)

// StaticServer is a static HTTP server.
type StaticServer struct {
	Addr                    string
	Dir                     string
	DisableH2               bool
	DisableLookupWithSuffix bool
	ShowDotFiles            bool
	Log                     bool
	PasswordFile            string
	TLSCert                 string
	TLSKey                  string
}

// IsHTTPS returns whether HTTPS is enabled.
func (s StaticServer) IsHTTPS() bool {
	return s.TLSCert != "" && s.TLSKey != ""
}

// getServer returns a configured server.
func (s StaticServer) getServer() *http.Server {
	fileSystem := NewFileSystem(
		s.Dir, !s.DisableLookupWithSuffix, !s.ShowDotFiles)
	mux := http.NewServeMux()
	mux.Handle("/", NewFileHandler(fileSystem))
	mux.Handle(
		AssetsPrefix,
		http.StripPrefix(AssetsPrefix, &AssetsHandler{Assets: staticAssets}))

	var handler http.Handler = mux

	if s.PasswordFile != "" {
		credentials, err := loadCredentials(s.PasswordFile)
		if err != nil {
			panic(err)
		}
		handler = &BasicAuthHandler{
			Handler:     handler,
			Credentials: credentials,
			Realm:       version.App.Name,
		}
	}

	if s.Log {
		handler = &LoggingHandler{Handler: handler}
	}
	handler = &CommonHeadersHandler{Handler: handler}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !s.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	return &http.Server{
		Addr:         s.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}
}

// Run starts the server.
func (s StaticServer) Run() error {
	var err error

	server := s.getServer()
	isHTTPS := s.IsHTTPS()
	if s.Log {
		kind := "HTTP"
		if isHTTPS {
			kind = "HTTPS"
		}
		absPath, err := filepath.Abs(s.Dir)
		if err != nil {
			return err
		}
		log.Printf("Starting %s server on %s, serving path %s", kind, s.Addr, absPath)
	}

	if isHTTPS {
		err = server.ListenAndServeTLS(s.TLSCert, s.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	return err
}
