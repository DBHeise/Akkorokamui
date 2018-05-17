package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var (
	dir      string
	host     string
	port     string
	proxy    string
	loglevel string
	logfile  string
	certfile string
	keyfile  string
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Howdy!"))
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	flag.StringVar(&dir, "dir", ".", "the directory to serve files from.")
	flag.StringVar(&host, "host", "localhost", "the host to listen with/on")
	flag.StringVar(&port, "port", "13780", "the port to listen on")
	flag.StringVar(&proxy, "proxy", "", "A http proxy to use for output traffic")
	flag.StringVar(&loglevel, "loglevel", "info", "Level of debugging {debug|info|warn|error|panic}")
	flag.StringVar(&logfile, "logfile", "/var/log/easyscan.log", "Location of log file")
	flag.StringVar(&certfile, "cert", "", "certificate file")
	flag.StringVar(&keyfile, "key", "", "private key file")
}

func main() {
	log.Debug("Main Starting")

	flag.Parse()

	logL, err := log.ParseLevel(loglevel)
	if err != nil {
		log.Warn("Unable to parse loglevel, setting to default: info")
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(logL)
	}
	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.SetOutput(os.Stdout)
		log.Warn("Error opening logfile: " + err.Error())
	} else {
		log.SetOutput(file)
	}

	r := mux.NewRouter()
	//r.HandleFunc("/", defaultHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(dir)))

	config := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	srv := &http.Server{
		Handler:      r,
		Addr:         host + ":" + port,
		WriteTimeout: 10 * time.Minute,
		ReadTimeout:  5 * time.Minute,
		TLSConfig:    config,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}
	if certfile == "" || keyfile == "" {

		log.Fatal(srv.ListenAndServe())
	} else {

		log.Fatal(srv.ListenAndServeTLS(certfile, keyfile))
	}
}
