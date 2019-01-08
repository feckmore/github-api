package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

type datastore struct {
	Context context.Context
	Client  *github.Client
	Service *github.GitService
}

const (
	port = 5000
)

func main() {
	data, err := New(os.Getenv("TOKEN"))
	if err != nil || data == nil || data.Client == nil {
		log.Fatal("Invalid Github client:", err)
	}
	router := NewRouter(data)

	// serve on specified port
	p := fmt.Sprintf(":%v", port)
	log.Println("listening on port", p)
	log.Fatal(http.ListenAndServe(p, router))
}

// NewRouter accepts a content.Service interface and returns the router/handler for content endpoints
func NewRouter(data *datastore) http.Handler {
	r := mux.NewRouter()

	r.Methods("GET").Path("/{owner}/repos/count").Handler(GetCount(data))

	return r
}

// New function, initiates and returns a Github datastore instance
func New(authToken string) (*datastore, error) {
	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: authToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	if tc == nil {
		return nil, errors.New("Access Token Invalid")
	}

	client := github.NewClient(tc)
	if client == nil {
		return nil, errors.New("Error creating Github client")
	}

	return &datastore{
		Context: ctx,
		Client:  client,
		Service: client.Git,
	}, nil
}

func GetCount(data *datastore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		owner := vars["owner"]

		repos, _, err := data.Client.Repositories.List(data.Context, owner, nil)
		if WriteError(w, err) {
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(w).Encode(len(repos))
		WriteError(w, err)
	}
}

func WriteError(w http.ResponseWriter, err error) bool {
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return true
	}
	return false
}
