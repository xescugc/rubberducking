package game

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func (g *Game) startHttpServer(p string, v bool) {

	r := mux.NewRouter()

	r.HandleFunc("/messages", g.createMessageHandler).Methods(http.MethodPost)

	hmux := http.NewServeMux()
	hmux.Handle("/", r)

	out := io.Discard
	if v {
		out = os.Stdout
	}
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", p),
		Handler: handlers.LoggingHandler(out, hmux),
	}

	if err := svr.ListenAndServe(); err != nil {
		panic(fmt.Errorf("server error: %w", err).Error())
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type CreateMessageRequest struct {
	Message string `json:"message"`
}

func (g *Game) createMessageHandler(w http.ResponseWriter, r *http.Request) {
	var b CreateMessageRequest
	err := json.NewDecoder(r.Body).Decode(&b)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	g.AD.AddMessage(b.Message)
}
