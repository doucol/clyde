// Package mock simulates a calico whisker-backend
package mock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/doucol/clyde/internal/flowdata"
	"github.com/doucol/clyde/internal/util"
)

const (
	Allow = "Allow"
	Deny  = "Deny"
	Src   = "Src"
	Dst   = "Dst"
)

type FlowKey struct {
	SrcNamespace   string
	SrcName        string
	DstNamespace   string
	DstName        string
	Protocol       string
	Port           int64
	SrcServiceName string
	SrcHash        string
	DstServiceName string
	DstHash        string
}

func (f *FlowKey) String() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d", f.SrcNamespace, f.SrcName, f.DstNamespace, f.DstName, f.Protocol, f.Port)
}

// ServerConfig holds configuration for the SSE server
type ServerConfig struct {
	Port           int
	StreamInterval time.Duration
	MaxClients     int
	Services       []string
	Namespaces     []string
	Protocols      []string
	Ports          []int
	FlowKeyCount   int
	AutoBroadcast  bool
}

// SSEServer manages the Server-Sent Events functionality
type SSEServer struct {
	config   ServerConfig
	clients  map[chan *flowdata.FlowResponse]bool
	flowKeys []*FlowKey
	server   *http.Server
	Connect  chan bool
}

// DefaultConfig returns a default server configuration
func DefaultConfig() ServerConfig {
	return ServerConfig{
		Port:           8080,
		StreamInterval: 15 * time.Second,
		MaxClients:     100,
		Services:       []string{"checkoutservice", "currencyservice", "paymentservice", "shippingservice", "cartservice"},
		Namespaces:     []string{"default", "production", "staging"},
		Protocols:      []string{"tcp", "udp"},
		Ports:          []int{80, 443, 8080, 8443, 3000, 5000, 7000},
		FlowKeyCount:   100,
		AutoBroadcast:  true,
	}
}

// NewSSEServer creates a new SSE server with the given configuration
func NewSSEServer(config ServerConfig) *SSEServer {
	return &SSEServer{
		config:   config,
		clients:  make(map[chan *flowdata.FlowResponse]bool),
		flowKeys: generateRandomFlowKeys(config),
		Connect:  make(chan bool, 100),
	}
}

func generateRandomFlowKeys(config ServerConfig) []*FlowKey {
	namespaces := config.Namespaces
	services := config.Services
	protocols := config.Protocols
	ports := config.Ports
	count := config.FlowKeyCount

	keys := make(map[string]*FlowKey, count)
	for {
		srcNamespace := namespaces[rand.Intn(len(namespaces))]
		dstNamespace := namespaces[rand.Intn(len(namespaces))]
		srcService := services[rand.Intn(len(services))]
		dstService := services[rand.Intn(len(services))]
		protocol := protocols[rand.Intn(len(protocols))]
		port := int64(ports[rand.Intn(len(ports))])
		srcHash := fmt.Sprintf("%x", rand.Uint32())
		dstHash := fmt.Sprintf("%x", rand.Uint32())

		fk := &FlowKey{
			SrcNamespace:   srcNamespace,
			SrcName:        fmt.Sprintf("%s-%s-*", srcService, srcHash),
			DstNamespace:   dstNamespace,
			DstName:        fmt.Sprintf("%s-%s-*", dstService, dstHash),
			Protocol:       protocol,
			Port:           port,
			SrcServiceName: srcService,
			DstServiceName: dstService,
			SrcHash:        srcHash,
			DstHash:        dstHash,
		}
		key := fk.String()
		if _, exists := keys[key]; exists {
			continue
		}
		keys[key] = fk
		if len(keys) >= count {
			return util.GetMapValues(keys)
		}
	}
}

type FlowPair struct {
	SrcFlow *flowdata.FlowResponse
	DstFlow *flowdata.FlowResponse
}

func (s *SSEServer) GenerateFlowPairs() map[string]FlowPair {
	count := rand.Intn(len(s.flowKeys) / 2)
	if count <= 0 {
		return map[string]FlowPair{}
	}
	keys := make(map[string]FlowPair)
	for {
		fk := s.flowKeys[rand.Intn(len(s.flowKeys))]
		key := fk.String()
		if _, exists := keys[key]; exists {
			continue
		}
		keys[key] = FlowPair{
			SrcFlow: s.generateFlow(fk, Src, Allow),
			DstFlow: s.generateFlow(fk, Dst, Allow),
		}
		if len(keys) >= count {
			return keys
		}
	}
}

