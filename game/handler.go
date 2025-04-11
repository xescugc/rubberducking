package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xescugc/rubberducking/src"
)

func (g *Game) startHttpServer(p string, v bool) {
	r := mux.NewRouter()

	r.HandleFunc("/messages", g.createMessageHandler).Methods(http.MethodPost)
	r.HandleFunc("/toggle", g.toggleHandler).Methods(http.MethodPost)

	hmux := http.NewServeMux()
	hmux.Handle("/", r)

	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", p),
		Handler: handlers.LoggingHandler(logWriter, hmux),
	}

	Logger.Info("Starting server", "port", p)
	if err := svr.ListenAndServe(); err != nil {
		Logger.Error("Server exited with err", "error", err)
		os.Exit(1)
	} else {
		Logger.Error("Server exited without err")
	}
}

func (g *Game) createMessageHandler(w http.ResponseWriter, r *http.Request) {
	var b src.CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		json.NewEncoder(w).Encode(src.ErrorResponse{Error: err.Error()})
		return
	}
	Logger.Info("Sending message to manager", "msg", b.Message)

	g.AD.AddMessage(b.Message)
}

func (g *Game) toggleHandler(w http.ResponseWriter, r *http.Request) {
	g.AD.Toggle()
}
