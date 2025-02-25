package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/edulinq/autograder/internal/api/core"
	"github.com/edulinq/autograder/internal/config"
	"github.com/edulinq/autograder/internal/lockmanager"
	"github.com/edulinq/autograder/internal/log"
)

var apiServer *http.Server = nil
var httpRedirectServer *http.Server = nil

const API_SERVER_LOCK = "internal.api.server.API_SERVER_LOCK"

func runAPIServer(routes *[]core.Route, subserverSetupWaitGroup *sync.WaitGroup) (err error) {
	defer func() {
		value := recover()
		if value == nil {
			return
		}

		err = errors.Join(err, fmt.Errorf("API server panicked: '%v'.", value))
	}()

	// Unlock API_SERVER_LOCK explicitly on each code path to ensure proper release regardless of the outcome.
	lockmanager.Lock(API_SERVER_LOCK)
	if apiServer != nil {
		subserverSetupWaitGroup.Done()
		lockmanager.Unlock(API_SERVER_LOCK)
		return fmt.Errorf("API server is already running.")
	}

	err = createAPIServer(routes)
	subserverSetupWaitGroup.Done()
	if err != nil {
		lockmanager.Unlock(API_SERVER_LOCK)
		return fmt.Errorf("Failed to create API server: '%w'.", err)
	}

	lockmanager.Unlock(API_SERVER_LOCK)

	runAPIServerInternal()
	if err != nil {
		log.Error("API server returned an error.", err)
	}

	log.Info("API Server Stopped.")

	return err
}

func createAPIServer(routes *[]core.Route) error {
	httpPort := config.WEB_HTTP_PORT.Get()
	httpRedirect := config.WEB_HTTP_REDIRECT.Get()

	httpsPort := config.WEB_HTTPS_PORT.Get()
	httpsCertPath := config.WEB_HTTPS_CERT.Get()
	httpsKeyPath := config.WEB_HTTPS_KEY.Get()
	httpsEnable := config.WEB_HTTPS_ENABLE.Get()

	mainPort := httpPort
	var tlsConfig *tls.Config = nil

	logAttrs := []any{
		log.NewAttr("https", httpsEnable),
	}

	// Fetch SSL certs.
	if httpsEnable {
		if (httpsCertPath == "") || (httpsKeyPath == "") {
			return fmt.Errorf("HTTPS is enabled, but no certificate and/or key file were specified.")
		}

		cert, err := tls.LoadX509KeyPair(httpsCertPath, httpsKeyPath)
		if err != nil {
			return fmt.Errorf("Failed to load SSL cert/key: '%w'.", err)
		}

		mainPort = httpsPort
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	apiServer = &http.Server{
		Addr:      fmt.Sprintf(":%d", mainPort),
		Handler:   core.GetRouteServer(routes),
		TLSConfig: tlsConfig,
	}

	// Setup the HTTP redirect.
	if httpsEnable && httpRedirect {
		logAttrs = append(logAttrs, log.NewAttr("http-redirect", httpPort))

		httpRedirectServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", httpPort),
			Handler: http.HandlerFunc(httpRedirectHandler),
		}
	}

	logAttrs = append(logAttrs, log.NewAttr("port", mainPort))

	log.Info("API Server Created.", logAttrs...)

	return nil
}

func runAPIServerInternal() error {
	if httpRedirectServer != nil {
		go func() {
			err := httpRedirectServer.ListenAndServe()
			if (err != nil) && (err == http.ErrServerClosed) {
				log.Error("HTTP redirect returned an error: '%w'.", err)
			}
		}()
	}

	var err error = nil

	if config.WEB_HTTPS_ENABLE.Get() {
		err = apiServer.ListenAndServeTLS("", "")
	} else {
		err = apiServer.ListenAndServe()
	}

	// Set err to nil if the API server stopped due to a graceful shutdown.
	if err == http.ErrServerClosed {
		err = nil
	}

	return err
}

func httpRedirectHandler(response http.ResponseWriter, request *http.Request) {
	// Start with the request URL and replace components for HTTPS.
	url := request.URL

	// Append the port to the host.
	host, _, _ := net.SplitHostPort(request.Host)
	url.Host = net.JoinHostPort(host, fmt.Sprintf("%d", config.WEB_HTTPS_PORT.Get()))

	// Change the scheme.
	url.Scheme = "https"

	http.Redirect(response, request, url.String(), http.StatusPermanentRedirect)
}

func stopAPIServer() {
	lockmanager.Lock(API_SERVER_LOCK)
	defer lockmanager.Unlock(API_SERVER_LOCK)

	if httpRedirectServer != nil {
		err := httpRedirectServer.Shutdown(context.Background())
		if err != nil {
			log.Error("Failed to stop the HTTP redirect server.", err)
		}
	}

	if apiServer != nil {
		err := apiServer.Shutdown(context.Background())
		if err != nil {
			log.Error("Failed to stop the API server.", err)
		}
	}

	httpRedirectServer = nil
	apiServer = nil
}
