package monitoring_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"

	"github.com/backtesting-org/kronos-cli/internal/services/monitoring"
	pkgmonitoring "github.com/backtesting-org/kronos-cli/pkg/monitoring"
	"github.com/backtesting-org/kronos-sdk/pkg/types/connector"
	"github.com/backtesting-org/kronos-sdk/pkg/types/health"
	"github.com/backtesting-org/kronos-sdk/pkg/types/kronos/numerical"
	"github.com/backtesting-org/kronos-sdk/pkg/types/portfolio"
	"github.com/backtesting-org/kronos-sdk/pkg/types/strategy"

	healthMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/health"
	kronosMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/kronos"
	activityMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/kronos/activity"
	analyticsMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/kronos/analytics"
	registryMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/registry"
	strategyMock "github.com/backtesting-org/kronos-sdk/mocks/github.com/backtesting-org/kronos-sdk/pkg/types/strategy"
)

var _ = Describe("ViewRegistry", func() {
	var (
		mockKronos           *kronosMock.Kronos
		mockHealthStore      *healthMock.HealthStore
		mockStrategyRegistry *registryMock.StrategyRegistry
		mockActivity         *activityMock.Activity
		//mockPnl              *activityMock.PNL
		mockPositions *activityMock.Positions
		mockMarket    *analyticsMock.Market
		registry      pkgmonitoring.ViewRegistry
	)

	BeforeEach(func() {
		mockKronos = kronosMock.NewKronos(GinkgoT())
		mockHealthStore = healthMock.NewHealthStore(GinkgoT())
		mockStrategyRegistry = registryMock.NewStrategyRegistry(GinkgoT())
		mockActivity = activityMock.NewActivity(GinkgoT())
		//mockPnl = activityMock.NewPNL(GinkgoT())
		mockPositions = activityMock.NewPositions(GinkgoT())
		mockMarket = analyticsMock.NewMarket(GinkgoT())

		registry = monitoring.NewViewRegistry(mockHealthStore, mockKronos, mockStrategyRegistry, nil)
	})

	Describe("GetHealth", func() {
		It("should return health report from health store", func() {
			expectedReport := &health.SystemHealthReport{
				OverallState: health.StateConnected,
				HasErrors:    false,
			}
			mockHealthStore.EXPECT().GetSystemHealth().Return(expectedReport)

			result := registry.GetHealth()

			Expect(result).To(Equal(expectedReport))
		})
	})

	//Describe("GetPnLView", func() {
	//	It("should return PnL from activity", func() {
	//		mockKronos.EXPECT().Activity().Return(mockActivity)
	//		mockActivity.EXPECT().PNL().Return(mockPnl)
	//
	//		result := registry.GetPnLView()
	//
	//		Expect(result).To(Equal("pnl-data"))
	//	})
	//})

	Describe("GetPositionsView", func() {
		Context("when strategy exists", func() {
			It("should return strategy execution", func() {
				mockStrategy := strategyMock.NewStrategy(GinkgoT())
				expectedExecution := &strategy.StrategyExecution{
					Orders: []connector.Order{},
					Trades: []connector.Trade{},
				}

				mockStrategy.EXPECT().GetName().Return(strategy.StrategyName("test-strategy"))
				mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{mockStrategy})
				mockKronos.EXPECT().Activity().Return(mockActivity)
				mockActivity.EXPECT().Positions().Return(mockPositions)
				mockPositions.EXPECT().GetStrategyExecution(mock.Anything, strategy.StrategyName("test-strategy")).Return(expectedExecution)

				result := registry.GetPositionsView()

				Expect(result).To(Equal(expectedExecution))
			})
		})

		Context("when no strategy exists", func() {
			It("should return nil", func() {
				mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{})

				result := registry.GetPositionsView()

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("GetOrderbookView", func() {
		It("should return orderbook for symbol", func() {
			asset := portfolio.NewAsset("BTC/USDT")
			expectedOrderbook := &connector.OrderBook{
				Bids: []connector.PriceLevel{
					{Price: numerical.NewFromFloat(42000), Quantity: numerical.NewFromFloat(1.5)},
				},
				Asks: []connector.PriceLevel{
					{Price: numerical.NewFromFloat(42001), Quantity: numerical.NewFromFloat(1.2)},
				},
			}

			mockKronos.EXPECT().Asset("BTC/USDT").Return(asset)
			mockKronos.EXPECT().Market().Return(mockMarket)
			mockMarket.EXPECT().OrderBook(mock.Anything, asset).Return(expectedOrderbook, nil)

			result := registry.GetOrderbookView("BTC/USDT")

			Expect(result).To(Equal(expectedOrderbook))
		})

		It("should return nil on error", func() {
			asset := portfolio.NewAsset("BTC/USDT")

			mockKronos.EXPECT().Asset("BTC/USDT").Return(asset)
			mockKronos.EXPECT().Market().Return(mockMarket)
			mockMarket.EXPECT().OrderBook(mock.Anything, asset).Return(nil, fmt.Errorf("not found"))

			result := registry.GetOrderbookView("BTC/USDT")

			Expect(result).To(BeNil())
		})
	})

	Describe("GetRecentTrades", func() {
		Context("when strategy exists", func() {
			It("should return trades", func() {
				mockStrategy := strategyMock.NewStrategy(GinkgoT())
				expectedTrades := []connector.Trade{
					{ID: "trade-1", Symbol: "BTC/USDT"},
					{ID: "trade-2", Symbol: "BTC/USDT"},
				}

				mockStrategy.EXPECT().GetName().Return(strategy.StrategyName("test-strategy"))
				mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{mockStrategy})
				mockKronos.EXPECT().Activity().Return(mockActivity)
				mockActivity.EXPECT().Positions().Return(mockPositions)
				mockPositions.EXPECT().GetTradesForStrategy(mock.Anything, strategy.StrategyName("test-strategy")).Return(expectedTrades)

				result := registry.GetRecentTrades(10)

				Expect(result).To(Equal(expectedTrades))
			})

			It("should limit trades when more than limit", func() {
				mockStrategy := strategyMock.NewStrategy(GinkgoT())
				allTrades := []connector.Trade{
					{ID: "trade-1"},
					{ID: "trade-2"},
					{ID: "trade-3"},
					{ID: "trade-4"},
					{ID: "trade-5"},
				}

				mockStrategy.EXPECT().GetName().Return(strategy.StrategyName("test-strategy"))
				mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{mockStrategy})
				mockKronos.EXPECT().Activity().Return(mockActivity)
				mockActivity.EXPECT().Positions().Return(mockPositions)
				mockPositions.EXPECT().GetTradesForStrategy(mock.Anything, strategy.StrategyName("test-strategy")).Return(allTrades)

				result := registry.GetRecentTrades(2)

				Expect(result).To(HaveLen(2))
				Expect(result[0].ID).To(Equal("trade-4"))
				Expect(result[1].ID).To(Equal("trade-5"))
			})
		})

		Context("when no strategy exists", func() {
			It("should return nil", func() {
				mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{})

				result := registry.GetRecentTrades(10)

				Expect(result).To(BeNil())
			})
		})
	})

	Describe("GetMetrics", func() {
		It("should return strategy metrics", func() {
			mockStrategy := strategyMock.NewStrategy(GinkgoT())
			mockStrategy.EXPECT().GetName().Return(strategy.StrategyName("test-strategy"))
			mockStrategyRegistry.EXPECT().GetAllStrategies().Return([]strategy.Strategy{mockStrategy})

			result := registry.GetMetrics()

			Expect(result.StrategyName).To(Equal("test-strategy"))
			Expect(result.Status).To(Equal("running"))
		})
	})

	Describe("GetProfilingStats", func() {
		It("should return nil when profiling store is nil", func() {
			result := registry.GetProfilingStats()
			Expect(result).To(BeNil())
		})
	})

	Describe("GetRecentExecutions", func() {
		It("should return nil when profiling store is nil", func() {
			result := registry.GetRecentExecutions(10)
			Expect(result).To(BeNil())
		})
	})
})
