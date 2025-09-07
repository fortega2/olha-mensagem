package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/go-chi/chi/v5"
)

const (
	messageLimitDefault int64 = 50
)

func (h *Handler) GetHistoryMessagesByChannel(w http.ResponseWriter, r *http.Request) {
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

	h.logger.Debug("Fetching messages for channel ID: ", channelIdStr)

	channelId, err := strconv.ParseInt(channelIdStr, 10, 64)
	if err != nil {
		h.logger.Error("Invalid channel ID", "error", err)
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	messages, err := h.queries.GetHistoryMessagesByChannel(ctx, repository.GetHistoryMessagesByChannelParams{
		ChannelID: channelId,
		Limit:     getMessageLimit(),
	})
	if err != nil {
		h.logger.Error("Failed to fetch messages", "error", err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Fetched messages", "chanedId", channelId, "count", len(messages))

	messagesDTO := make([]dto.MessageDTO, len(messages))
	for i, msg := range messages {
		messagesDTO[i] = dto.NewMessageDTO(msg)
	}

	respondWithJSON(w, http.StatusOK, messagesDTO, failedEncodeMessageDataErrMsg)

	h.logger.Info("Successfully fetched messages", "channelId", channelId, "count", len(messagesDTO))
}

func getMessageLimit() int64 {
	messagesLimit := os.Getenv("MESSAGES_LIMIT")
	if messagesLimit == "" {
		return messageLimitDefault
	}

	limit, err := strconv.ParseInt(messagesLimit, 10, 64)
	if err != nil || limit <= 0 {
		return messageLimitDefault
	}

	return limit
}
