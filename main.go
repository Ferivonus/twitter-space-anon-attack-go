package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	driverPath = `C:\Users\fahre\Desktop\yazilim\chromedriver_win32\chromedriver.exe`
	basePort   = 9515 // portlar buradan başlayıp +i olarak gidecek
	threads    = 40   // kaç paralel oturum
)

// worker, verilen Space URL'i kullanarak oturum açar ve dinleme başlatır
func worker(id, port int, spaceURL string, wg *sync.WaitGroup) {
	defer wg.Done()

	// ChromeDriver servisini başlat
	service, err := selenium.NewChromeDriverService(driverPath, port)
	if err != nil {
		log.Printf("[Thread %02d] Servis başlatılamadı (port %d): %v\n", id, port, err)
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
		log.Printf("[Thread %02d] WebDriver’a bağlanılamadı: %v\n", id, err)
		return
	}
	defer wd.Quit()

	// Space peek sayfasına git
	if err := wd.Get(spaceURL); err != nil {
		log.Printf("[Thread %02d] Sayfa yüklenemedi: %v\n", id, err)
		return
	}

	// JS’nin yüklenmesi için bekle
	time.Sleep(5 * time.Second)

	// "Anonim olarak dinlemeye başla" butonunu bul ve tıkla
	btn, err := wd.FindElement(
		selenium.ByXPATH,
		`//span[contains(text(), 'Anonim olarak dinlemeye başla')]`,
	)
	if err != nil {
		log.Printf("[Thread %02d] Buton bulunamadı: %v\n", id, err)
		return
	}
	if err := btn.Click(); err != nil {
		log.Printf("[Thread %02d] Butona tıklama hatası: %v\n", id, err)
		return
	}
	log.Printf("[Thread %02d] Butona tıklandı, dinleme başladı.\n", id)

	// İstediğin süre kadar bekle
	time.Sleep(30 * time.Second)
	log.Printf("[Thread %02d] Bekleme tamamlandı, çıkılıyor.\n", id)
}

func main() {
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
		go worker(i+1, port, spaceURL, &wg)
	}

	wg.Wait()
	fmt.Println("Tüm thread’ler tamamlandı.")
}
