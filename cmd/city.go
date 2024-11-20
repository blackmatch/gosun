package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type city struct {
	Name   string `json:"name"`
	WebURL string `json:"web_url"`
}

type province struct {
	Name   string `json:"name"`
	WebURL string `json:"web_url"`
	Cities []city `json:"cities"`
}

func getWebDocument(webURL string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", webURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return goquery.NewDocumentFromReader(resp.Body)
}

func getData() error {
	doc, err := getWebDocument("https://weather.cma.cn/web/text/HB/ABJ.html")
	if err != nil {
		return err
	}

	prvs := make([]*province, 0)
	doc.Find(".province-list .province-item").Each(func(i int, s *goquery.Selection) {
		p := &province{}
		p.Name = s.Text()
		if path, ok := s.Attr("href"); ok {
			p.WebURL = fmt.Sprintf("https://weather.cma.cn%s", path)
		}
		prvs = append(prvs, p)
	})

	wg := sync.WaitGroup{}
	c := make(chan *province, 3)
	go func() {
		defer close(c)
		for _, prv := range prvs {
			c <- prv
			wg.Add(1)
		}
	}()

	for {
		p, ok := <-c
		if !ok {
			break
		}
		if p.WebURL != "" {
			go func() {
				defer wg.Done()
				cts, err := getCities(p.WebURL)
				if err != nil {
					fmt.Printf("get cities for provics %s failed: %v", p.Name, err)
					return
				}
				p.Cities = cts
			}()
		}
	}

	wg.Wait()

	// save to file
	jsonData, err := json.MarshalIndent(prvs, "", "  ")
	if err != nil {
		fmt.Println("marshal :", err)
		return err
	}
	err = os.WriteFile("data.json", jsonData, 0644)
	if err != nil {
		fmt.Println("write to file err:", err)
	}
	return err
}

func getCities(webURL string) ([]city, error) {
	doc, err := getWebDocument(webURL)
	if err != nil {
		return nil, err
	}

	cts := make([]city, 0)
	doc.Find(".tab-pane.active .day-table tbody tr td a").Each(func(i int, s *goquery.Selection) {
		if s.Text() != "" && s.Text() != "详情>>" {
			c := city{}
			c.Name = s.Text()
			if path, ok := s.Attr("href"); ok {
				c.WebURL = fmt.Sprintf("https://weather.cma.cn%s.html", path)
			}

			cts = append(cts, c)
		}
	})

	return cts, nil
}
