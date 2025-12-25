package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
)

func main() {

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(
		context.Background(),
		append(
			chromedp.DefaultExecAllocatorOptions[:],
			chromedp.ProxyServer("socks5://127.0.0.1:9050"),
			chromedp.Headless,
		)...,
	)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	logfile, err := os.Create("log.json")
	if err != nil {
		fmt.Println("Log dosyası açılamadı:", err)
		return
	}
	log.SetOutput(logfile)
	defer logfile.Close()

	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	tbTransport := &http.Transport{Dial: dialer.Dial}
	client := &http.Client{Transport: tbTransport,
		Timeout: 10 * time.Second,
	}

	file, err := os.Open("targets.yaml")
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		fmt.Println(url)

		var buf []byte
		var htmlContent string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.FullScreenshot(&buf, 90),
			chromedp.OuterHTML("html", &htmlContent),
		)
		if err != nil {
			log.Println("SS HATA:", url, err)
			continue
		}

		resp, err := client.Get(url)
		if err != nil {
			fmt.Println("URL hata:", url, err)
			continue
		}
		if resp.StatusCode != 200 {
			log.Println("PASİF URL:", url, "Code:", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		name := strings.ReplaceAll(url, "https://", "")
		name = strings.ReplaceAll(name, "http://", "")
		name = strings.ReplaceAll(name, "/", "_")
		os.WriteFile(name+".png", buf, 0644)
		os.WriteFile(name+".html", []byte(htmlContent), 0644)
		log.Println("AKTİF URL:", url, "Status:", resp.StatusCode)
		resp.Body.Close()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}

}
