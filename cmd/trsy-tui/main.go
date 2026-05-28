package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hirokisan/bybit/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC"))
)

type tab int

const (
	tabMarket tab = iota
	tabOrderBook
	tabTickers
	tabPositions
)

func (t tab) String() string {
	return [...]string{"Market", "OrderBook", "Tickers", "Positions"}[t]
}

type model struct {
	client        *bybit.Client
	width         int
	height        int
	currentTab    tab
	tabs          []tab
	klines        []bybit.V5GetKlineItem
	orderbook     *bybit.V5GetOrderbookResponse
	tickers       []bybit.V5GetTickersLinearInverseItem
	positions     []bybit.V5GetPositionInfoItem
	err           error
	loading       bool
	statusMessage string
	symbol        bybit.SymbolV5
	category      bybit.CategoryV5
}

func initialModel() model {
	apiKey := os.Getenv("BYBIT_API_KEY")
	apiSecret := os.Getenv("BYBIT_API_SECRET")

	client := bybit.NewClient()
	if apiKey != "" && apiSecret != "" {
		client = client.WithAuth(apiKey, apiSecret)
	}

	return model{
		client:     client,
		tabs:       []tab{tabMarket, tabOrderBook, tabTickers, tabPositions},
		currentTab: tabMarket,
		symbol:     bybit.SymbolV5BTCUSDT,
		category:   bybit.CategoryV5Linear,
		loading:    true,
	}
}

func (m model) Init() tea.Cmd {
	return fetchMarketData(m.client, m.symbol, m.category)
}

func fetchMarketData(client *bybit.Client, symbol bybit.SymbolV5, category bybit.CategoryV5) tea.Cmd {
	return func() tea.Msg {
		marketService := client.V5().Market()

		klineParam := bybit.V5GetKlineParam{
			Category: category,
			Symbol:   symbol,
			Interval: bybit.Interval15,
			Limit:    IntPtr(20),
		}
		klineResp, err := marketService.GetKline(klineParam)
		if err != nil {
			return errorMsg{err: err}
		}

		orderbookParam := bybit.V5GetOrderbookParam{
			Category: category,
			Symbol:   symbol,
			Limit:    IntPtr(10),
		}
		orderbookResp, err := marketService.GetOrderbook(orderbookParam)
		if err != nil {
			return errorMsg{err: err}
		}

		tickerParam := bybit.V5GetTickersParam{
			Category: category,
			Symbol:   &symbol,
		}
		tickerResp, err := marketService.GetTickers(tickerParam)
		if err != nil {
			return errorMsg{err: err}
		}

		var tickerList []bybit.V5GetTickersLinearInverseItem
		if tickerResp.Result.LinearInverse != nil {
			tickerList = tickerResp.Result.LinearInverse.List
		}

		return marketDataMsg{
			klines:    klineResp.Result.List,
			orderbook: orderbookResp,
			tickers:   tickerList,
		}
	}
}

func fetchPositions(client *bybit.Client, category bybit.CategoryV5) tea.Cmd {
	return func() tea.Msg {
		positionService := client.V5().Position()

		param := bybit.V5GetPositionInfoParam{
			Category: category,
		}
		resp, err := positionService.GetPositionInfo(param)
		if err != nil {
			return errorMsg{err: err}
		}

		return positionsMsg{positions: resp.Result.List}
	}
}

func IntPtr(i int) *int {
	return &i
}

type marketDataMsg struct {
	klines    []bybit.V5GetKlineItem
	orderbook *bybit.V5GetOrderbookResponse
	tickers   []bybit.V5GetTickersLinearInverseItem
}

type positionsMsg struct {
	positions []bybit.V5GetPositionInfoItem
}

type errorMsg struct {
	err error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "right", "l":
			m.currentTab++
			if int(m.currentTab) >= len(m.tabs) {
				m.currentTab = 0
			}
			if m.currentTab == tabPositions {
				return m, fetchPositions(m.client, m.category)
			}
			return m, nil
		case "shift+tab", "left", "h":
			m.currentTab--
			if int(m.currentTab) < 0 {
				m.currentTab = tab(len(m.tabs) - 1)
			}
			return m, nil
		case "r":
			m.loading = true
			return m, fetchMarketData(m.client, m.symbol, m.category)
		}

	case marketDataMsg:
		m.klines = msg.klines
		m.orderbook = msg.orderbook
		m.tickers = msg.tickers
		m.loading = false
		m.statusMessage = fmt.Sprintf("Updated at %d", m.orderbook.Result.Timestamp)
		return m, nil

	case positionsMsg:
		m.positions = msg.positions
		m.loading = false
		return m, nil

	case errorMsg:
		m.err = msg.err
		m.loading = false
		m.statusMessage = fmt.Sprintf("Error: %v", msg.err)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("📊 TRSY TUI - ByBit Client (%s)", m.symbol)))
	b.WriteString("\n\n")

	tabs := make([]string, len(m.tabs))
	for i, t := range m.tabs {
		style := normalStyle
		if m.currentTab == t {
			style = selectedStyle
		}
		tabs[i] = style.Render(fmt.Sprintf("[%s] %s", t.String()[0:1], t.String()))
	}
	b.WriteString(strings.Join(tabs, "  "))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading...\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %v\n", m.err)))
		return b.String()
	}

	switch m.currentTab {
	case tabMarket:
		b.WriteString(m.viewMarket())
	case tabOrderBook:
		b.WriteString(m.viewOrderBook())
	case tabTickers:
		b.WriteString(m.viewTickers())
	case tabPositions:
		b.WriteString(m.viewPositions())
	}

	if m.statusMessage != "" {
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render(m.statusMessage))
	}

	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Press 'q' to quit • 'tab' to switch tabs • 'r' to refresh"))

	return b.String()
}

