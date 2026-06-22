// Package controller exposes user HTTP + WebSocket handlers (gorilla).
package controller

import (
	"encoding/json"
	"net/http"

	"archview-example-gorilla-ws/internal/user/service"

	"github.com/gorilla/websocket"
)

type UserController struct {
	svc      service.UserService
	upgrader websocket.Upgrader
}

func New(svc service.UserService) *UserController { return &UserController{svc: svc} }

// List handles GET /users.
func (c *UserController) List(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(c.svc.ListUsers())
}

// Create handles POST /users.
func (c *UserController) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(c.svc.CreateUser(body.Name))
}

// Stream handles GET /ws — a WebSocket upgrade, echoing the user list.
func (c *UserController) Stream(w http.ResponseWriter, r *http.Request) {
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		mt, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
		payload, _ := json.Marshal(c.svc.ListUsers())
		if conn.WriteMessage(mt, payload) != nil {
			break
		}
	}
}
