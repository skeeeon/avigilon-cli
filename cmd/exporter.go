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
	serviceAction string // "install", "uninstall", "start", "stop"
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
		// We exit here so the service manager (systemd/Windows Services) knows we failed and can handle restarts.
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

// --- COLLECTOR LOGIC ---

type AvigilonCollector struct {
	Client *client.AvigilonClient
	Mutex  sync.Mutex
}

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
		if strings.Contains(healthStr, "GOOD") {
			healthVal = 1.0
		}
		if strings.Contains(healthStr, "WARN") {
			healthVal = 0.5
		}
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
			if strings.EqualFold(cam.ConnectionState, "CONNECTED") {
				isUp = 1.0
			}
			ip := cam.IPAddress
			if ip == "" {
				ip = "unknown"
			}

			ch <- prometheus.MustNewConstMetric(cameraUpDesc, prometheus.GaugeValue, isUp, cam.ID, cam.Name, cam.Model, ip)

			hasRec := 0.0
			if cam.RecordedData {
				hasRec = 1.0
			}
			ch <- prometheus.MustNewConstMetric(cameraRecordingDesc, prometheus.GaugeValue, hasRec, cam.ID, cam.Name)

			st := strings.ToUpper(cam.ConnectionState)
			if st == "" {
				st = "UNKNOWN"
			}
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
			if st == "" {
				st = "UNKNOWN"
			}
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
	if err == nil {
		return res, nil
	}
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil {
			return c.Client.GetCameras()
		}
	}
	return nil, err
}
func (c *AvigilonCollector) fetchAlarmsWithRetry() ([]models.Alarm, error) {
	res, err := c.Client.GetAlarms()
	if err == nil {
		return res, nil
	}
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil {
			return c.Client.GetAlarms()
		}
	}
	return nil, err
}
func (c *AvigilonCollector) fetchServersWithRetry() ([]models.Server, error) {
	res, err := c.Client.GetServers()
	if err == nil {
		return res, nil
	}
	if isAuthError(err) {
		if _, e := c.Client.Login(); e == nil {
			return c.Client.GetServers()
		}
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
Can be installed as a system service on Windows or Linux.

Environment Variable Support:
If flags are not provided, the exporter will look for the following environment variables. 
This is the recommended way to configure the Windows Service via Registry 
(HKLM\SYSTEM\CurrentControlSet\Services\avigilon-exporter\Environment).

  AVIGILON_HOST
  AVIGILON_USERNAME
  AVIGILON_PASSWORD
  AVIGILON_NONCE
  AVIGILON_KEY
  AVIGILON_INTEGRATION_ID
  AVIGILON_PORT
`,
	Run: func(cmd *cobra.Command, args []string) {

		// ---------------------------------------------------------
		// 1. Runtime Fallback: Check Env Vars if flags are empty
		//    This allows the Service to pick up Registry keys at runtime.
		// ---------------------------------------------------------
		if expHost == "" {
			expHost = os.Getenv("AVIGILON_HOST")
		}
		if expPass == "" {
			expPass = os.Getenv("AVIGILON_PASSWORD")
		}
		if expNonce == "" {
			expNonce = os.Getenv("AVIGILON_NONCE")
		}
		if expKey == "" {
			expKey = os.Getenv("AVIGILON_KEY")
		}
		if expIntID == "" {
			expIntID = os.Getenv("AVIGILON_INTEGRATION_ID")
		}

		// Handle Port override
		if envPort := os.Getenv("AVIGILON_PORT"); envPort != "" && expPort == "9100" {
			expPort = envPort
		}

		// Handle Username override (Flag default is "administrator")
		// If the user hasn't explicitly set the flag (hard to tell with Cobra default),
		// but has set the ENV var, we prioritize the ENV var to allow overriding default.
		if envUser := os.Getenv("AVIGILON_USERNAME"); envUser != "" {
			expUser = envUser
		}

		// ---------------------------------------------------------
		// 2. Prepare Client Configuration
		// ---------------------------------------------------------
		hostClean := strings.TrimRight(expHost, "/")
		cfg := client.ClientConfig{
			BaseURL:       hostClean,
			Username:      expUser,
			Password:      expPass,
			UserNonce:     expNonce,
			UserKey:       expKey,
			IntegrationID: expIntID,
		}

		// ---------------------------------------------------------
		// 3. Define Service Configuration & Arguments
		// ---------------------------------------------------------

		// dynamically build arguments based on what is currently set.
		// If a variable is empty here, it means it wasn't provided by Flag OR Env.
		// We do NOT want to bake empty flags into the service arguments,
		// because that would override the Registry/Env check on the next run.
		svcArgs := []string{"exporter"}

		if expHost != "" {
			svcArgs = append(svcArgs, "--host", expHost)
		}
		if expUser != "administrator" { // Only bake in if different from default
			svcArgs = append(svcArgs, "--username", expUser)
		}
		// NOTE: If you are installing the service and want it to use Registry Credentials,
		// do NOT pass the --password/--key/--nonce flags during install.
		if expPass != "" {
			svcArgs = append(svcArgs, "--password", expPass)
		}
		if expNonce != "" {
			svcArgs = append(svcArgs, "--nonce", expNonce)
		}
		if expKey != "" {
			svcArgs = append(svcArgs, "--key", expKey)
		}
		if expIntID != "" {
			svcArgs = append(svcArgs, "--integration-id", expIntID)
		}
		if expPort != "9100" {
			svcArgs = append(svcArgs, "--port", expPort)
		}

		svcConfig := &service.Config{
			Name:        "avigilon-exporter",
			DisplayName: "Avigilon Prometheus Exporter",
			Description: "Exposes Avigilon VMS metrics to Prometheus",
			Arguments:   svcArgs,
		}

		prg := &program{
			api: client.New(cfg),
		}

		s, err := service.New(prg, svcConfig)
		if err != nil {
			log.Fatal(err)
		}

		// ---------------------------------------------------------
		// 4. Handle Service Control Actions
		// ---------------------------------------------------------
		if serviceAction != "" {
			// INSTALL VALIDATION
			if serviceAction == "install" {
				// We no longer Fatal() here. If credentials are missing, we assume
				// the user intends to configure them via Registry/Env later.
				if expPass == "" || expNonce == "" || expKey == "" {
					fmt.Println("Warning: Credentials not provided via flags.")
					fmt.Println("To complete setup, you must set the following Environment Variables")
					fmt.Println("(or Registry Keys in HKLM\\SYSTEM\\CurrentControlSet\\Services\\avigilon-exporter\\Environment):")
					fmt.Println("  - AVIGILON_HOST")
					fmt.Println("  - AVIGILON_PASSWORD")
					fmt.Println("  - AVIGILON_NONCE")
					fmt.Println("  - AVIGILON_KEY")
				}
			}

			err = service.Control(s, serviceAction)
			if err != nil {
				log.Fatalf("Failed to %s service: %v", serviceAction, err)
			}
			fmt.Printf("Service action '%s' completed successfully.\n", serviceAction)
			return
		}

		// ---------------------------------------------------------
		// 5. Run the Service (Interactive or Service Manager)
		// ---------------------------------------------------------

		// STRICT VALIDATION: If we are actually trying to run (not just install),
		// we must have credentials now (either from Flags or Env/Registry).
		if expHost == "" || expPass == "" || expNonce == "" || expKey == "" {
			log.Fatal("Fatal Error: Missing required credentials.\nPlease provide flags or set AVIGILON_* environment variables.")
		}

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

	// Service Control Flag
	exporterCmd.Flags().StringVar(&serviceAction, "service", "", "Service action: install, uninstall, start, stop")
}
