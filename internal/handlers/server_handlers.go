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

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	failedToCreateUserErrMsg := "Failed to create user"

	type createUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, invalidRequestBodyErrMsg, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.logger.Error(usernameAndPasswordEmptyErrMsg)
		http.Error(w, usernameAndPasswordEmptyErrMsg, http.StatusBadRequest)
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

	user, err := h.queries.CreateUser(r.Context(), params)
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

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	type loginUserRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, invalidRequestBodyErrMsg, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		h.logger.Error(usernameAndPasswordEmptyErrMsg)
		http.Error(w, usernameAndPasswordEmptyErrMsg, http.StatusBadRequest)
		return
	}

	user, err := h.queries.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		h.logger.Error("Failed to retrieve user", "error", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		h.logger.Error("Invalid password", "username", req.Username)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	userDto := dto.NewUserDTO(user.ID, user.Username)
	setContentTypeJSON(w)
	if err := json.NewEncoder(w).Encode(userDto); err != nil {
		h.logger.Error(failedEncodeuserDataErrMsg, "error", err)
		http.Error(w, failedEncodeuserDataErrMsg, http.StatusInternalServerError)
		return
	}

	h.logger.Info("User logged in successfully", "username", user.Username, "userID", user.ID)
}

func setContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}
