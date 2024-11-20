package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getWebURL(t *testing.T) {
	tests := []struct {
		city   string
		webURL string
	}{
		{"深圳", "https://weather.cma.cn/web/weather/59493.html"},
		{"广州", "https://weather.cma.cn/web/weather/59287.html"},
		{"北京", "https://weather.cma.cn/web/weather/54511.html"},
		{"上海", "https://weather.cma.cn/web/weather/58367.html"},
		{"龙国", ""},
	}

	for _, test := range tests {
		t.Run(test.city, func(t *testing.T) {
			webURL := getWebURL(test.city)
			require.Equal(t, test.webURL, webURL)
		})
	}
}
