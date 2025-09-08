package handlers

import (
	"net/http"
	"os"

	"github.com/fortega2/real-time-chat/internal/dto"
)

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if ctx.Err() != nil {
		h.logger.Error(reqCtxErrMsg, "error", ctx.Err())
		http.Error(w, reqCtxCancelledOrTimedOutErrMsg, http.StatusRequestTimeout)
		return
	}

	if err := h.db.PingContext(ctx); err != nil {
		h.logger.Error("Database ping failed", "error", err)
		respondWithJSON(w, http.StatusServiceUnavailable, dto.NewHealthCheckResponse("unhealthy", getVersion(), err), failedEncodeHealthCheckErrMsg)
		return
	}

	respondWithJSON(w, http.StatusOK, dto.NewHealthCheckResponse("healthy", getVersion(), nil), failedEncodeHealthCheckErrMsg)
	h.logger.Info("Health check successful")
}

func getVersion() string {
	if version := os.Getenv("APP_VERSION"); version != "" {
		return version
	}
	return "1.0.0"
}
