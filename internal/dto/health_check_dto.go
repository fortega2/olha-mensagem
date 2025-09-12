package dto

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

type HealthCheckResponseDTO struct {
	Status  HealthStatus `json:"status"`
	Error   string       `json:"error,omitempty"`
	Version string       `json:"version"`
}

func NewHealthCheckResponse(status HealthStatus, version string, err error) HealthCheckResponseDTO {
	health := HealthCheckResponseDTO{
		Status:  status,
		Version: version,
	}
	if err != nil {
		health.Error = err.Error()
	}
	return health
}
