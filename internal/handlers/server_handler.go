package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

const (
	failedEncodeuserDataErrMsg     = "Failed to encode user data"
	usernameAndPasswordEmptyErrMsg = "Username and password cannot be empty"
	invalidRequestBodyErrMsg       = "Invalid request body"
)

type userCreateLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req userCreateLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, invalidRequestBodyErrMsg, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Create user attempt", "username", req.Username, "password_provided", req.Password != "")

	if req.Username == "" || req.Password == "" {
		h.logger.Error(usernameAndPasswordEmptyErrMsg)
		http.Error(w, usernameAndPasswordEmptyErrMsg, http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err, "username", req.Username)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Password hashed successfully", "username", req.Username)

	params := repository.CreateUserParams{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	user, err := h.queries.CreateUser(r.Context(), params)
	if err != nil {
		h.logger.Error("Failed to create user", "error", err, "username", req.Username)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	userDto := dto.NewUserDTO(user.ID, user.Username)
	respondWithJSON(w, http.StatusCreated, userDto, failedEncodeuserDataErrMsg)

	h.logger.Info("User created successfully", "userID", user.ID, "username", user.Username)
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req userCreateLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, invalidRequestBodyErrMsg, http.StatusBadRequest)
		return
	}

	h.logger.Debug("Login attempt", "username", req.Username, "password_provided", req.Password != "")

	if req.Username == "" || req.Password == "" {
		h.logger.Error(usernameAndPasswordEmptyErrMsg)
		http.Error(w, usernameAndPasswordEmptyErrMsg, http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to retrieve user", "error", err, "username", req.Username)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	h.logger.Debug("User retrieved", "userID", user.ID, "username", user.Username)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.logger.Error("Invalid password", "username", req.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	userDto := dto.NewUserDTO(user.ID, user.Username)
	respondWithJSON(w, http.StatusOK, userDto, failedEncodeuserDataErrMsg)

	h.logger.Info("User logged in successfully", "username", user.Username, "userID", user.ID)
}
