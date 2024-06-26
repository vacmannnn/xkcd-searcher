package xkcd

import (
	"courses/core"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

type server struct {
	ctlg         core.Catalog
	logger       slog.Logger
	mux          *http.ServeMux
	users        []userInfo
	tokenMaxTime int
	requests     chan struct{}
}

func NewServerHandler(ctlg core.Catalog, logger slog.Logger, pathToUsers string,
	rateLimit, concurrencyLimit, tokenMaxTime int) http.Handler {
	users, err := getUsers(pathToUsers)
	if err != nil {
		logger.Error("Failed to load users", "error", err.Error())
	}
	myServ := server{ctlg: ctlg, logger: logger, mux: http.NewServeMux(), users: users, tokenMaxTime: tokenMaxTime,
		requests: make(chan struct{}, concurrencyLimit)}

	myServ.mux.HandleFunc("GET /pics", myServ.protectedSearch())
	myServ.mux.HandleFunc("POST /update", myServ.protectedUpdate())
	myServ.mux.HandleFunc("POST /login", myServ.login)

	return limit(myServ.mux, rateLimit)
}

const (
	admin = iota
	user
)

type userInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	role     int
}

// TODO: хранить не тут
func getUsers(path string) ([]userInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var myUsers []struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
		Role     string `json:"role"`
	}

	err = json.Unmarshal(data, &myUsers)
	if err != nil {
		return nil, err
	}

	users := make([]userInfo, len(myUsers))
	for i, u := range myUsers {
		users[i].Username = u.Username
		users[i].Password = u.Password
		if u.Role == "admin" {
			users[i].role = admin
		} else {
			users[i].role = user
		}
	}
	return users, nil
}
