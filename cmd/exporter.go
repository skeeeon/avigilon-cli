package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"avigilon-cli/internal/client"
	"avigilon-cli/pkg/models"
)

// Variables to hold flag values
var (
	expHost       string
	expUser       string
	expPass       string
	expNonce      string
	expKey        string
	expIntID      string
	expPort       string
	serviceAction string // New flag for "install", "uninstall", "start", "stop"
)

// --- SERVICE WRAPPER ---

// program implements the kardianos/service interface
type program struct {
	exit    chan struct{}
	server  *http.Server
	api     *client.AvigilonClient
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() {
	// 1. Initial Login
	log.Println("Attempting initial login...")
	if _, err := p.api.Login(); err != nil {
		log.Printf("Fatal: Initial login failed: %v", err)
		// In a service context, we might want to retry loop here instead of dying,
		// but for now we exit so the service manager attempts a restart.
		os.Exit(1) 
	}
	log.Println("Initial login successful.")

	// 2. Setup Prometheus
	registry := prometheus.NewRegistry()
	collector := &AvigilonCollector{Client: p.api}
	registry.MustRegister(collector)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog: log.Default(),
	})

	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)

	addr := fmt.Sprintf(":%s", expPort)
	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Avigilon Exporter listening on %s", addr)
	
	// Blocking call to listen
	if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP Server error: %v", err)
	}
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Signal the app to stop.
	log.Println("Stopping service...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if p.server != nil {
		if err := p.server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}
	close(p.exit)
	return nil
}

// --- COLLECTOR LOGIC (Same as before) ---

type AvigilonCollector struct {
	Client *client.AvigilonClient
	Mutex  sync.Mutex
}

// ... (Keep existing Metric Descriptors: upDesc, cameraUpDesc, etc.) ...
// For brevity in this snippet, ensure you keep the var definitions from the previous exporter.go here.
// I will include the variable definitions below to ensure the file is copy-pasteable.

var (
	upDesc = prometheus.NewDesc(
		"avigilon_up", "Was the last scrape successful.", nil, nil,
	)
	scrapeDurationDesc = prometheus.NewDesc(
		"avigilon_scrape_duration_seconds", "Time taken to scrape API.", nil, nil,
	)
	systemHealthDesc = prometheus.NewDesc(
		"avigilon_system_health", "VMS Health Status (1.0=GOOD, 0.5=WARN, 0.0=BAD).", nil, nil,
	)
	serverCountDesc = prometheus.NewDesc(
		"avigilon_servers_total", "Number of servers detected.", nil, nil,
	)
	cameraUpDesc = prometheus.NewDesc(
		"avigilon_camera_up", "Connection status.", []string{"id", "name", "model", "ip"}, nil,
	)
	cameraRecordingDesc = prometheus.NewDesc(
		"avigilon_camera_has_recorded_data", "Has recorded data.", []string{"id", "name"}, nil,
	)
	cameraCountDesc = prometheus.NewDesc(
		"avigilon_cameras_total", "Total cameras grouped by state.", []string{"state"}, nil,
	)
	alarmsCountDesc = prometheus.NewDesc(
		"avigilon_alarms_total", "Total alarms grouped by state.", []string{"state"}, nil,
	)
)

func (c *AvigilonCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- scrapeDurationDesc
	ch <- systemHealthDesc
	ch <- serverCountDesc
	ch <- cameraUpDesc
	ch <- cameraRecordingDesc
	ch <- cameraCountDesc
	ch <- alarmsCountDesc
}

func (c *AvigilonCollector) Collect(ch chan<- prometheus.Metric) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	start := time.Now()
	success := 1.0

	// 1. Health
	healthStr, err := c.Client.GetHealth()
	healthVal := 0.0
	if err == nil {
		if strings.Contains(healthStr, "GOOD") { healthVal = 1.0 }
		if strings.Contains(healthStr, "WARN") { healthVal = 0.5 }
	}
	ch <- prometheus.MustNewConstMetric(systemHealthDesc, prometheus.GaugeValue, healthVal)

	// 2. Servers
	if srvs, err := c.fetchServersWithRetry(); err == nil {
		ch <- prometheus.MustNewConstMetric(serverCountDesc, prometheus.GaugeValue, float64(len(srvs)))
	}

	// 3. Cameras
	if cams, err := c.fetchCamerasWithRetry(); err == nil {
		stateCounts := make(map[string]float64)
		for _, cam := range cams {
			isUp := 0.0
			if strings.EqualFold(cam.ConnectionState, "CONNECTED") { isUp = 1.0 }
			ip := cam.IPAddress
			if ip == "" { ip = "unknown" }
			
			ch <- prometheus.MustNewConstMetric(cameraUpDesc, prometheus.GaugeValue, isUp, cam.ID, cam.Name, cam.Model, ip)
			
			hasRec := 0.0
			if cam.RecordedData { hasRec = 1.0 }
			ch <- prometheus.MustNewConstMetric(cameraRecordingDesc, prometheus.GaugeValue, hasRec, cam.ID, cam.Name)

			st := strings.ToUpper(cam.ConnectionState)
			if st == "" { st = "UNKNOWN" }
			stateCounts[st]++
		}
		for st, cnt := range stateCounts {
			ch <- prometheus.MustNewConstMetric(cameraCountDesc, prometheus.GaugeValue, cnt, st)
		}
	} else {
		success = 0.0
		log.Printf("Error scraping cameras: %v", err)
	}

	// 4. Alarms
	if alarms, err := c.fetchAlarmsWithRetry(); err == nil {
		alarmStates := make(map[string]float64)
		for _, a := range alarms {
			st := strings.ToUpper(a.State)
			if st == "" { st = "UNKNOWN" }
			alarmStates[st]++
		}
		for st, cnt := range alarmStates {
			ch <- prometheus.MustNewConstMetric(alarmsCountDesc, prometheus.GaugeValue, cnt, st)
		}
	} else {
		success = 0.0
		log.Printf("Error scraping alarms: %v", err)
	}

	ch <- prometheus.MustNewConstMetric(upDesc, prometheus.GaugeValue, success)
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, time.Since(start).Seconds())
}

