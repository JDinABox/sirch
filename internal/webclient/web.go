package webclient

import (
	_ "embed"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"golang.org/x/net/html/charset"
	"resty.dev/v3"
)

func Get(url string) (string, error) {
	client := resty.New()
	defer client.Close()

	res, err := client.R().SetHeader("Accept", "text/html").
		SetHeader("Accept-Language", "*").
		Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var c *Cleaner
	if strings.Contains(url, "wikipedia.org") {
		c = &WikipediaCleaner
	}

	if c == nil {
		b, err := htmltomarkdown.ConvertReader(res.Body)
		return CleanupMD(string(b)), err

	}

	ct := res.Header().Get("Content-Type")
	// bodyReader converts the content of resp.Body to UTF-8 if in need
	bodyReader, err := charset.NewReader(res.Body, ct)
	if err != nil {
		return "", err
	}
	p, err := c.Parse(bodyReader)
	if err != nil {
		return "", err
	}
	md, err := htmltomarkdown.ConvertString(p)
	return CleanupMD(md), err
}

func CleanupMD(md string) string {
	return strings.ReplaceAll(md, "\\[]\n", "")
}
