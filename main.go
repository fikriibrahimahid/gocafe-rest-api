package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sync"
)

var (
	listUsersRe  = regexp.MustCompile(`^\/users[\/]*$`)
	getUserRe    = regexp.MustCompile(`^\/users\/(\d+)$`)
	createUserRe = regexp.MustCompile(`^\/users[\/]*$`)
)

type user struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type datastore struct {
	m map[string]user
	*sync.RWMutex
}

type userHandler struct {
	store *datastore
}

func (h *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	switch {
	case r.Method == http.MethodGet && listUsersRe.MatchString(r.URL.Path):
		h.List(w, r)
		return
	case r.Method == http.MethodGet && getUserRe.MatchString(r.URL.Path):
		h.Get(w, r)
		return
	case r.Method == http.MethodPost && createUserRe.MatchString(r.URL.Path):
		h.Create(w, r)
		return
	default:
		notFound(w, r)
		return
	}
}

func (h *userHandler) List(w http.ResponseWriter, r *http.Request) {
	users := make([]user, 0, len(h.store.m))
	h.store.RLock()
	for _, u := range h.store.m {
		users = append(users, u)
	}
	h.store.RUnlock()
	jsonBytes, err := json.Marshal(users)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *userHandler) Get(w http.ResponseWriter, r *http.Request) {
	matches := getUserRe.FindStringSubmatch(r.URL.Path)
	if len(matches) < 2 {
		notFound(w, r)
	}
	h.store.RLock()
	user, ok := h.store.m[matches[1]]
	h.store.RUnlock()
	if !ok {
		notFound(w, r)
		return
	}
	jsonBytes, err := json.Marshal(user)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *userHandler) Create(w http.ResponseWriter, r *http.Request) {
	u := user{}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		badRequest(w, r)
		return
	}
	h.store.Lock()
	h.store.m[u.ID] = u
	h.store.Unlock()
	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func badRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"error" : "bad request"}`))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error" : "not found"}`))
}

func internalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error" : "interval server error"}`))
}

// .
// .
//.
// .
// .

func main() {
	mux := http.NewServeMux()
	userH := &userHandler{
		store: &datastore{
			m: map[string]user{
				"1":  {ID: "MCI001", Name: "Kevin De Bruyne"},
				"2":  {ID: "MCI002", Name: "Bernardo Silva"},
				"3":  {ID: "MCI003", Name: "Erling Braut Haaland"},
				"4":  {ID: "MCI004", Name: "Ederson Moraes"},
				"5":  {ID: "MCI005", Name: "Jack Grealish"},
				"6":  {ID: "MCI006", Name: "Kyle Walker"},
				"7":  {ID: "MCI007", Name: "Joao Cancelo"},
				"8":  {ID: "MCI008", Name: "Ruben Dias"},
				"9":  {ID: "MCI009", Name: "Aymeric Laporte"},
				"10": {ID: "MCI010", Name: "John Stones"},
				"11": {ID: "MCI011", Name: "Manuel Akanji"},
				"12": {ID: "MCI012", Name: "Ilkay Gundogan"},
				"13": {ID: "MCI013", Name: "Phil Foden"},
				"14": {ID: "MCI014", Name: "Riyad Mahrez"},
			},
			RWMutex: &sync.RWMutex{},
		},
	}
	mux.Handle("/users/", userH)
	mux.Handle("/users", userH)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
