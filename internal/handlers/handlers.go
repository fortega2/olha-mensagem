package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	failedEncodeuserDataErrMsg = "Failed to encode user data"
)

type handler struct {
	logger  logger.Logger
	queries *repository.Queries
}

func NewHandler(l logger.Logger, q *repository.Queries) *handler {
	return &handler{
		logger:  l,
		queries: q,
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

	setContentTypeJSON(w)
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

	setContentTypeJSON(w)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, failedEncodeuserDataErrMsg, http.StatusInternalServerError)
		return
	}

	h.logger.Info("User data retrieved", "username", user.Username, "userID", user.ID)
}

func (h *handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	failedToCreateUserErrMsg := "Failed to create user"

	type createUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.logger.Error("Username and password cannot be empty")
		http.Error(w, "Username and password cannot be empty", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		http.Error(w, failedToCreateUserErrMsg, http.StatusInternalServerError)
		return
	}

	params := repository.CreateUserParams{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	user, err := h.queries.CreateUser(context.Background(), params)
	if err != nil {
		h.logger.Error(failedToCreateUserErrMsg, "error", err)
		http.Error(w, failedToCreateUserErrMsg, http.StatusInternalServerError)
		return
	}

	userDto := dto.NewUserDTO(user.ID, user.Username)
	setContentTypeJSON(w)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(userDto); err != nil {
		h.logger.Error(failedEncodeuserDataErrMsg, "error", err)
		http.Error(w, failedEncodeuserDataErrMsg, http.StatusInternalServerError)
		return
	}

	h.logger.Info("User created successfully", "username", user.Username, "userID", user.ID)
}

func setContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
