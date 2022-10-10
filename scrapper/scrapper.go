package scrapper

import (
	"bytes"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"webscrapper/models"
)

var count = 0
var ableToScroll = true

//https://www.mvideo.ru/noutbuki-planshety-komputery-8/noutbuki-118?reff=menu_main
type Scrapper struct {
	collyInstance *colly.Collector
	goods         map[string]models.Notebook
	logger        *logrus.Logger
}

func removeNotDigits(s string) string {
	return strings.Map(
		func(r rune) rune {
			if unicode.IsDigit(r) {
				return r
			}
			return -1
		},
		s,
	)
}
func (wc Scrapper) SetTriggers() {
	wc.collyInstance.OnHTML(".Specifications__row", func(element *colly.HTMLElement) {

		if element.Response.StatusCode != http.StatusOK {
			wc.logger.Errorf("status code %v", element.Response.StatusCode)
			return
		}
		url := element.Request.URL.String()
		if reg, _ := regexp.Compile("https:\\/\\/www\\.citilink\\.ru\\/product\\/\\S*"); !reg.MatchString(url) {
			return
		}

		regStorageFlag, err := regexp.Compile("Объем")
		regResolutionFlag, err := regexp.Compile("Разрешение экрана")
		regCPUFreqFlag, err := regexp.Compile("Процессор, частота")
		regCPUCoresFlag, err := regexp.Compile("Количество ядер процессора")
		regRAMFlag, err := regexp.Compile("Оперативная память")
		regGPURAMFlag, err := regexp.Compile("Графический процессор")
		if err != nil {
			wc.logger.Error(err)
			return
		}

		if regResolutionFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			resolution := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := resolution //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regResolution, err := regexp.Compile("\\d{3,5}х\\d{3,5}")
			if err != nil {
				wc.logger.Error(err)
				return
			}

			if regResolution.MatchString(text) {
				//element.ForEach(".Specifications__column", func(_ int, elem *colly.HTMLElement) {
				elemText := string(bytes.Trim([]byte(element.Text), "\n\t "))
				//	if regResolution.MatchString(elemText) {
				if good, ok := wc.goods[url]; ok {
					good.ScreenResolution = regResolution.FindStringSubmatch(elemText)[0]
					wc.goods[url] = good
				}
				//	}
				//})
			}
		}
		if regCPUFreqFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			freq := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := freq //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regCPUFreq, err := regexp.Compile("\\d.\\d+ ГГц")
			regCPUFreqVal, err := regexp.Compile("^\\d.\\d+")
			if err != nil {
				wc.logger.Error(err)
				return
			}

			if regCPUFreq.MatchString(text) {
				elemText := regCPUFreqVal.FindStringSubmatch(text)[0]
				freqValue, err := strconv.ParseFloat(elemText, 32)
				if err != nil {
					wc.logger.Error(err)
					return
				}
				if good, ok := wc.goods[url]; ok {

					good.CPUFrequency = freqValue * 1000
					wc.goods[url] = good
				}
			}
		}
		if regCPUCoresFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			cores := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := cores //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regCPUCores, err := regexp.Compile("\\d-ядерный")
			regCPUCoresVal, err := regexp.Compile("\\d")
			if err != nil {
				wc.logger.Error(err)
				return
			}
			if regCPUCores.MatchString(text) {
				elemText := regCPUCoresVal.FindStringSubmatch(text)[0]
				coresValue, err := strconv.ParseInt(elemText, 10, 32)
				if err != nil {
					wc.logger.Error(err)
					return
				}
				if good, ok := wc.goods[url]; ok {

					good.CPUCores = int(coresValue)
					wc.goods[url] = good
				}
			}
		}

		if regRAMFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			ram := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := ram //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regRAM, err := regexp.Compile("\\d{1,3} ГБ")
			regRAMVal, err := regexp.Compile("\\d{1,3}")
			if err != nil {
				wc.logger.Error(err)
				return
			}
			if regRAM.MatchString(text) {
				elemText := regRAMVal.FindStringSubmatch(text)[0]
				RAMValue, err := strconv.ParseInt(elemText, 10, 32)
				if err != nil {
					fmt.Println(err)
					return
				}
				if good, ok := wc.goods[url]; ok {

					good.RAM = int(RAMValue)
					wc.goods[url] = good
				}
			}
		}
		if regGPURAMFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			gpuram := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := gpuram //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regGPURAM, err := regexp.Compile("(\\w|\\s)*- \\d{1,5} Мб")
			regGPURAMVal, err := regexp.Compile("- \\d{1,5} Мб")
			if err != nil {
				wc.logger.Error(err)
				return
			}
			if regGPURAM.MatchString(text) {
				elemText := regGPURAMVal.FindStringSubmatch(text)[0]
				elemText = removeNotDigits(elemText)
				GPURAMValue, err := strconv.ParseInt(elemText, 10, 32)
				if err != nil {
					fmt.Println(err)
					return
				}
				if good, ok := wc.goods[url]; ok {

					good.GPURAM = int(GPURAMValue)
					wc.goods[url] = good
				}
			}
		}

		if regStorageFlag.MatchString(element.ChildText(".Specifications__column_name")) {
			storage := element.ChildText(".Specifications__column_value") //element.DOM.Nodes[0].NextSibling.LastChild.Data

			text := storage //string(bytes.Trim([]byte(element.Text), "\n\t "))
			regStorage, err := regexp.Compile("\\s*(\\d{1,5}) (ГБ|ТБ)\\W*")
			if err != nil {
				wc.logger.Error(err)
				return
			}
			if regStorage.MatchString(text) {
				storageString := regStorage.FindStringSubmatch(text)
				storageVal := storageString[1]
				storageUnit := storageString[2]
				storageValNum, err := strconv.ParseInt(storageVal, 10, 32)
				if err != nil {
					fmt.Println(err)
					return
				}
				if good, ok := wc.goods[url]; ok {
					if storageUnit == "ГБ" {
						good.Storage += int(storageValNum) * 1000
					} else if storageUnit == "ТБ" {
						good.Storage += int(storageValNum) * 1000000
					}
					wc.goods[url] = good
				}
			}
		}

	})
	wc.collyInstance.OnHTML(".ProductHeader__price-default", func(element *colly.HTMLElement) {
		if element.Response.StatusCode != http.StatusOK {
			wc.logger.Errorf("status code %v", element.Response.StatusCode)
			return
		}
		url := element.Request.URL.String()
		if reg, _ := regexp.Compile("https:\\/\\/www\\.citilink\\.ru\\/product\\/\\S*"); !reg.MatchString(url) {
			return
		}
		price, err := strconv.Atoi(removeNotDigits(element.Text))
		if good, ok := wc.goods[url]; ok && price != 0 {

			if err != nil {
				fmt.Println(err)
			}
			good.Price = price
			wc.goods[url] = good
		}

	})
	wc.collyInstance.OnHTML(".ProductCardCategoryList__grid , .ProductCardVertical__name", func(element *colly.HTMLElement) {
		if reg, _ := regexp.Compile("https:\\/\\/www\\.citilink\\.ru\\/catalog\\/noutbuki\\/\\S*"); !reg.MatchString(element.Request.URL.String()) {
			return
		}
		//debug limiter off in production
		//if count >= 5 {
		//		//	ableToScroll = false
		//		//	return
		//		//}
		if element.Response.StatusCode != http.StatusOK {

			wc.logger.Errorf("status code %v", element.Response.StatusCode)
			return
		}
		if element.Attr("href") == "" {
			return
		}

		fmt.Println("Found")
		fmt.Println(element.Text)
		newRef := element.Attr("href") + "properties/"
		time.Sleep(time.Duration((rand.Int()%18)+8) * time.Second)
		wc.goods["https://www.citilink.ru"+newRef] = models.Notebook{Ref: "https://www.citilink.ru" + newRef, Name: element.Text}
		err := element.Request.Visit(newRef)
		if err != nil {
			delete(wc.goods, "https://www.citilink.ru"+newRef)
			wc.logger.Error(err)
			return
		}

	})

	wc.collyInstance.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting")
		//r.Headers.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64; rv:103.0) Gecko/20100101 Firefox/103.0")
		//r.Headers.Set("user-agent", "Yandex")
		r.Headers.Set("Referer", "https://google.com/")
		fmt.Println(r.URL)

	})
	wc.collyInstance.OnResponse(func(response *colly.Response) {
		if response.StatusCode != http.StatusOK {
			ableToScroll = false
		}
		if reg, _ := regexp.Compile("https:\\/\\/www\\.citilink\\.ru\\/catalog\\/noutbuki\\/\\S*"); !reg.MatchString(response.Request.URL.String()) {
			ableToScroll = false
		}
	})

}
func NewScrapper() *Scrapper {
	c := colly.NewCollector()
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     true,
	})
	scrapper := Scrapper{
		collyInstance: c,
		goods:         make(map[string]models.Notebook),
		logger:        logger,
	}
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	c.DisableCookies()

	return &scrapper

}
func (wc Scrapper) StartCrawling(lowPrice int, highPrice int) (map[string]models.Notebook, error) {
	var i int = 1
	var err error
	for ableToScroll {
		if i == 1 {

			err = wc.collyInstance.Visit(fmt.Sprintf("https://www.citilink.ru/catalog/noutbuki/?price_min=%d&price_max=%d", lowPrice, highPrice))
		} else {
			err = wc.collyInstance.Visit(fmt.Sprintf("https://www.citilink.ru/catalog/noutbuki/?price_min=%d&price_max=%d&p=%d", lowPrice, highPrice, i))

		}
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		i += 1
	}
	//err := wc.collyInstance.Visit("https://www.citilink.ru/product/noutbuk-asus-k3400pa-kp110w-i5-11300h-16gb-ssd512gb-14-wqxga-w11-blue-1583435/properties/")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	for _, val := range wc.goods {
		fmt.Println(val)
	}
	return wc.goods, nil

}

//https://www.citilink.ru/product/noutbuk-digma-eve-15-p4xx-15-6-ips-intel-pentium-j3710-4gb-128gb-ssd-i-1470838/properties/
