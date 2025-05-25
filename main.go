package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	basePort = 9515 // portlar buradan başlayıp +i olarak gidecek
	threads  = 40   // kaç paralel oturum
)

// worker, verilen Space URL'i kullanarak oturum açar ve dinleme başlatır
func worker(id, port int, driverPath, spaceURL string, wg *sync.WaitGroup) {
	defer wg.Done()

	// ChromeDriver servisini başlat
	service, err := selenium.NewChromeDriverService(driverPath, port)
	if err != nil {
		log.Printf("[Thread %02d] Servis başlatılamadı (port %d): %v", id, port, err)
		return
	}
	defer service.Stop()

	// WebDriver kabiliyetleri
	caps := selenium.Capabilities{"browserName": "chrome"}
	chromeCaps := chrome.Capabilities{
		Args: []string{"--headless", "--disable-gpu", "--window-size=1280,800"},
	}
	caps.AddChrome(chromeCaps)

	// Remote WebDriver’a bağlan
	url := fmt.Sprintf("http://localhost:%d/wd/hub", port)
	wd, err := selenium.NewRemote(caps, url)
	if err != nil {
		log.Printf("[Thread %02d] WebDriver’a bağlanılamadı: %v", id, err)
		return
	}
	defer wd.Quit()

	// Space peek sayfasına git
	if err := wd.Get(spaceURL); err != nil {
		log.Printf("[Thread %02d] Sayfa yüklenemedi: %v", id, err)
		return
	}

	// JS’nin yüklenmesi için bekle
	time.Sleep(5 * time.Second)

	// "Anonim olarak dinlemeye başla" butonunu bul ve tıkla
	btn, err := wd.FindElement(
		selenium.ByXPATH,
		"//span[contains(text(), 'Anonim olarak dinlemeye başla')]",
	)
	if err != nil {
		log.Printf("[Thread %02d] Buton bulunamadı: %v", id, err)
		return
	}
	if err := btn.Click(); err != nil {
		log.Printf("[Thread %02d] Butona tıklama hatası: %v", id, err)
		return
	}
	log.Printf("[Thread %02d] Butona tıklandı, dinleme başladı.", id)

	// İstediğin süre kadar bekle
	time.Sleep(30 * time.Second)
	log.Printf("[Thread %02d] Bekleme tamamlandı, çıkılıyor.", id)
}

func main() {
	// chromedriver.exe, çalıştırılabilir dosyanın bulunduğu dizinde olmalı
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("Executable path alınamadı: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	driverPath := filepath.Join(exeDir, "chromedriver.exe")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Lütfen Twitter Spaces peek linkini girin: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("Girdi okunamadı: %v", err)
	}
	spaceURL := strings.TrimSpace(input)

	var wg sync.WaitGroup
	wg.Add(threads)

	for i := 0; i < threads; i++ {
		port := basePort + i
		go worker(i+1, port, driverPath, spaceURL, &wg)
	}

	wg.Wait()
	fmt.Println("Tüm thread’ler tamamlandı.")
}
