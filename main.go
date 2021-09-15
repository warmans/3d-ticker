package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/tarm/serial"
	"github.com/warmans/3d-ticker/icon"
	"math"
	"net/http"
	"os"
	"sync"
	"time"
)

var logFile *os.File

func init() {
	var err error
	logFile, err = os.Create("3d-ticker.log")
	if err != nil {
		panic(err.Error())
		return
	}
}

type Config struct {
	Tickers              []string
	FetchIntervalSeconds int
	ComPort              string
	BaudRate             int
}

type TickerClick struct {
	button *systray.MenuItem
	ticker string
}

func main() {

	var configPath string
	flag.StringVar(&configPath, "-config-path", "./conf.json", "specify config path")
	flag.Parse()

	cfg, err := loadConfig(configPath)
	if err != nil {
		panic(err.Error())
	}
	systray.Run(
		func() {
			onReady(cfg)
		},
		func() {

		},
	)
}

func writeToLog(v ...interface{}) {
	fmt.Fprintln(logFile, v...)
}

func loadConfig(path string) (Config, error) {

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return initConfig(path)
		}
		return Config{}, err
	}
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	return cfg, json.NewDecoder(f).Decode(&cfg)
}

func initConfig(path string) (Config, error) {
	cfg := Config{
		Tickers: []string{
			"ethereum",
			"polkadot",
			"cosmos",
		},
		FetchIntervalSeconds: 60 * 10, // 10 mins
		ComPort:              "COM5",
		BaudRate:             9600,
	}
	f, err := os.Create(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	return cfg, json.NewEncoder(f).Encode(cfg)
}

func onReady(cfg Config) {

	s, err := serial.OpenPort(&serial.Config{Name: cfg.ComPort, Baud: cfg.BaudRate})
	if err != nil {
		writeToLog(err.Error())
	}
	defer s.Close()

	systray.SetIcon(icon.Data)
	systray.SetTitle("3d Ticker")

	closeChannel := make(chan struct{})
	clickChannel := make(chan TickerClick)
	wg := sync.WaitGroup{}

	tickerbuttons := make([]*systray.MenuItem, len(cfg.Tickers))
	for k, v := range cfg.Tickers {
		tickerbuttons[k] = systray.AddMenuItemCheckbox(v, "v", k == 0)
		wg.Add(1)
		go func(ticker string, menuItem *systray.MenuItem) {
			for {
				select {
				case <-menuItem.ClickedCh:
					clickChannel <- TickerClick{ticker: ticker, button: menuItem}
				case <-closeChannel:
					wg.Done()
					return
				}
			}
		}(v, tickerbuttons[k])
	}

	quitButton := systray.AddMenuItem("QUIT", "Kill the host app")

	defer func() {
		close(closeChannel)
		wg.Wait()
		close(closeChannel)
		systray.Quit()
	}()

	if cfg.FetchIntervalSeconds < 10 {
		cfg.FetchIntervalSeconds = 60
	}
	ticker := time.NewTicker(time.Second * time.Duration(cfg.FetchIntervalSeconds))

	var selected = cfg.Tickers[0]
	for {
		select {
		case <-quitButton.ClickedCh:
			return
		case clicked := <-clickChannel:
			for _, v := range tickerbuttons {
				v.Uncheck()
			}
			clicked.button.Check()

			if err := dispatchSerialData(s, clicked.ticker); err != nil {
				writeToLog(err.Error())
			}
			selected = clicked.ticker
		case <-ticker.C:
			if err := dispatchSerialData(s, selected); err != nil {
				writeToLog(err.Error())
			}
		}
	}
}

func dispatchSerialData(s *serial.Port, id string) error {

	res, err := getTickerData(id)
	if err != nil {
		return err
	}
	bytes := make([]byte, 5)
	for i := 0; i < 5; i++ {
		bytes[i] = fmt.Sprintf("%d", res[i])[0]
	}
	_, err = s.Write(append(bytes, []byte("\n")...))
	return err
}

func getTickerData(id string) ([]uint8, error) {

	res, err := http.DefaultClient.Get(fmt.Sprintf("https://api.coingecko.com/api/v3/coins/%s/market_chart?vs_currency=eur&days=1&interval=hourly", id))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed: %s", res.Status)
	}
	data := &PriceData{}
	if err := json.NewDecoder(res.Body).Decode(data); err != nil {
		return nil, err
	}
	return FormatDataForDisplay(data.Prices), nil
}

// FormatDataForDisplay takes 24 time -> value pairs and averages them into 5 buckets
func FormatDataForDisplay(prices [][]float64) []uint8 {
	if prices == nil {
		return make([]uint8, 5)
	}

	// take the arbitrary sized slice and group it to the number of bars.
	series := GroupSeries(prices)

	// scale the averages to a range of 0-9
	scaledResult := make([]uint8, 5, 5)
	for k, v := range series {
		scaledResult[k] = ScalePrice(series, v)
	}

	return scaledResult
}

func GroupSeries(prices [][]float64) []float64 {
	result := make([]*aggregatedPrice, 5, 5)
	for k, v := range prices {
		bucket := int((float64(k) / float64(len(prices))) * 5)
		if result[bucket] == nil {
			result[bucket] = &aggregatedPrice{}
		}
		result[bucket].value += v[1]
		result[bucket].count++
	}

	// average the prices
	series := make([]float64, 5, 5)
	for k, r := range result {
		if r == nil {
			series[k] = 0
		} else {
			series[k] = r.value / float64(r.count)
		}
	}

	return series
}

// ScalePrice makes a price fit into the 10 LED bar
func ScalePrice(prices []float64, price float64) uint8 {
	maxPrice := float64(0)
	minPrice := math.MaxFloat32
	for _, v := range prices {
		if v > maxPrice {
			maxPrice = v
		}
		if v < minPrice {
			minPrice = v
		}
	}
	// LEDs are indexed 0 - 9
	return uint8(((price - minPrice) / (maxPrice - minPrice)) * 9)
}

type PriceData struct {
	Prices [][]float64 `json:"prices"`
}

type aggregatedPrice struct {
	value float64
	count int
}
