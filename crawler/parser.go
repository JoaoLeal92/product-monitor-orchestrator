package crawler

import (
	"fmt"

	"github.com/JoaoLeal92/product-monitor-orchestrator/entities"
	"github.com/dlclark/regexp2"
	"github.com/mitchellh/mapstructure"
)

type ResultParser struct {
	reMap map[string]string
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

func (r *ResultParser) parseCrawlerResult(crawlerOutput string) *entities.CrawlerResult {
	fmt.Println("Extraindo dados do retorno do crawler")
	crawlerResult, err := r.parseCrawlerOutput(crawlerOutput)
	if err != nil {
		panic(err)
	}

	return crawlerResult
}

func (r *ResultParser) parseCrawlerOutput(out string) (*entities.CrawlerResult, error) {
	crawlerResult := entities.CrawlerResult{}
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
			parsedData[k] = m.String()
		}
	}

	err := mapstructure.Decode(parsedData, &crawlerResult)
	if err != nil {
		return &entities.CrawlerResult{}, err
	}

	return &crawlerResult, nil
}
