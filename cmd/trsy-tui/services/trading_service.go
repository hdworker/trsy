package services

import (
	"fmt"
	"math/rand"
	"time"

	"trsy-tui/models"
)

// TradingService manages demo trading operations
type TradingService struct {
	trades    []models.Trade
	portfolio models.Portfolio
}

// NewTradingService creates a new trading service with initial balance
func NewTradingService(initialBalance float64) *TradingService {
	return &TradingService{
		trades: make([]models.Trade, 0),
		portfolio: models.Portfolio{
			InitialBalance: initialBalance,
			CurrentBalance: initialBalance,
			UpdatedAt:      time.Now(),
		},
	}
}

// OpenTrade opens a new demo trade position
func (s *TradingService) OpenTrade(symbol string, side models.TradeSide, quantity float64, entryPrice float64, leverage int, stopLoss, takeProfit *float64) (*models.Trade, error) {
	if quantity <= 0 || entryPrice <= 0 {
		return nil, fmt.Errorf("invalid quantity or price")
	}

	margin := (entryPrice * quantity) / float64(leverage)
	if margin > s.portfolio.CurrentBalance {
		return nil, fmt.Errorf("insufficient balance: required %.2f, available %.2f", margin, s.portfolio.CurrentBalance)
	}

	trade := models.Trade{
		ID:         generateTradeID(),
		Symbol:     symbol,
		Side:       side,
		Type:       models.TradeTypeMarket,
		Quantity:   quantity,
		EntryPrice: entryPrice,
		StopLoss:   stopLoss,
		TakeProfit: takeProfit,
		Leverage:   leverage,
		Status:     models.StatusOpen,
		OpenedAt:   time.Now(),
		Commission: 0, // Demo trading has no commission
	}

	s.trades = append(s.trades, trade)
	s.portfolio.UpdateFromTrades(s.trades)

	return &trade, nil
}

// CloseTrade closes an existing trade position
func (s *TradingService) CloseTrade(tradeID string, exitPrice float64) (*models.Trade, error) {
	for i, trade := range s.trades {
		if trade.ID == tradeID && trade.Status == models.StatusOpen {
			now := time.Now()
			s.trades[i].ExitPrice = exitPrice
			s.trades[i].Status = models.StatusClosed
			s.trades[i].ClosedAt = &now

			// Calculate final PnL
			if trade.Side == models.SideBuy {
				s.trades[i].PnL = (exitPrice - trade.EntryPrice) * trade.Quantity
			} else {
				s.trades[i].PnL = (trade.EntryPrice - exitPrice) * trade.Quantity
			}
			s.trades[i].PnL *= float64(trade.Leverage)

			margin := (trade.EntryPrice * trade.Quantity) / float64(trade.Leverage)
			if margin > 0 {
				s.trades[i].PnLPercent = (s.trades[i].PnL / margin) * 100
			}

			s.portfolio.UpdateFromTrades(s.trades)
			return &s.trades[i], nil
		}
	}

	return nil, fmt.Errorf("trade not found or already closed: %s", tradeID)
}

// UpdateTradesPnL updates PnL for all open trades based on current price
func (s *TradingService) UpdateTradesPnL(symbol string, currentPrice float64) {
	for i := range s.trades {
		if s.trades[i].Symbol == symbol && s.trades[i].Status == models.StatusOpen {
			s.trades[i].CalculatePnL(currentPrice)
		}
	}
	s.portfolio.UpdateFromTrades(s.trades)
}

// GetOpenTrades returns all open trades
func (s *TradingService) GetOpenTrades() []models.Trade {
	openTrades := make([]models.Trade, 0)
	for _, trade := range s.trades {
		if trade.Status == models.StatusOpen {
			openTrades = append(openTrades, trade)
		}
	}
	return openTrades
}

// GetClosedTrades returns all closed trades
func (s *TradingService) GetClosedTrades() []models.Trade {
	closedTrades := make([]models.Trade, 0)
	for _, trade := range s.trades {
		if trade.Status == models.StatusClosed {
			closedTrades = append(closedTrades, trade)
		}
	}
	return closedTrades
}

// GetAllTrades returns all trades
func (s *TradingService) GetAllTrades() []models.Trade {
	return s.trades
}

// GetPortfolio returns the current portfolio state
func (s *TradingService) GetPortfolio() models.Portfolio {
	return s.portfolio
}

// GetTradeByID returns a specific trade by ID
func (s *TradingService) GetTradeByID(tradeID string) *models.Trade {
	for i := range s.trades {
		if s.trades[i].ID == tradeID {
			return &s.trades[i]
		}
	}
	return nil
}

// CheckStopLossTakeProfit checks and automatically closes trades that hit SL/TP
func (s *TradingService) CheckStopLossTakeProfit(symbol string, currentPrice float64) []*models.Trade {
	closedTrades := make([]*models.Trade, 0)

	for i := range s.trades {
		trade := &s.trades[i]
		if trade.Symbol == symbol && trade.Status == models.StatusOpen {
			shouldClose := false

			// Check Stop Loss
			if trade.StopLoss != nil {
				if trade.Side == models.SideBuy && currentPrice <= *trade.StopLoss {
					shouldClose = true
				} else if trade.Side == models.SideSell && currentPrice >= *trade.StopLoss {
					shouldClose = true
				}
			}

			// Check Take Profit
			if trade.TakeProfit != nil && !shouldClose {
				if trade.Side == models.SideBuy && currentPrice >= *trade.TakeProfit {
					shouldClose = true
				} else if trade.Side == models.SideSell && currentPrice <= *trade.TakeProfit {
					shouldClose = true
				}
			}

			if shouldClose {
				closedTrade, err := s.CloseTrade(trade.ID, currentPrice)
				if err == nil {
					closedTrades = append(closedTrades, closedTrade)
				}
			}
		}
	}

	return closedTrades
}

func generateTradeID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("TRD-%d", rand.Intn(100000))
}
