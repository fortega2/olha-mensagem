package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
)

const (
	failedEncodeuserDataErrMsg = "Failed to encode user data"
)

type handler struct {
	logger logger.Logger
}

func NewHandler(l logger.Logger) *handler {
	return &handler{
		logger: l,
	}
}

func (h *handler) Root(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "internal/templates/index.html")
}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}

	user := websocket.NewUser(requestData.Username)
	websocket.AddUser(user)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		h.logger.Error(failedEncodeuserDataErrMsg, "error", err)
		http.Error(w, failedEncodeuserDataErrMsg, http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged in", "username", user.Username, "userID", user.ID)
}

func (h *handler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	if idStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user := websocket.GetUserByID(id)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, failedEncodeuserDataErrMsg, http.StatusInternalServerError)
		return
	}

	h.logger.Info("User data retrieved", "username", user.Username, "userID", user.ID)
}