func (m model) viewMarket() string {
	var b strings.Builder

	b.WriteString("📈 Recent Klines (15m):\n\n")

	if len(m.klines) == 0 {
		b.WriteString("No data available\n")
		return b.String()
	}

	headers := []string{"Time", "Open", "High", "Low", "Close", "Volume"}

	headerLine := fmt.Sprintf("%-19s %-12s %-12s %-12s %-12s %-15s", headers[0], headers[1], headers[2], headers[3], headers[4], headers[5])
	b.WriteString(headerLine)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("-", 85))
	b.WriteString("\n")

	for _, kline := range m.klines {
		timeStr := kline.StartTime
		if len(timeStr) > 19 {
			timeStr = timeStr[:19]
		}
		line := fmt.Sprintf("%-19s %-12s %-12s %-12s %-12s %-15s",
			timeStr,
			kline.Open,
			kline.High,
			kline.Low,
			kline.Close,
			kline.Volume,
		)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) viewOrderBook() string {
	var b strings.Builder

	b.WriteString("📊 Order Book:\n\n")

	if m.orderbook == nil {
		b.WriteString("No data available\n")
		return b.String()
	}

	b.WriteString("Bids:\n")
	b.WriteString(fmt.Sprintf("%-15s %-15s\n", "Price", "Size"))
	b.WriteString(strings.Repeat("-", 32))
	b.WriteString("\n")

	for _, bid := range m.orderbook.Result.Bids {
		price, _ := strconv.ParseFloat(bid.Price, 64)
		size, _ := strconv.ParseFloat(bid.Quantity, 64)
		b.WriteString(fmt.Sprintf("%-15.2f %-15.4f\n", price, size))
	}

	b.WriteString("\nAsks:\n")
	b.WriteString(fmt.Sprintf("%-15s %-15s\n", "Price", "Size"))
	b.WriteString(strings.Repeat("-", 32))
	b.WriteString("\n")

	for _, ask := range m.orderbook.Result.Asks {
		price, _ := strconv.ParseFloat(ask.Price, 64)
		size, _ := strconv.ParseFloat(ask.Quantity, 64)
		b.WriteString(fmt.Sprintf("%-15.2f %-15.4f\n", price, size))
	}

	return b.String()
}

func (m model) viewTickers() string {
	var b strings.Builder

	b.WriteString("📊 Tickers:\n\n")

	if len(m.tickers) == 0 {
		b.WriteString("No data available\n")
		return b.String()
	}

	for _, ticker := range m.tickers {
		b.WriteString(fmt.Sprintf("Symbol: %s\n", ticker.Symbol))
		b.WriteString(fmt.Sprintf("  Last Price: %s\n", ticker.LastPrice))
		b.WriteString(fmt.Sprintf("  24h Change: %s\n", ticker.Price24HPcnt))
		b.WriteString(fmt.Sprintf("  24h High: %s\n", ticker.HighPrice24H))
		b.WriteString(fmt.Sprintf("  24h Low: %s\n", ticker.LowPrice24H))
		b.WriteString(fmt.Sprintf("  24h Volume: %s\n", ticker.Volume24H))
		b.WriteString(fmt.Sprintf("  Turnover: %s\n", ticker.Turnover24H))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) viewPositions() string {
	var b strings.Builder

	b.WriteString("📊 Positions:\n\n")

	if len(m.positions) == 0 {
		b.WriteString("No open positions\n")
		return b.String()
	}

	headers := []string{"Symbol", "Side", "Size", "Entry Price", "Mark Price", "Unrealized PnL"}
	b.WriteString(fmt.Sprintf("%-15s %-8s %-12s %-15s %-15s %-15s\n", headers[0], headers[1], headers[2], headers[3], headers[4], headers[5]))
	b.WriteString(strings.Repeat("-", 85))
	b.WriteString("\n")

	for _, pos := range m.positions {
		b.WriteString(fmt.Sprintf("%-15s %-8s %-12s %-15s %-15s %-15s\n",
			string(pos.Symbol),
			string(pos.Side),
			pos.Size,
			pos.AvgPrice,
			pos.MarkPrice,
			pos.UnrealisedPnl,
		))
	}

	return b.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
