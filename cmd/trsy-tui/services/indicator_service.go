package services

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"trsy-tui/models"
)

// IndicatorService provides technical analysis indicators
type IndicatorService struct{}

// NewIndicatorService creates a new indicator service
func NewIndicatorService() *IndicatorService {
	return &IndicatorService{}
}

// CalculateSMA calculates Simple Moving Average
func (s *IndicatorService) CalculateSMA(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	
	result := make([]float64, len(values)-period+1)
	for i := 0; i <= len(values)-period; i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += values[i+j]
		}
		result[i] = sum / float64(period)
	}
	return result
}

// CalculateEMA calculates Exponential Moving Average
func (s *IndicatorService) CalculateEMA(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	
	multiplier := 2.0 / float64(period+1)
	result := make([]float64, len(values))
	
	// First EMA is SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	result[period-1] = sum / float64(period)
	
	// Calculate rest
	for i := period; i < len(values); i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}
	
	return result[period-1:]
}

// CalculateRSI calculates Relative Strength Index
func (s *IndicatorService) CalculateRSI(values []float64, period int) []float64 {
	if len(values) < period+1 {
		return nil
	}
	
	// Calculate changes
	changes := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		changes[i-1] = values[i] - values[i-1]
	}
	
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	
	for i, change := range changes {
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}
	
	// First average gain/loss
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)
	
	rsi := make([]float64, len(changes)-period+1)
	
	// First RSI
	if avgLoss == 0 {
		rsi[0] = 100
	} else {
		rs := avgGain / avgLoss
		rsi[0] = 100 - (100 / (1 + rs))
	}
	
	// Calculate rest
	for i := 1; i < len(rsi); i++ {
		avgGain = (avgGain*float64(period-1) + gains[period+i-1]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[period+i-1]) / float64(period)
		
		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}
	
	return rsi
}

// CalculateMACD calculates MACD indicator
func (s *IndicatorService) CalculateMACD(values []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
	if len(values) < slowPeriod {
		return nil, nil, nil
	}
	
	fastEMA := s.CalculateEMA(values, fastPeriod)
	slowEMA := s.CalculateEMA(values, slowPeriod)
	
	// Align lengths
	minLen := len(fastEMA)
	if len(slowEMA) < minLen {
		minLen = len(slowEMA)
	}
	
	macdLine := make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		macdLine[i] = fastEMA[len(fastEMA)-minLen+i] - slowEMA[len(slowEMA)-minLen+i]
	}
	
	signalLine := s.CalculateEMA(macdLine, signalPeriod)
	if signalLine == nil {
		return macdLine, nil, nil
	}
	
	histogram := make([]float64, len(signalLine))
	for i := 0; i < len(signalLine); i++ {
		histogram[i] = macdLine[len(macdLine)-len(signalLine)+i] - signalLine[i]
	}
	
	return macdLine, signalLine, histogram
}

// CalculateBollingerBands calculates Bollinger Bands
func (s *IndicatorService) CalculateBollingerBands(values []float64, period int, stdDev float64) ([]float64, []float64) {
	if len(values) < period {
		return nil, nil
	}
	
	sma := s.CalculateSMA(values, period)
	if sma == nil {
		return nil, nil
	}
	
	upper := make([]float64, len(sma))
	lower := make([]float64, len(sma))
	
	for i := range sma {
		// Calculate standard deviation
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := values[i+j] - sma[i]
			sum += diff * diff
		}
		std := math.Sqrt(sum / float64(period))
		
		upper[i] = sma[i] + (std * stdDev)
		lower[i] = sma[i] - (std * stdDev)
	}
	
	return upper, lower
}

// CalculateStdev calculates standard deviation
func (s *IndicatorService) CalculateStdev(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	
	result := make([]float64, len(values)-period+1)
	for i := 0; i <= len(values)-period; i++ {
		sum := 0.0
		mean := 0.0
		for j := 0; j < period; j++ {
			mean += values[i+j]
		}
		mean /= float64(period)
		
		for j := 0; j < period; j++ {
			diff := values[i+j] - mean
			sum += diff * diff
		}
		result[i] = math.Sqrt(sum / float64(period))
	}
	return result
}

// GenerateSignal generates a trading signal based on indicators
func (s *IndicatorService) GenerateSignal(symbol string, chartData *models.ChartData) *models.Signal {
	if len(chartData.Close) < 50 {
		return nil
	}
	
	latestClose := chartData.Close[len(chartData.Close)-1]
	latestRSI := chartData.RSI[len(chartData.RSI)-1]
	latestMACD := chartData.MACD[len(chartData.MACD)-1]
	latestMACDSignal := chartData.MACDSignal[len(chartData.MACDSignal)-1]
	latestSMA20 := chartData.SMA20[len(chartData.SMA20)-1]
	
	strength := 0.0
	direction := models.SideBuy
	reasons := []string{}
	
	// RSI signals
	if latestRSI < 30 {
		strength += 30
		reasons = append(reasons, "RSI oversold")
	} else if latestRSI > 70 {
		strength -= 30
		direction = models.SideSell
		reasons = append(reasons, "RSI overbought")
	}
	
	// MACD signals
	if latestMACD > latestMACDSignal {
		strength += 25
		reasons = append(reasons, "MACD bullish crossover")
	} else {
		strength -= 25
		direction = models.SideSell
		reasons = append(reasons, "MACD bearish crossover")
	}
	
	// SMA trend
	if latestClose > latestSMA20 {
		strength += 20
		reasons = append(reasons, "Price above SMA20")
	} else {
		strength -= 20
		direction = models.SideSell
		reasons = append(reasons, "Price below SMA20")
	}
	
	// Normalize strength to 0-100
	if strength < 0 {
		strength = 0
	} else if strength > 100 {
		strength = 100
	}
	
	notes := fmt.Sprintf("Strength: %.1f | %s", strength, joinStrings(reasons, ", "))
	
	return &models.Signal{
		ID:         generateID(),
		Symbol:     symbol,
		Direction:  direction,
		Strength:   strength,
		Price:      latestClose,
		Timestamp:  time.Now(),
		IsActive:   true,
		Indicators: map[string]float64{
			"RSI":    latestRSI,
			"MACD":   latestMACD,
			"SMA20":  latestSMA20,
		},
		Notes: notes,
	}
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func generateID() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("SIG-%d", rand.Intn(100000))
}

// ProcessChartData processes raw price data and calculates all indicators
func (s *IndicatorService) ProcessChartData(closePrices []float64) *models.ChartData {
	if len(closePrices) < 50 {
		return nil
	}
	
	chartData := &models.ChartData{}
	
	// Calculate SMA20 and SMA50
	chartData.SMA20 = s.CalculateSMA(closePrices, 20)
	chartData.SMA50 = s.CalculateSMA(closePrices, 50)
	
	// Calculate RSI
	chartData.RSI = s.CalculateRSI(closePrices, 14)
	
	// Calculate MACD
	chartData.MACD, chartData.MACDSignal, chartData.MACDHist = s.CalculateMACD(closePrices, 12, 26, 9)
	
	// Calculate Bollinger Bands
	chartData.BollingerUpper, chartData.BollingerLower = s.CalculateBollingerBands(closePrices, 20, 2.0)
	
	return chartData
}
