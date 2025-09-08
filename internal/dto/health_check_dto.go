package dto

type HealthCheckResponseDTO struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Version string `json:"version"`
}

func NewHealthCheckResponse(status, version string, err error) HealthCheckResponseDTO {
	health := HealthCheckResponseDTO{
		Status:  status,
		Version: version,
	}
	if err != nil {
		health.Error = err.Error()
	}
	return health
}
