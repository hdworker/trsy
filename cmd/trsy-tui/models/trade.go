package models

import "time"

// TradeSide represents buy or sell
type TradeSide string

const (
	SideBuy  TradeSide = "Buy"
	SideSell TradeSide = "Sell"
)

// TradeType represents the type of trade
type TradeType string

const (
	TradeTypeMarket TradeType = "Market"
	TradeTypeLimit  TradeType = "Limit"
)

// TradeStatus represents the status of a trade
type TradeStatus string

const (
	StatusOpen     TradeStatus = "Open"
	StatusClosed   TradeStatus = "Closed"
	StatusPending  TradeStatus = "Pending"
	StatusCancelled TradeStatus = "Cancelled"
)

// Trade represents a demo trade position
type Trade struct {
	ID           string
	Symbol       string
	Side         TradeSide
	Type         TradeType
	Quantity     float64
	EntryPrice   float64
	ExitPrice    float64
	StopLoss     *float64
	TakeProfit   *float64
	Leverage     int
	Status       TradeStatus
	PnL          float64
	PnLPercent   float64
	OpenedAt     time.Time
	ClosedAt     *time.Time
	Commission   float64
	Notes        string
}

// CalculatePnL calculates profit and loss for the trade
func (t *Trade) CalculatePnL(currentPrice float64) {
	if t.Status == StatusClosed {
		return
	}

	t.PnL = 0
	if t.Side == SideBuy {
		t.PnL = (currentPrice - t.EntryPrice) * t.Quantity
	} else {
		t.PnL = (t.EntryPrice - currentPrice) * t.Quantity
	}

	// Apply leverage
	t.PnL *= float64(t.Leverage)

	// Calculate percentage
	margin := (t.EntryPrice * t.Quantity) / float64(t.Leverage)
	if margin > 0 {
		t.PnLPercent = (t.PnL / margin) * 100
	}
}

// Portfolio represents the demo trading portfolio
type Portfolio struct {
	InitialBalance float64
	CurrentBalance float64
	TotalPnL       float64
	TotalPnLPercent float64
	OpenTrades     int
	ClosedTrades   int
	WinRate        float64
	Wins           int
	Losses         int
	UpdatedAt      time.Time
}

// UpdateFromTrades updates portfolio statistics from trades
func (p *Portfolio) UpdateFromTrades(trades []Trade) {
	p.OpenTrades = 0
	p.ClosedTrades = 0
	p.Wins = 0
	p.Losses = 0
	totalPnL := 0.0

	for _, t := range trades {
		if t.Status == StatusOpen {
			p.OpenTrades++
		} else if t.Status == StatusClosed {
			p.ClosedTrades++
			totalPnL += t.PnL
			if t.PnL > 0 {
				p.Wins++
			} else if t.PnL < 0 {
				p.Losses++
			}
		}
	}

	p.CurrentBalance = p.InitialBalance + totalPnL
	p.TotalPnL = totalPnL
	if p.InitialBalance > 0 {
		p.TotalPnLPercent = (totalPnL / p.InitialBalance) * 100
	}

	totalClosed := p.Wins + p.Losses
	if totalClosed > 0 {
		p.WinRate = (float64(p.Wins) / float64(totalClosed)) * 100
	}

	p.UpdatedAt = time.Now()
}

// Signal represents a trading signal
type Signal struct {
	ID            string
	Symbol        string
	Direction     TradeSide
	Strength      float64 // 0-100
	Indicators    map[string]float64
	Price         float64
	Timestamp     time.Time
	IsActive      bool
	ExecutedTradeID *string
	Notes         string
}

// ChartData represents data for chart visualization
type ChartData struct {
	Symbol     string
	Interval   string
	Timestamps []time.Time
	Open       []float64
	High       []float64
	Low        []float64
	Close      []float64
	Volume     []float64
	
	// Indicators
	SMA20      []float64
	SMA50      []float64
	RSI        []float64
	MACD       []float64
	MACDSignal []float64
	MACDHist   []float64
	BollingerUpper []float64
	BollingerLower []float64
}
