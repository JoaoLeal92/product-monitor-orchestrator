package crawlerparser

import (
	"fmt"
	"strconv"

	"github.com/dlclark/regexp2"
	"github.com/mitchellh/mapstructure"
)

type ResultParser struct {
	reMap map[string]string
}

type crawlerResult struct {
	Price         int    `mapstructure:"price"`
	OriginalPrice int    `mapstructure:"originalPrice"`
	Discount      string `mapstructure:"discount"`
	Link          string `mapstructure:"link"`
}

func NewResultParser() *ResultParser {
	reMap := map[string]string{
		"price":         `(?<=price=).*(?=,\s?original_price)`,
		"originalPrice": `(?<=original_price=).*(?=,\s?discount)`,
		"discount":      `(?<=discount=).*(?=,\s?link)`,
		"link":          `(?<=link=').*(?='\))`,
	}

	return &ResultParser{
		reMap: reMap,
	}
}

func (r *ResultParser) ParseCrawlerResult(crawlerOutput string) *crawlerResult {
	fmt.Println("Extraindo dados do retorno do crawler")
	crawlerResult, err := r.parseCrawlerOutput(crawlerOutput)
	if err != nil {
		panic(err)
	}

	return crawlerResult
}

func (r *ResultParser) parseCrawlerOutput(out string) (*crawlerResult, error) {
	crawlerResult := crawlerResult{}
	parsedData := make(map[string]interface{})

	for k, v := range r.reMap {
		re := regexp2.MustCompile(v, 0)
		m, err := re.FindStringMatch(out)
		if err != nil {
			fmt.Println(err.Error())
		}

		if m.String() == "None" {
			parsedData[k] = nil
		} else {
			if k == "price" || k == "originalPrice" {
				intPrice, err := r.priceStringToInt(m.String())
				if err != nil {
					fmt.Println(err)
				}
				parsedData[k] = intPrice
			} else {
				parsedData[k] = m.String()
			}
		}
	}

	err := mapstructure.Decode(parsedData, &crawlerResult)
	if err != nil {
		return &crawlerResult, err
	}

	return &crawlerResult, nil
}

func (r *ResultParser) priceStringToInt(priceString string) (int, error) {
	var priceInt int
	var err error

	if priceString != "" {
		priceInt, err = strconv.Atoi(priceString)
		if err != nil {
			return 0, err
		}
	}

	return priceInt, nil
}