func (s *SSEServer) generateFlow(fk *FlowKey, reporter, action string) *flowdata.FlowResponse {
	now := time.Now().UTC().Add(time.Second * -30)
	startTime := now.Round(time.Second * 15)
	endTime := startTime.Add(time.Second * 15)
	flow := &flowdata.FlowResponse{
		StartTime:       startTime,
		EndTime:         endTime,
		SourceNamespace: fk.SrcNamespace,
		SourceName:      fk.SrcName,
		DestNamespace:   fk.DstNamespace,
		DestName:        fk.DstName,
		Protocol:        fk.Protocol,
		DestPort:        fk.Port,
		SourceLabels:    fmt.Sprintf("app=%s | projectcalico.org/orchestrator=k8s", fk.SrcServiceName),
		DestLabels:      fmt.Sprintf("app=%s | projectcalico.org/orchestrator=k8s", fk.DstServiceName),
		Reporter:        reporter,
		Action:          action,
		Policies: flowdata.PolicyTrace{
			Enforced: []*flowdata.PolicyHit{
				{
					Kind:        "CalicoNetworkPolicy",
					Name:        fmt.Sprintf("allow-%s-ingress", "foo"),
					Namespace:   fk.SrcNamespace,
					Tier:        "default",
					Action:      action,
					PolicyIndex: 0,
					RuleIndex:   0,
					Trigger:     nil,
				},
			},
			Pending: []*flowdata.PolicyHit{
				{
					Kind:        "CalicoNetworkPolicy",
					Name:        fmt.Sprintf("allow-%s-ingress", "bar"),
					Namespace:   fk.SrcNamespace,
					Tier:        "default",
					Action:      action,
					PolicyIndex: 0,
					RuleIndex:   0,
					Trigger:     nil,
				},
			},
		},
		PacketsIn:  int64(rand.Intn(100) + 1),
		PacketsOut: int64(rand.Intn(100) + 1),
		BytesIn:    int64(rand.Intn(10000) + 100),
		BytesOut:   int64(rand.Intn(10000) + 100),
	}

	return flow
}

// addClient adds a new SSE client
func (s *SSEServer) addClient(client chan *flowdata.FlowResponse) {
	s.clients[client] = true
	s.Connect <- true
	log.Printf("Client connected. Total clients: %d", len(s.clients))
}

// removeClient removes an SSE client
func (s *SSEServer) removeClient(client chan *flowdata.FlowResponse) {
	delete(s.clients, client)
	close(client)
	log.Printf("Client disconnected. Total clients: %d", len(s.clients))
}

// Broadcast sends data to all connected clients
func (s *SSEServer) Broadcast(flow *flowdata.FlowResponse) {
	if len(s.clients) == 0 {
		panic("No clients connected to broadcast flow")
	}
	for client := range s.clients {
		select {
		case client <- flow:
		default:
			log.Printf("Client buffer full, removing client")
			// Client buffer is full, remove it
			s.removeClient(client)
		}
	}
}

func (s *SSEServer) BroadcastFlowPairs(flowPairs map[string]FlowPair) {
	log.Printf("Broadcasting %d flow pairs", len(flowPairs))
	for _, fp := range flowPairs {
		s.Broadcast(fp.SrcFlow)
		s.Broadcast(fp.DstFlow)
	}
}

// startBroadcaster begins the periodic data generation and broadcasting
func (s *SSEServer) startBroadcaster() {
	ticker := time.NewTicker(s.config.StreamInterval)
	go func() {
		for range ticker.C {
			if len(s.clients) > 0 {
				flowPairs := s.GenerateFlowPairs()
				s.BroadcastFlowPairs(flowPairs)
			}
		}
	}()
}

// sseHandler handles Server-Sent Events connections
func (s *SSEServer) sseHandler(w http.ResponseWriter, r *http.Request) {
	// Check client limit
	if len(s.clients) >= s.config.MaxClients {
		http.Error(w, "Maximum clients reached", http.StatusTooManyRequests)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create client channel
	client := make(chan *flowdata.FlowResponse, 1000)
	s.addClient(client)

	// Handle client disconnect
	defer s.removeClient(client)

	// Stream data to client
	for {
		select {
		case flow := <-client:
			jsonData, err := json.Marshal(flow)
			if err != nil {
				log.Printf("Error marshaling JSON: %v", err)
				return
			}
			_, err = fmt.Fprintf(w, "data: %s\n\n", jsonData)
			if err != nil {
				log.Printf("Error writing to client: %v", err)
				return
			}
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			log.Printf("Client disconnected")
			return
		}
	}
}

func (s *SSEServer) Close() error {
	ctx, release := context.WithTimeout(context.Background(), time.Second*5)
	defer release()
	defer func() {
		util.ChanClose(s.Connect)
	}()
	return s.server.Shutdown(ctx)
}

// Start begins the SSE server
func (s *SSEServer) Start(ready chan bool) error {
	if s.config.AutoBroadcast {
		s.startBroadcaster()
	}

	http.HandleFunc("/events", s.sseHandler)

	addr := ":" + strconv.Itoa(s.config.Port)
	s.server = &http.Server{Addr: addr}

	log.Printf("Starting SSE server on %s", addr)
	log.Printf("Stream interval: %v", s.config.StreamInterval)
	log.Printf("Max clients: %d", s.config.MaxClients)

	if ready != nil {
		close(ready)
	}
	if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *SSEServer) URL() string {
	return fmt.Sprintf("http://localhost:%d/events", s.config.Port)
}
