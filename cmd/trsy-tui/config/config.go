package config

import "time"

// Config holds application configuration
type Config struct {
	// API Configuration
	APIKey    string
	APISecret string
	
	// Trading Configuration
	DefaultSymbol    string
	DefaultCategory  string
	DefaultLeverage  int
	DefaultQuantity  float64
	
	// Demo Trading Configuration
	DemoInitialBalance float64
	
	// Update Intervals
	MarketDataInterval time.Duration
	PositionInterval   time.Duration
	SignalInterval     time.Duration
	
	// Chart Configuration
	ChartWidth  int
	ChartHeight int
	
	// Risk Management
	MaxPositions      int
	MaxPositionSize   float64
	StopLossPercent   float64
	TakeProfitPercent float64
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultSymbol:        "BTCUSDT",
		DefaultCategory:      "linear",
		DefaultLeverage:      10,
		DefaultQuantity:      0.01,
		DemoInitialBalance:   10000.0,
		MarketDataInterval:   5 * time.Second,
		PositionInterval:     10 * time.Second,
		SignalInterval:       30 * time.Second,
		ChartWidth:           80,
		ChartHeight:          20,
		MaxPositions:         5,
		MaxPositionSize:      1.0,
		StopLossPercent:      2.0,
		TakeProfitPercent:    5.0,
	}
}
