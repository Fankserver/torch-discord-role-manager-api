package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	httpInterface = flag.String("interface", "", "http listening interface")
	httpPort      = flag.Int("port", 8080, "http listening port")
	password      = flag.String("password", "", "http header password")
)

func main() {
	flag.Parse()

	// subscribe to SIGINT signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	s, err := newStorage()
	if err != nil {
		logrus.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			DiscordTag string `json:"discord_tag"`
			SteamID    int64  `json:"steam_id"`
		}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.Add(request.SteamID, request.DiscordTag)
	}).Methods(http.MethodPost)
	router.HandleFunc("/steamid/{steam_id:[0-9]+}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		i, err := strconv.ParseInt(vars["steam_id"], 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			DiscordTag string `json:"discord_tag"`
		}{
			DiscordTag: s.GetDiscordTag(i),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *httpInterface, *httpPort),
		Handler: authenticate(router),
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
		}
	}()
	<-quit
	logrus.Println("shutting down server...")
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxTimeout); err != nil {
		logrus.Fatalf("could not shutdown: %v", err)
	}
	logrus.Println("server gracefully stopped")
}

func authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != *password {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
