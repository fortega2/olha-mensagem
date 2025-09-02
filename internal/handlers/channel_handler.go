package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fortega2/real-time-chat/internal/repository"
)

const (
	failedEncodeChannelDataErrMsg string = "Failed to encode channel data"
)

type createChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int64  `json:"userId"`
}

func (ccr createChannelRequest) isValid() bool {
	return ccr.Name != "" && ccr.UserID != 0
}

type channelResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"createdBy"`
	CreatedAt   string `json:"createdAt"`
}

func newChannelResponse(channel repository.Channel) channelResponse {
	var description string
	if channel.Description.Valid {
		description = channel.Description.String
	} else {
		description = ""
	}

	return channelResponse{
		ID:          channel.ID,
		Name:        channel.Name,
		Description: description,
		CreatedBy:   channel.CreatedBy,
		CreatedAt:   channel.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) GetAllChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if ctx.Err() != nil {
		h.logger.Error("Request context error", "error", ctx.Err())
		http.Error(w, "Request cancelled or timed out", http.StatusRequestTimeout)
		return
	}

	h.logger.Debug("Get all channels attempt")

	channelsRepoRsp, err := h.queries.GetAllChannels(ctx)
	if err != nil {
		h.logger.Error("Failed to get channels", "error", err)
		http.Error(w, "Failed to get channels", http.StatusInternalServerError)
		return
	}

	channels := make([]channelResponse, len(channelsRepoRsp))
	for i, channel := range channelsRepoRsp {
		channels[i] = newChannelResponse(channel)
	}

	respondWithJSON(w, http.StatusOK, channels, failedEncodeChannelDataErrMsg)

	h.logger.Info("Channels retrieved successfully", "count", len(channels))
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if ctx.Err() != nil {
		h.logger.Error("Request context error", "error", ctx.Err())
		http.Error(w, "Request cancelled or timed out", http.StatusRequestTimeout)
		return
	}

	var req createChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !req.isValid() {
		h.logger.Error("Invalid channel data", "name", req.Name)
		http.Error(w, "Channel name cannot be empty", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Create channel attempt", "name", req.Name, "description", req.Description, "userID", req.UserID)

	createChannelParams := repository.CreateChannelParams{
		Name: req.Name,
		Description: sql.NullString{
			String: req.Description,
			Valid:  req.Description != "",
		},
		CreatedBy: req.UserID,
	}

	channel, err := h.queries.CreateChannel(ctx, createChannelParams)
	if err != nil {
		h.logger.Error("Failed to create channel", "error", err)
		http.Error(w, "Failed to create channel", http.StatusInternalServerError)
		return
	}

	response := newChannelResponse(channel)
	respondWithJSON(w, http.StatusCreated, response, failedEncodeChannelDataErrMsg)

	h.logger.Info(
		"Channel created successfully",
		"channelID", channel.ID,
		"name", channel.Name,
		"description", channel.Description,
		"createdBy", channel.CreatedBy,
	)
}
