package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	defaultPort    = 8080
	requestTimeout = 3 * time.Second
	shutdownGrace  = 10 * time.Second
)

type serverResponse struct {
	Headers map[string][]string `json:"Headers"`
	Params  map[string][]string `json:"Params"`
	Body    any                 `json:"Body"`
	Path    string              `json:"Path"`
}

func decodeBody(raw []byte) any {
	if len(raw) == 0 {
		return ""
	}
	if json.Valid(raw) {
		return json.RawMessage(raw)
	}
	return string(raw)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("write response: %v", err)
	}
}

func globalHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not read request body"})
		return
	}

	headers := map[string][]string(r.Header.Clone())
	if headers == nil {
		headers = map[string][]string{}
	}
	params := map[string][]string(r.URL.Query())
	if params == nil {
		params = map[string][]string{}
	}

	writeJSON(w, http.StatusOK, serverResponse{
		Headers: headers,
		Params:  params,
		Body:    decodeBody(raw),
		Path:    r.URL.Path,
	})
}

func newRouter() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger, middleware.Recoverer)
	router.Handle("/*", http.HandlerFunc(globalHandler))
	router.NotFound(globalHandler)
	router.MethodNotAllowed(globalHandler)

	return router
}

func newServer(addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           newRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       requestTimeout,
		WriteTimeout:      requestTimeout,
		IdleTimeout:       60 * time.Second,
	}
}

func resolvePort() int {
	port := defaultPort
	if env := os.Getenv("PORT"); env != "" {
		parsed, err := strconv.Atoi(env)
		if err != nil {
			log.Fatalf("invalid PORT value %q: must be an integer", env)
		}
		port = parsed
	}

	return port
}

func runServer(srv *http.Server, errCh chan error) {
	log.Printf("http server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- err
	}
}

func main() {
	port := resolvePort()
	srv := newServer(net.JoinHostPort("", strconv.Itoa(port)))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go runServer(srv, errCh)

	select {
	case err := <-errCh:
		log.Fatalf("server error: %v", err)
	case <-ctx.Done():
		log.Print("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
}
