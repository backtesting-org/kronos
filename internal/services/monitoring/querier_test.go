package monitoring_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	monitoring2 "github.com/backtesting-org/kronos-sdk/pkg/types/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backtesting-org/kronos-cli/internal/services/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

var _ = Describe("Querier", func() {
	var (
		tmpDir     string
		querier    monitoring2.ViewQuerier
		server     net.Listener
		httpServer *http.Server
		instanceID string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "querier-test-*")
		Expect(err).NotTo(HaveOccurred())

		instanceID = "test-strategy"
		querier = monitoring.NewQuerierWithConfig(tmpDir, 5*time.Second)
	})

	AfterEach(func() {
		if httpServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = httpServer.Shutdown(ctx)
		}
		if server != nil {
			_ = server.Close()
		}
		os.RemoveAll(tmpDir)
	})

	// Helper to start a mock server
	startMockServer := func(handler http.Handler) {
		socketPath := filepath.Join(tmpDir, instanceID+".sock")
		var err error
		server, err = net.Listen("unix", socketPath)
		Expect(err).NotTo(HaveOccurred())

		httpServer = &http.Server{Handler: handler}
		go func() {
			_ = httpServer.Serve(server)
		}()

		// Wait for server to be ready
		time.Sleep(10 * time.Millisecond)
	}

	Describe("QueryPnL", func() {
		It("should return PnL data from running instance", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/pnl", func(w http.ResponseWriter, r *http.Request) {
				pnl := monitoring2.PnLView{
					StrategyName:  "test-strategy",
					RealizedPnL:   numerical.NewFromFloat(100.50),
					UnrealizedPnL: numerical.NewFromFloat(25.25),
					TotalPnL:      numerical.NewFromFloat(125.75),
					TotalFees:     numerical.NewFromFloat(5.00),
				}
				json.NewEncoder(w).Encode(pnl)
			})
			startMockServer(mux)

			result, err := querier.QueryPnL(instanceID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StrategyName).To(Equal("test-strategy"))
			Expect(result.TotalPnL.String()).To(Equal("125.75"))
		})

		It("should return error when instance not found", func() {
			_, err := querier.QueryPnL("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("QueryPositions", func() {
		It("should return positions from running instance", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/positions", func(w http.ResponseWriter, r *http.Request) {
				positions := strategy.StrategyExecution{
					Orders: []connector.Order{
						{ID: "order-1", Symbol: "BTC", Side: connector.OrderSideBuy},
					},
					Trades: []connector.Trade{
						{ID: "trade-1", Symbol: "BTC"},
					},
				}
				json.NewEncoder(w).Encode(positions)
			})
			startMockServer(mux)

			result, err := querier.QueryPositions(instanceID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Orders).To(HaveLen(1))
			Expect(result.Orders[0].ID).To(Equal("order-1"))
		})
	})

	Describe("QueryOrderbook", func() {
		It("should return orderbook for asset", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/orderbook", func(w http.ResponseWriter, r *http.Request) {
				asset := r.URL.Query().Get("asset")
				Expect(asset).To(Equal("BTC"))

				orderbook := connector.OrderBook{
					Asset: portfolio.NewAsset("BTC"),
					Bids: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50000), Quantity: numerical.NewFromFloat(1.5)},
					},
					Asks: []connector.PriceLevel{
						{Price: numerical.NewFromFloat(50100), Quantity: numerical.NewFromFloat(2.0)},
					},
				}
				json.NewEncoder(w).Encode(orderbook)
			})
			startMockServer(mux)

			result, err := querier.QueryOrderbook(instanceID, "BTC", "binance")
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Bids).To(HaveLen(1))
			Expect(result.Asks).To(HaveLen(1))
		})
	})

	Describe("QueryRecentTrades", func() {
		It("should return recent trades with limit", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/trades", func(w http.ResponseWriter, r *http.Request) {
				limit := r.URL.Query().Get("limit")
				Expect(limit).To(Equal("10"))

				trades := []connector.Trade{
					{ID: "trade-1", Symbol: "BTC"},
					{ID: "trade-2", Symbol: "ETH"},
				}
				json.NewEncoder(w).Encode(trades)
			})
			startMockServer(mux)

			result, err := querier.QueryRecentTrades(instanceID, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveLen(2))
		})
	})

	Describe("QueryMetrics", func() {
		It("should return strategy metrics", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, r *http.Request) {
				metrics := monitoring2.StrategyMetrics{
					StrategyName:     "test-strategy",
					Status:           "running",
					SignalsGenerated: 42,
					SignalsExecuted:  38,
				}
				json.NewEncoder(w).Encode(metrics)
			})
			startMockServer(mux)

			result, err := querier.QueryMetrics(instanceID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.StrategyName).To(Equal("test-strategy"))
			Expect(result.Status).To(Equal("running"))
			Expect(result.SignalsGenerated).To(Equal(42))
		})
	})

	Describe("HealthCheck", func() {
		It("should succeed when instance is healthy", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
			})
			startMockServer(mux)

			err := querier.HealthCheck(instanceID)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should fail when instance is not found", func() {
			err := querier.HealthCheck("nonexistent")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListInstances", func() {
		It("should return empty list when no instances", func() {
			instances, err := querier.ListInstances()
			Expect(err).NotTo(HaveOccurred())
			Expect(instances).To(BeEmpty())
		})

		It("should return list of instance IDs from socket files", func() {
			// Create some socket files
			for _, name := range []string{"momentum", "arbitrage", "mean-reversion"} {
				socketPath := filepath.Join(tmpDir, name+".sock")
				f, err := os.Create(socketPath)
				Expect(err).NotTo(HaveOccurred())
				f.Close()
			}

			instances, err := querier.ListInstances()
			Expect(err).NotTo(HaveOccurred())
			Expect(instances).To(HaveLen(3))
			Expect(instances).To(ContainElements("momentum", "arbitrage", "mean-reversion"))
		})

		It("should ignore non-socket files", func() {
			// Create socket and non-socket files
			os.Create(filepath.Join(tmpDir, "momentum.sock"))
			os.Create(filepath.Join(tmpDir, "config.json"))
			os.Create(filepath.Join(tmpDir, "readme.txt"))

			instances, err := querier.ListInstances()
			Expect(err).NotTo(HaveOccurred())
			Expect(instances).To(HaveLen(1))
			Expect(instances[0]).To(Equal("momentum"))
		})
	})

	Describe("QueryProfilingStats", func() {
		It("should return profiling stats from running instance", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/profiling/stats", func(w http.ResponseWriter, r *http.Request) {
				stats := monitoring2.ProfilingStats{}
				json.NewEncoder(w).Encode(stats)
			})
			startMockServer(mux)

			result, err := querier.QueryProfilingStats(instanceID)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})

		It("should return error when instance not found", func() {
			_, err := querier.QueryProfilingStats("nonexistent")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})

	Describe("QueryRecentExecutions", func() {
		It("should return recent executions with limit parameter", func() {
			mux := http.NewServeMux()
			mux.HandleFunc("/profiling/executions", func(w http.ResponseWriter, r *http.Request) {
				limit := r.URL.Query().Get("limit")
				Expect(limit).To(Equal("10"))

				executions := []monitoring2.ProfilingMetrics{}
				json.NewEncoder(w).Encode(executions)
			})
			startMockServer(mux)

			result, err := querier.QueryRecentExecutions(instanceID, 10)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
		})

		It("should return error when instance not found", func() {
			_, err := querier.QueryRecentExecutions("nonexistent", 10)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("not found"))
		})
	})
})
