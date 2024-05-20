package handler

import (
	"courses/internal/core"
	"courses/pkg/words"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// TODO: handle errors on fprint

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var u User
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		// TODO
		return
	}

	role, err := auth(u)
	if err == nil {
		u.role = role
		tokenString, err := createToken(u)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.logger.Error("No username found")
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, tokenString)
		return
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid credentials")
	}
}

func (s *Server) protectHandler(next func(http.ResponseWriter, *http.Request), checkForAdmin bool,
	w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Missing authorization header")
		return
	}

	isAdmin, err := verifyToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}

	if checkForAdmin && !isAdmin {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, "Only admin can update comics set")
		return
	}

	next(w, r)
}

func (s *Server) searchRequest(wr http.ResponseWriter, r *http.Request) {
	comicsKeywords := r.URL.Query().Get("search")
	// TODO: validate
	clearedKeywords := words.StemStringWithClearing(strings.Split(comicsKeywords, " "))
	res := s.ctlg.FindByIndex(clearedKeywords)

	if len(res) == 0 {
		wr.WriteHeader(404)
		_, err := wr.Write([]byte("404 not found"))
		if err != nil {
			s.logger.Error("writing response for GET /pics", "err", err)
		}
		return
	}

	comicsToSend := min(len(res), core.MaxComicsToShow)
	data, err := json.Marshal(res[:comicsToSend])
	if err != nil {
		s.logger.Error("marshalling res of catalog", "err", err)
		data = []byte("")
	}

	_, err = wr.Write(data)
	if err != nil {
		s.logger.Error("writing response for GET /pics", "err", err)
	}
}

func (s *Server) protectedSearch() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.protectHandler(s.searchRequest, false, w, r)
	}
}

func (s *Server) updateRequest(wr http.ResponseWriter, _ *http.Request) {
	diff, err := s.ctlg.UpdateComics()
	if err != nil {
		s.logger.Error("updating comics", "err", err)
	}
	data, err := json.Marshal(diff)
	if err != nil {
		s.logger.Error("marshalling diff of comics update", "err", err)
		data = []byte("")
	}
	_, err = wr.Write(data)
	if err != nil {
		s.logger.Error("writing response for POST /update", "err", err)
	}
}

func (s *Server) protectedUpdate() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		s.protectHandler(s.updateRequest, true, w, r)
	}
}
