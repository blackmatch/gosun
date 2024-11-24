package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"net/http"
	"strings"

	"github.com/spf13/cobra"

	_ "embed"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
)

var listFlag bool
var rootCmd = &cobra.Command{
	Use:   "gosun 城市名",
	Short: "gosun 是一款用于查询中国城市天气的终端工具",
	Long:  `gosun 是一款用于查询中国城市天气的终端工具`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("请输入城市名")
			return
		}

		webURL := getWebURL(args[0])
		if webURL == "" {
			fmt.Println("查询的城市名称有误！")
			return
		}

		if listFlag {
			cts := getCityList(args[0])
			if len(cts) == 0 {
				fmt.Println("请输入正确的省份名/直辖市名")
				return
			}
			fmt.Printf("以下是%s的所有城市/地区名：\n", args[0])
			for _, c := range cts {
				fmt.Println(c)
			}
			return
		}

		wd, err := getWeatherData(webURL)
		if err != nil {
			fmt.Println(err)
			return
		}
		showWeatherData(wd)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

//go:embed data.json
var data []byte
var provinces []province

func init() {
	_ = json.Unmarshal(data, &provinces)

	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "List all cities")
}

func getWebURL(city string) string {
	for _, prvs := range provinces {
		for _, c := range prvs.Cities {
			if strings.Contains(c.Name, city) {
				return c.WebURL
			}
		}

		if strings.Contains(prvs.Name, city) {
			return prvs.Cities[0].WebURL
		}
	}

	return ""
}

type dayWeather struct {
	weather string
	maxTemp string
	wind    string
	windDir string
}

type nightWeather struct {
	weather string
	minTemp string
	wind    string
	windDir string
}

type weatherInfo struct {
	date     string
	week     string
	dayWtr   dayWeather
	nightWtr nightWeather
}

type weatherData struct {
	city       string
	updateTime string
	weathers   []weatherInfo
}

func getWeatherData(webURL string) (*weatherData, error) {
	resp, err := http.Get(webURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	wd := &weatherData{}
	doc.Find(".hp .hd").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "7天天气预报") {
			wd.updateTime = normalizeText(s.Text())
		}
	})
	wd.city = normalizeText(doc.Find("#breadcrumb li.active").Text())
	wd.weathers = make([]weatherInfo, 0)
	doc.Find("#dayList .pull-left.day").Each(func(i int, s *goquery.Selection) {
		info := weatherInfo{}
		var dateWeek string
		s.Find(".day-item").Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				// no need to normalize
				dateWeek = s.Text()
			case 2:
				info.dayWtr.weather = normalizeText(s.Text())
			case 3:
				info.dayWtr.wind = normalizeText(s.Text())
			case 4:
				info.dayWtr.windDir = normalizeText(s.Text())
			case 7:
				info.nightWtr.weather = normalizeText(s.Text())
			case 8:
				info.nightWtr.wind = normalizeText(s.Text())
			case 9:
				info.nightWtr.windDir = normalizeText(s.Text())
			}
		})
		maxTemp := s.Find(".day-item.bardiv .high").Text()
		info.dayWtr.maxTemp = normalizeText(maxTemp)
		minTemp := s.Find(".day-item.bardiv .low").Text()
		info.nightWtr.minTemp = normalizeText(minTemp)
		dateWeek = strings.Trim(dateWeek, "\n")
		dateWeek = strings.Trim(dateWeek, " ")
		dateWeek = strings.Trim(dateWeek, "\n")
		arr := strings.Split(dateWeek, "\n")
		info.week = arr[0]
		info.date = strings.Trim(arr[1], " ")
		wd.weathers = append(wd.weathers, info)
	})

	return wd, nil
}

func normalizeText(s string) string {
	return strings.Trim(strings.ReplaceAll(s, "\n", ""), " ")
}

func showWeatherData(wd *weatherData) {
	fmt.Println(wd.city, " ", wd.updateTime)
	for idx, info := range wd.weathers {
		if idx == 0 {
			color.RGB(56, 114, 185).Printf("%s（%s）%s 最高气温%s %s %s -> %s 最低气温%s %s %s\n", info.week,
				info.date, info.dayWtr.weather, info.dayWtr.maxTemp, info.dayWtr.wind, info.dayWtr.windDir,
				info.nightWtr.weather, info.nightWtr.minTemp, info.nightWtr.wind, info.nightWtr.windDir)
		} else {
			fmt.Printf("%s（%s）%s 最高气温%s %s %s -> %s 最低气温%s %s %s\n", info.week,
				info.date, info.dayWtr.weather, info.dayWtr.maxTemp, info.dayWtr.wind, info.dayWtr.windDir,
				info.nightWtr.weather, info.nightWtr.minTemp, info.nightWtr.wind, info.nightWtr.windDir)
		}
	}
}

func getCityList(province string) []string {
	cities := make([]string, 0)
	for _, prv := range provinces {
		if strings.Contains(prv.Name, province) {
			for _, ct := range prv.Cities {
				cities = append(cities, ct.Name)
			}
			break
		}
	}

	return cities
}
