package contracts

import "time"

type HealthResponse struct {
	Status        string    `json:"status"`
	Service       string    `json:"service"`
	Version       string    `json:"version"`
	ServerTime    time.Time `json:"server_time"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}