// --- RETRY HELPERS ---
func (c *AvigilonCollector) fetchCamerasWithRetry() ([]models.Camera, error) {
	res, err := c.Client.GetCameras()
	if err == nil { return res, nil }
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil { return c.Client.GetCameras() }
	}
	return nil, err
}
func (c *AvigilonCollector) fetchAlarmsWithRetry() ([]models.Alarm, error) {
	res, err := c.Client.GetAlarms()
	if err == nil { return res, nil }
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil { return c.Client.GetAlarms() }
	}
	return nil, err
}
func (c *AvigilonCollector) fetchServersWithRetry() ([]models.Server, error) {
	res, err := c.Client.GetServers()
	if err == nil { return res, nil }
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil { return c.Client.GetServers() }
	}
	return nil, err
}
func isAuthError(err error) bool {
	return strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "403")
}

// --- COMMAND ---

var exporterCmd = &cobra.Command{
	Use:   "exporter",
	Short: "Start Prometheus Exporter service",
	Long: `Starts a long-running HTTP server that exposes Avigilon metrics.
Can be installed as a system service.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Setup Client Config
		hostClean := strings.TrimRight(expHost, "/")
		cfg := client.ClientConfig{
			BaseURL:       hostClean,
			Username:      expUser,
			Password:      expPass,
			UserNonce:     expNonce,
			UserKey:       expKey,
			IntegrationID: expIntID,
		}

		// 2. Define Service Configuration
		svcConfig := &service.Config{
			Name:        "avigilon-exporter",
			DisplayName: "Avigilon Prometheus Exporter",
			Description: "Exposes Avigilon VMS metrics to Prometheus",
			// Arguments passed to the binary when run as a service
			Arguments: []string{
				"exporter",
				"--host", expHost,
				"--username", expUser,
				"--password", expPass,
				"--nonce", expNonce,
				"--key", expKey,
				"--port", expPort,
			},
		}
		if expIntID != "" {
			svcConfig.Arguments = append(svcConfig.Arguments, "--integration-id", expIntID)
		}

		prg := &program{
			api: client.New(cfg),
		}

		s, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal(err)
		}

		// 3. Handle Service Control Actions (Install, Start, Stop, Uninstall)
		if serviceAction != "" {
			if serviceAction == "install" {
				// Validate required flags before installing
				if expHost == "" || expPass == "" || expNonce == "" || expKey == "" {
					log.Fatal("Error: You must provide all credentials (--host, --password, --nonce, --key) to install the service.")
				}
			}
			
			err = service.Control(s, serviceAction)
			if err != nil {
				log.Fatalf("Failed to %s service: %v", serviceAction, err)
			}
			fmt.Printf("Service action '%s' completed successfully.\n", serviceAction)
			return
		}

		// 4. Run the Service (Blocking)
		// This happens when the Service Manager starts the binary, OR when run interactively without flags
		logger, err := s.Logger(nil)
		if err != nil {
			log.Fatal(err)
		}
		if err = s.Run(); err != nil {
			logger.Error(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(exporterCmd)
	exporterCmd.Flags().StringVar(&expHost, "host", "", "API Base URL")
	exporterCmd.Flags().StringVar(&expUser, "username", "administrator", "ACC Username")
	exporterCmd.Flags().StringVar(&expPass, "password", "", "ACC Password")
	exporterCmd.Flags().StringVar(&expNonce, "nonce", "", "User Nonce")
	exporterCmd.Flags().StringVar(&expKey, "key", "", "User Key")
	exporterCmd.Flags().StringVar(&expIntID, "integration-id", "", "Integration ID")
	exporterCmd.Flags().StringVar(&expPort, "port", "9100", "Port to listen on")
	
	// New Flag for Service Control
	exporterCmd.Flags().StringVar(&serviceAction, "service", "", "Service action: install, uninstall, start, stop")
}
