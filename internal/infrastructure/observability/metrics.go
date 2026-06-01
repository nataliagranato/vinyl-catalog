package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

var (
	coverUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_cover_uploads_total",
			Help: "Total cover image uploads",
		},
		[]string{"status", "ext"},
	)

	favoritesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_favorites_total",
			Help: "Total favorite toggle events",
		},
		[]string{"action"},
	)

	profilePhotoUploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vinyl_profile_photo_uploads_total",
			Help: "Total profile photo uploads",
		},
		[]string{"status"},
	)
)

func RecordCoverUpload(status, ext string) {
	coverUploadsTotal.WithLabelValues(status, ext).Inc()
}

func RecordFavoriteToggle(action string) {
	favoritesTotal.WithLabelValues(action).Inc()
}

func RecordProfilePhotoUpload(status string) {
	profilePhotoUploadsTotal.WithLabelValues(status).Inc()
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}
