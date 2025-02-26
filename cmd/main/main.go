package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rivo/tview"
)

var (
	totalRequests      uint64
	requestsPerSecond  float64
	requestsPerMinute  float64
	lastRequests       uint64
	lastMinuteRequests uint64
)

// Prometheus metrics
var (
	requestCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "total_requests",
			Help: "Total number of processed requests.",
		},
	)
	requestsPerSecondGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "requests_per_second",
			Help: "Requests processed per second.",
		},
	)
	requestsPerMinuteGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "requests_per_minute",
			Help: "Requests processed per minute.",
		},
	)
)

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(requestCounter, requestsPerSecondGauge, requestsPerMinuteGauge)

	// Start metrics tracking goroutines
	go trackMetrics()

	// Start TUI
	go startTUI()

	// Initialize Fiber app
	app := fiber.New()

	// Endpoint to handle POST requests
	app.Post("/data", func(c *fiber.Ctx) error {
		var input struct {
			Username string `json:"username"`
			Age      int    `json:"age"`
		}
		if err := json.Unmarshal(c.Body(), &input); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		// Increment metrics
		atomic.AddUint64(&totalRequests, 1)
		atomic.AddUint64(&lastMinuteRequests, 1)
		requestCounter.Inc()

		return c.JSON(fiber.Map{
			"message": "Data received",
			"user":    input.Username,
			"age":     input.Age,
		})
	})

	// Expose Prometheus metrics
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Fatal(http.ListenAndServe(":13100", nil)) // Metrics on separate port
	}()

	// Start Fiber server
	log.Fatal(app.Listen(":13000"))
}

func trackMetrics() {
	secondTicker := time.NewTicker(time.Second)
	minuteTicker := time.NewTicker(time.Minute)
	defer secondTicker.Stop()
	defer minuteTicker.Stop()

	for {
		select {
		case <-secondTicker.C:
			currentRequests := atomic.LoadUint64(&totalRequests)
			delta := currentRequests - lastRequests
			lastRequests = currentRequests
			requestsPerSecond = float64(delta)
			requestsPerSecondGauge.Set(requestsPerSecond)
		case <-minuteTicker.C:
			requestsPerMinute = float64(atomic.LoadUint64(&lastMinuteRequests))
			requestsPerMinuteGauge.Set(requestsPerMinute)
			atomic.StoreUint64(&lastMinuteRequests, 0) // Reset for the next minute
		}
	}
}

func startTUI() {
	app := tview.NewApplication()
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("Loading...")

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		AddItem(textView, 1, 0, 1, 1, 0, 0, false)

	go func() {
		for {
			total := atomic.LoadUint64(&totalRequests)
			sPerSec := requestsPerSecond
			sPerMin := requestsPerMinute
			app.QueueUpdateDraw(func() {
				textView.SetText(fmt.Sprintf("Total Requests: %d\nRequests per Second: %.2f\nRequests per Minute: %.2f", total, sPerSec, sPerMin))
			})
			time.Sleep(time.Second)
		}
	}()

	if err := app.SetRoot(grid, true).Run(); err != nil {
		log.Fatal(err)
	}
}
