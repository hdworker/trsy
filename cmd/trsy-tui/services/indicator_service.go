package services

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cinar/indicator"
	"trsy-tui/models"
)

// IndicatorService provides technical analysis indicators using cinar/indicator library
type IndicatorService struct{}

// NewIndicatorService creates a new indicator service
func NewIndicatorService() *IndicatorService {
	return &IndicatorService{}
}

// CalculateSMA calculates Simple Moving Average using cinar/indicator
func (s *IndicatorService) CalculateSMA(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	return indicator.Sma(period, values)
}

// CalculateEMA calculates Exponential Moving Average using cinar/indicator
func (s *IndicatorService) CalculateEMA(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	return indicator.Ema(period, values)
}

// CalculateRSI calculates Relative Strength Index using cinar/indicator
func (s *IndicatorService) CalculateRSI(values []float64, period int) []float64 {
	if len(values) < period+1 {
		return nil
	}
	// RSI in cinar/indicator expects closing prices and returns RSI values
	rso := indicator.Rso(period, values)
	// Convert RSO to RSI: RSI = 100 - (100 / (1 + RSO))
	rsi := make([]float64, len(rso))
	for i, v := range rso {
		if v == 0 {
			rsi[i] = 100
		} else {
			rsi[i] = 100 - (100 / (1 + v))
		}
	}
	return rsi
}

// CalculateMACD calculates MACD indicator using cinar/indicator
func (s *IndicatorService) CalculateMACD(values []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
	if len(values) < slowPeriod {
		return nil, nil, nil
	}
	macdLine, signalLine := indicator.Macd(fastPeriod, slowPeriod, signalPeriod, values)
	
	// Calculate histogram
	histogram := make([]float64, len(macdLine))
	minLen := len(signalLine)
	if len(macdLine) < minLen {
		minLen = len(macdLine)
	}
	for i := 0; i < minLen; i++ {
		histogram[i] = macdLine[len(macdLine)-minLen+i] - signalLine[len(signalLine)-minLen+i]
	}
	
	return macdLine, signalLine, histogram
}

// CalculateBollingerBands calculates Bollinger Bands using cinar/indicator
func (s *IndicatorService) CalculateBollingerBands(values []float64, period int, stdDev float64) ([]float64, []float64) {
	if len(values) < period {
		return nil, nil
	}
	upper, middle := indicator.BollingerBands(stdDev, period, values)
	lower := make([]float64, len(upper))
	for i := range upper {
		lower[i] = 2*middle[i] - upper[i]
	}
	return upper, lower
}

// CalculateStdev calculates standard deviation using cinar/indicator
func (s *IndicatorService) CalculateStdev(values []float64, period int) []float64 {
	if len(values) < period {
		return nil
	}
	return indicator.StdDev(period, values)
}

// CalculateATR calculates Average True Range using cinar/indicator
func (s *IndicatorService) CalculateATR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period || len(lows) < period || len(closes) < period {
		return nil
	}
	return indicator.Atr(period, highs, lows, closes)
}

// CalculateADX calculates Average Directional Index using cinar/indicator
func (s *IndicatorService) CalculateADX(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period || len(lows) < period || len(closes) < period {
		return nil
	}
	return indicator.Adx(period, highs, lows, closes)
}

// CalculateCCI calculates Commodity Channel Index using cinar/indicator
func (s *IndicatorService) CalculateCCI(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period || len(lows) < period || len(closes) < period {
		return nil
	}
	return indicator.Cci(period, highs, lows, closes)
}

// CalculateStochastic calculates Stochastic Oscillator using cinar/indicator
func (s *IndicatorService) CalculateStochastic(highs, lows, closes []float64) ([]float64, []float64) {
	if len(highs) < 14 || len(lows) < 14 || len(closes) < 14 {
		return nil, nil
	}
	return indicator.Stoch(14, 3, highs, lows, closes)
}

// CalculateOBV calculates On-Balance Volume using cinar/indicator
func (s *IndicatorService) CalculateOBV(volumes []float64, closes []float64) []float64 {
	if len(volumes) < 2 || len(closes) < 2 {
		return nil
	}
	return indicator.Obv(volumes, closes)
}

// CalculateVWAP calculates Volume Weighted Average Price
func (s *IndicatorService) CalculateVWAP(highs, lows, closes, volumes []float64) []float64 {
	if len(highs) == 0 || len(volumes) == 0 {
		return nil
	}
	vwap := make([]float64, len(closes))
	cumulativeTPV := 0.0
	cumulativeVolume := 0.0
	
	for i := range closes {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3.0
		cumulativeTPV += typicalPrice * volumes[i]
		cumulativeVolume += volumes[i]
		vwap[i] = cumulativeTPV / cumulativeVolume
	}
	return vwap
}

// CalculateIchimoku calculates Ichimoku Cloud components
func (s *IndicatorService) CalculateIchimoku(highs, lows, closes []float64) (conversionLine, baseLine, leadingSpanA, leadingSpanB, laggingSpan []float64) {
	if len(highs) < 52 || len(lows) < 52 || len(closes) < 52 {
		return nil, nil, nil, nil, nil
	}
	
	// Tenkan-sen (Conversion Line): (9-period high + 9-period low)/2
	conversionLine = make([]float64, len(closes)-8)
	for i := 8; i < len(closes); i++ {
		high := highs[i-8]
		low := lows[i-8]
		for j := i - 8; j <= i; j++ {
			if highs[j] > high {
				high = highs[j]
			}
			if lows[j] < low {
				low = lows[j]
			}
		}
		conversionLine[i-8] = (high + low) / 2.0
	}
	
	// Kijun-sen (Base Line): (26-period high + 26-period low)/2
	baseLine = make([]float64, len(closes)-25)
	for i := 25; i < len(closes); i++ {
		high := highs[i-25]
		low := lows[i-25]
		for j := i - 25; j <= i; j++ {
			if highs[j] > high {
				high = highs[j]
			}
			if lows[j] < low {
				low = lows[j]
			}
		}
		baseLine[i-25] = (high + low) / 2.0
	}
	
	// Senkou Span A (Leading Span A): (Conversion Line + Base Line)/2
	minLen := len(conversionLine)
	if len(baseLine) < minLen {
		minLen = len(baseLine)
	}
	leadingSpanA = make([]float64, minLen)
	for i := 0; i < minLen; i++ {
		leadingSpanA[i] = (conversionLine[i] + baseLine[i]) / 2.0
	}
	
	// Senkou Span B (Leading Span B): (52-period high + 52-period low)/2
	leadingSpanB = make([]float64, len(closes)-51)
	for i := 51; i < len(closes); i++ {
		high := highs[i-51]
		low := lows[i-51]
		for j := i - 51; j <= i; j++ {
			if highs[j] > high {
				high = highs[j]
			}
			if lows[j] < low {
				low = lows[j]
			}
		}
		leadingSpanB[i-51] = (high + low) / 2.0
	}
	
	// Chikou Span (Lagging Span): Close shifted back 26 periods
	laggingSpan = make([]float64, len(closes)-26)
	for i := 26; i < len(closes); i++ {
		laggingSpan[i-26] = closes[i]
	}
	
	return conversionLine, baseLine, leadingSpanA, leadingSpanB, laggingSpan
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
