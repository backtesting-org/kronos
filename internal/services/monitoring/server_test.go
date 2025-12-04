package monitoring_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backtesting-org/kronos-cli/internal/services/monitoring"
	monitoringMock "github.com/backtesting-org/kronos-cli/mocks/github.com/backtesting-org/kronos-cli/pkg/monitoring"
	pkgmonitoring "github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

var _ = Describe("Server", func() {
	var (
		tmpDir       string
		viewRegistry *monitoringMock.ViewRegistry
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "monitoring-test-*")
		Expect(err).NotTo(HaveOccurred())

		viewRegistry = monitoringMock.NewViewRegistry(GinkgoT())
	})

	AfterEach(func() {
		_ = os.RemoveAll(tmpDir)
	})

	Describe("NewServer", func() {
		It("should require instance ID", func() {
			_, err := monitoring.NewServer(pkgmonitoring.ServerConfig{}, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("instance ID"))
		})

		It("should create socket directory", func() {
			socketDir := filepath.Join(tmpDir, "sockets")

			server, err := monitoring.NewServer(pkgmonitoring.ServerConfig{
				InstanceID: "test-instance",
				SocketDir:  socketDir,
			}, viewRegistry)

			Expect(err).NotTo(HaveOccurred())
			Expect(server).NotTo(BeNil())

			_, err = os.Stat(socketDir)
			Expect(err).NotTo(HaveOccurred())

			expectedPath := filepath.Join(socketDir, "test-instance.sock")
			Expect(server.SocketPath()).To(Equal(expectedPath))
		})
	})

	Describe("Start and Stop", func() {
		It("should start and create socket file", func() {
			server, err := monitoring.NewServer(pkgmonitoring.ServerConfig{
				InstanceID: "test-instance",
				SocketDir:  tmpDir,
			}, viewRegistry)
			Expect(err).NotTo(HaveOccurred())

			errChan := make(chan error, 1)
			go func() {
				errChan <- server.Start()
			}()

			Eventually(func() bool {
				_, err := os.Stat(server.SocketPath())
				return err == nil
			}, "2s", "100ms").Should(BeTrue())

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			_ = server.Stop(ctx)

			_, err = os.Stat(server.SocketPath())
			Expect(os.IsNotExist(err)).To(BeTrue())

			Eventually(errChan, "2s").Should(Receive(BeNil()))
		})

		It("should fail when starting twice", func() {
			server, err := monitoring.NewServer(pkgmonitoring.ServerConfig{
				InstanceID: "test-double-start",
				SocketDir:  tmpDir,
			}, viewRegistry)
			Expect(err).NotTo(HaveOccurred())

			go func() { _ = server.Start() }()

			Eventually(func() bool {
				_, err := os.Stat(server.SocketPath())
				return err == nil
			}, "2s", "100ms").Should(BeTrue())

			err = server.Start()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already started"))

			_ = server.Stop(context.Background())
		})
	})

	Describe("HTTP Endpoints", func() {
		var (
			server pkgmonitoring.Server
			client *http.Client
		)

		BeforeEach(func() {
			var err error
			server, err = monitoring.NewServer(pkgmonitoring.ServerConfig{
				InstanceID: "test-endpoints",
				SocketDir:  tmpDir,
			}, viewRegistry)
			Expect(err).NotTo(HaveOccurred())

			go func() { _ = server.Start() }()

			Eventually(func() bool {
				_, err := os.Stat(server.SocketPath())
				return err == nil
			}, "2s", "100ms").Should(BeTrue())

			client = &http.Client{
				Transport: &http.Transport{
					DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
						return net.Dial("unix", server.SocketPath())
					},
				},
				Timeout: 5 * time.Second,
			}
		})

		AfterEach(func() {
			_ = server.Stop(context.Background())
		})

		Describe("/health", func() {
			It("should return health status", func() {
				viewRegistry.EXPECT().GetHealth().Return(&health.SystemHealthReport{
					OverallState: health.StateConnected,
					HasErrors:    false,
				})

				resp, err := client.Get("http://unix/health")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var result health.SystemHealthReport
				err = json.NewDecoder(resp.Body).Decode(&result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.OverallState).To(Equal(health.StateConnected))
			})

			It("should return default health when nil", func() {
				viewRegistry.EXPECT().GetHealth().Return(nil)

				resp, err := client.Get("http://unix/health")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				var result health.SystemHealthReport
				err = json.NewDecoder(resp.Body).Decode(&result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.OverallState).To(Equal(health.StateConnected))
			})
		})

		Describe("/api/pnl", func() {
			It("should return PnL data", func() {
				viewRegistry.EXPECT().GetPnLView().Return(map[string]interface{}{
					"realized": 1000.0,
				})

				resp, err := client.Get("http://unix/api/pnl")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var result map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&result)
				Expect(err).NotTo(HaveOccurred())
				Expect(result["realized"]).To(Equal(1000.0))
			})

			It("should return empty object when nil", func() {
				viewRegistry.EXPECT().GetPnLView().Return(nil)

				resp, err := client.Get("http://unix/api/pnl")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Describe("/api/positions", func() {
			It("should return positions", func() {
				viewRegistry.EXPECT().GetPositionsView().Return(&strategy.StrategyExecution{
					Orders: []connector.Order{{ID: "order-1"}},
					Trades: []connector.Trade{{ID: "trade-1"}},
				})

				resp, err := client.Get("http://unix/api/positions")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should return empty object when nil", func() {
				viewRegistry.EXPECT().GetPositionsView().Return(nil)

				resp, err := client.Get("http://unix/api/positions")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Describe("/api/orderbook", func() {
			It("should return orderbook for asset", func() {
				viewRegistry.EXPECT().GetOrderbookView("BTC/USDT").Return(&connector.OrderBook{
					Bids: []connector.PriceLevel{},
					Asks: []connector.PriceLevel{},
				})

				resp, err := client.Get("http://unix/api/orderbook?asset=BTC/USDT")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should require asset parameter", func() {
				resp, err := client.Get("http://unix/api/orderbook")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return 404 for unknown asset", func() {
				viewRegistry.EXPECT().GetOrderbookView("DOGE/USDT").Return(nil)

				resp, err := client.Get("http://unix/api/orderbook?asset=DOGE/USDT")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Describe("/api/trades", func() {
			It("should return trades with default limit", func() {
				viewRegistry.EXPECT().GetRecentTrades(50).Return([]connector.Trade{
					{ID: "1", Symbol: "BTC/USDT"},
					{ID: "2", Symbol: "BTC/USDT"},
				})

				resp, err := client.Get("http://unix/api/trades")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				var trades []connector.Trade
				err = json.NewDecoder(resp.Body).Decode(&trades)
				Expect(err).NotTo(HaveOccurred())
				Expect(trades).To(HaveLen(2))
			})

			It("should respect limit parameter", func() {
				viewRegistry.EXPECT().GetRecentTrades(10).Return([]connector.Trade{
					{ID: "1"},
				})

				resp, err := client.Get("http://unix/api/trades?limit=10")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should return empty array when nil", func() {
				viewRegistry.EXPECT().GetRecentTrades(50).Return(nil)

				resp, err := client.Get("http://unix/api/trades")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				var trades []connector.Trade
				err = json.NewDecoder(resp.Body).Decode(&trades)
				Expect(err).NotTo(HaveOccurred())
				Expect(trades).To(HaveLen(0))
			})
		})

		Describe("/api/metrics", func() {
			It("should return metrics", func() {
				viewRegistry.EXPECT().GetMetrics().Return(&pkgmonitoring.StrategyMetrics{
					StrategyName: "momentum",
					Status:       "running",
				})

				resp, err := client.Get("http://unix/api/metrics")
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				var metrics pkgmonitoring.StrategyMetrics
				err = json.NewDecoder(resp.Body).Decode(&metrics)
				Expect(err).NotTo(HaveOccurred())
				Expect(metrics.StrategyName).To(Equal("momentum"))
			})
		})

		Describe("Method not allowed", func() {
			DescribeTable("should reject POST requests",
				func(endpoint string) {
					resp, err := client.Post("http://unix"+endpoint, "application/json", nil)
					Expect(err).NotTo(HaveOccurred())
					defer resp.Body.Close()

					Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
				},
				Entry("/health", "/health"),
				Entry("/api/pnl", "/api/pnl"),
				Entry("/api/positions", "/api/positions"),
				Entry("/api/trades", "/api/trades"),
				Entry("/api/metrics", "/api/metrics"),
			)
		})
	})
})
