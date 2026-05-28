package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"trsy-tui/models"
)

var (
	profitStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	lossStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	boldStyle   = lipgloss.NewStyle().Bold(true)
)

// RenderPortfolio renders portfolio summary
func RenderPortfolio(portfolio models.Portfolio, width int) string {
	var b strings.Builder

	b.WriteString(boldStyle.Render("📊 Portfolio Summary\n"))
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	balanceStr := fmt.Sprintf("Balance: $%.2f", portfolio.CurrentBalance)
	pnlStr := fmt.Sprintf("Total PnL: $%.2f (%.2f%%)", portfolio.TotalPnL, portfolio.TotalPnLPercent)
	
	if portfolio.TotalPnL >= 0 {
		pnlStr = profitStyle.Render(pnlStr)
	} else {
		pnlStr = lossStyle.Render(pnlStr)
	}

	b.WriteString(fmt.Sprintf("%-30s %s\n", balanceStr, pnlStr))
	b.WriteString(fmt.Sprintf("Open Trades: %-10d Closed Trades: %d\n", portfolio.OpenTrades, portfolio.ClosedTrades))
	b.WriteString(fmt.Sprintf("Wins: %-12d Losses: %-10d Win Rate: %.1f%%\n", 
		portfolio.Wins, portfolio.Losses, portfolio.WinRate))

	return b.String()
}

// RenderTrade renders a single trade
func RenderTrade(trade models.Trade, width int) string {
	var b strings.Builder

	statusIcon := "🟢"
	if trade.Status == models.StatusClosed {
		statusIcon = "⚪"
	}

	b.WriteString(fmt.Sprintf("%s %s %s\n", statusIcon, trade.Symbol, trade.Side))
	b.WriteString(fmt.Sprintf("  ID: %s\n", trade.ID))
	b.WriteString(fmt.Sprintf("  Entry: $%.2f | Exit: $%.2f\n", trade.EntryPrice, trade.ExitPrice))
	b.WriteString(fmt.Sprintf("  Qty: %.4f | Lev: %dx\n", trade.Quantity, trade.Leverage))

	if trade.Status == models.StatusOpen {
		pnlStr := fmt.Sprintf("  PnL: $%.2f (%.2f%%)", trade.PnL, trade.PnLPercent)
		if trade.PnL >= 0 {
			pnlStr = profitStyle.Render(pnlStr)
		} else {
			pnlStr = lossStyle.Render(pnlStr)
		}
		b.WriteString(pnlStr + "\n")
		
		if trade.StopLoss != nil {
			b.WriteString(fmt.Sprintf("  SL: $%.2f", *trade.StopLoss))
		}
		if trade.TakeProfit != nil {
			b.WriteString(fmt.Sprintf(" | TP: $%.2f", *trade.TakeProfit))
		}
		b.WriteString("\n")
	} else {
		pnlStr := fmt.Sprintf("  Final PnL: $%.2f (%.2f%%)", trade.PnL, trade.PnLPercent)
		if trade.PnL >= 0 {
			pnlStr = profitStyle.Render(pnlStr)
		} else {
			pnlStr = lossStyle.Render(pnlStr)
		}
		b.WriteString(pnlStr + "\n")
	}

	b.WriteString(infoStyle.Render(fmt.Sprintf("  Opened: %s\n", trade.OpenedAt.Format("2006-01-02 15:04"))))
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	return b.String()
}

// RenderSignal renders a trading signal
func RenderSignal(signal models.Signal, width int) string {
	var b strings.Builder

	directionIcon := "📈"
	if signal.Direction == models.SideSell {
		directionIcon = "📉"
	}

	strengthColor := "#00FF00"
	if signal.Strength < 40 {
		strengthColor = "#FF5555"
	} else if signal.Strength < 70 {
		strengthColor = "#FFFF00"
	}

	b.WriteString(fmt.Sprintf("%s %s Signal\n", directionIcon, signal.Symbol))
	b.WriteString(fmt.Sprintf("  Direction: %s\n", signal.Direction))
	
	strengthStr := fmt.Sprintf("  Strength: %.1f/100", signal.Strength)
	strengthStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(strengthColor))
	b.WriteString(strengthStyle.Render(strengthStr) + "\n")
	
	b.WriteString(fmt.Sprintf("  Price: $%.2f\n", signal.Price))
	b.WriteString(fmt.Sprintf("  Time: %s\n", signal.Timestamp.Format("15:04:05")))
	
	if len(signal.Indicators) > 0 {
		b.WriteString("  Indicators:\n")
		for name, value := range signal.Indicators {
			b.WriteString(fmt.Sprintf("    %s: %.2f\n", name, value))
		}
	}
	
	if signal.Notes != "" {
		b.WriteString(infoStyle.Render(fmt.Sprintf("  Note: %s\n", signal.Notes)))
	}

	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	return b.String()
}

// RenderChartASCII renders a simple ASCII chart
func RenderChartASCII(data []float64, width, height int, title string) string {
	if len(data) == 0 {
		return "No data available\n"
	}

	// Find min/max
	minVal := data[0]
	maxVal := data[0]
	for _, v := range data {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	if rangeVal == 0 {
		rangeVal = 1
	}

	var b strings.Builder
	b.WriteString(boldStyle.Render(title + "\n"))
	b.WriteString(strings.Repeat("─", width))
	b.WriteString("\n")

	// Create chart grid
	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot data points
	step := float64(len(data)) / float64(width)
	for x := 0; x < width && int(float64(x)*step) < len(data); x++ {
		idx := int(float64(x) * step)
		if idx >= len(data) {
			idx = len(data) - 1
		}
		
		normalized := (data[idx] - minVal) / rangeVal
		y := int((1.0 - normalized) * float64(height-1))
		if y < 0 {
			y = 0
		}
		if y >= height {
			y = height - 1
		}
		
		grid[y][x] = '█'
	}

	// Add price labels
	priceRange := fmt.Sprintf("$%.2f - $%.2f", minVal, maxVal)
	b.WriteString(infoStyle.Render(priceRange) + "\n")

	// Render grid
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			b.WriteRune(grid[i][j])
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderIndicatorValue renders an indicator value with formatting
func RenderIndicatorValue(name string, value float64, thresholds ...float64) string {
	style := infoStyle
	
	if len(thresholds) == 2 {
		if value < thresholds[0] {
			style = lossStyle // Oversold
		} else if value > thresholds[1] {
			style = profitStyle // Overbought
		}
	}
	
	return style.Render(fmt.Sprintf("%s: %.2f", name, value))
}
