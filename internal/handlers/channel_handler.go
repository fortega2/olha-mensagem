package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	reqCtxErrMsg                       = "Request context error"
	reqCtxCancelledOrTimedOutErrMsg    = "Request cancelled or timed out"
	failedEncodeChannelDataErrMsg      = "Failed to encode channel data"
	failedEncodeDeleteChannelRspErrMsg = "Failed to encode delete channel response"
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

type deleteChannelResponse struct {
	Message   string `json:"message"`
	ChannelID int64  `json:"channelId"`
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
		h.logger.Error(reqCtxErrMsg, "error", ctx.Err())
		http.Error(w, reqCtxCancelledOrTimedOutErrMsg, http.StatusRequestTimeout)
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
		h.logger.Error(reqCtxErrMsg, "error", ctx.Err())
		http.Error(w, reqCtxCancelledOrTimedOutErrMsg, http.StatusRequestTimeout)
		return
	}

	var req createChannelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !req.isValid() {
		h.logger.Error("Invalid channel data", "name", req.Name, "userID", req.UserID)
		http.Error(w, "Channel name and user ID are required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Create channel attempt", "name", req.Name, "description", req.Description, "userID", req.UserID)

	_, err := h.queries.GetUserByID(ctx, req.UserID)
	if err != nil {
		h.logger.Error("User not found", "userID", req.UserID, "error", err)
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

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

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if ctx.Err() != nil {
		h.logger.Error(reqCtxErrMsg, "error", ctx.Err())
		http.Error(w, reqCtxCancelledOrTimedOutErrMsg, http.StatusRequestTimeout)
		return
	}

	channelIdStr := chi.URLParam(r, "channelId")
	if channelIdStr == "" {
		h.logger.Error("Channel ID is required")
		http.Error(w, "Channel ID is required", http.StatusBadRequest)
		return
	}

	userIdStr := chi.URLParam(r, "userId")
	if userIdStr == "" {
		h.logger.Error("User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Delete channel attempt", "channelID", channelIdStr, "userID", userIdStr)

	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid channel ID", "error", err)
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid user ID", "error", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	channel, err := h.queries.GetChannelByID(ctx, channelId)
	if err != nil {
		h.logger.Error("Channel not found", "channelID", channelId, "error", err)
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	if channel.CreatedBy != userId {
		h.logger.Error("User is not the creator of the channel",
			"channelID", channelId,
			"userID", userId,
			"createdBy", channel.CreatedBy)
		http.Error(w, "Only the channel creator can delete this channel", http.StatusForbidden)
		return
	}

	err = h.queries.DeleteChannel(ctx, repository.DeleteChannelParams{
		ID:        channelId,
		CreatedBy: userId,
	})
	if err != nil {
		h.logger.Error("Failed to delete channel", "error", err)
		http.Error(w, "Failed to delete channel", http.StatusInternalServerError)
		return
	}

	response := deleteChannelResponse{
		Message:   fmt.Sprintf("Channel '%s' deleted successfully", channel.Name),
		ChannelID: channelId,
	}
	respondWithJSON(w, http.StatusOK, response, failedEncodeDeleteChannelRspErrMsg)

	h.logger.Info("Channel deleted successfully",
		"channelID", channelId,
		"channelName", channel.Name,
		"userID", userId)
}
