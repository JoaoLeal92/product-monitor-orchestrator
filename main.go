package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dlclark/regexp2"
	"github.com/mitchellh/mapstructure"

	"github.com/JoaoLeal92/product-monitor-orchestrator/config"
	"github.com/JoaoLeal92/product-monitor-orchestrator/data"
)

type Product struct {
	Price         string `mapstructure:"price"`
	OriginalPrice string `mapstructure:"originalPrice"`
	Discount      string `mapstructure:"discount"`
	Link          string `mapstructure:"link"`
}

func main() {

	crawlerPathMap := map[string]string{
		"amazon": "../crawlers/amazon",
	}

	regexMap := map[string]string{
		"price":         `(?<=price=).*(?=,\s?original_price)`,
		"originalPrice": `(?<=original_price=).*(?=,\s?discount)`,
		"discount":      `(?<=discount=).*(?=,\s?link)`,
		"link":          `(?<=link=').*(?='\))`,
	}

	fmt.Println("Start orchestrator")
	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	db, err := data.Instance(cfg.Db)

	users, err := db.Users().GetUsers()
	if err != nil {
		panic(err)
	}

	for _, user := range users {
		products, err := db.Products().GetActiveProductsByUser(user.ID)
		if err != nil {
			panic(err)
		}

		for _, product := range products {
			crawlerName, _ := db.Crawlers().GetCrawlerName(product.CrawlerID)
			if err != nil {
				panic(err)
			}

			crawlerPath, err := filepath.Abs(crawlerPathMap[crawlerName])
			if err != nil {
				panic(err)
			}

			pipfile, err := filepath.Abs(
				filepath.Join(
					crawlerPathMap[crawlerName],
					"Pipfile",
				),
			)
			if err != nil {
				panic(err)
			}

			os.Setenv("PIPENV_PIPFILE", pipfile)

			cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", product.Link))

			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb

			cmd.Run()
			fmt.Println("output: ", outb.String())

			product, err := parseCrawlerOutput(regexMap, outb.String())
			if err != nil {
				panic(err)
			}

			fmt.Printf("%+v\n", product)
		}
	}

	fmt.Println("Fim")

	// crawlerPath, err := filepath.Abs("../crawlers/amazon")
	// fmt.Println(crawlerPath)
	// if err != nil {
	// 	panic(err)
	// }

	// pipfile, err := filepath.Abs("../crawlers/amazon/Pipfile")
	// if err != nil {
	// 	panic(err)
	// }

	// regexMap := map[string]string{
	// 	"price":         `(?<=price=).*(?=,\s?original_price)`,
	// 	"originalPrice": `(?<=original_price=).*(?=,\s?discount)`,
	// 	"discount":      `(?<=discount=).*(?=,\s?link)`,
	// 	"link":          `(?<=link=').*(?='\))`,
	// }

	// url := "https://www.amazon.com.br/dp/B07N6RMSY3/?coliid=I3IBUR6XCSG0M6&colid=2BQG6XN6AZAHT&psc=1&ref_=lv_ov_lig_dp_it"
	// url := "https://www.amazon.com.br/Adaptador-3x1-Tipo-c-Thunderbolt-Hdmi/dp/B077NGXPSV/?_encoding=UTF8&pd_rd_w=7friT&pf_rd_p=e824797f-48a0-405f-a37a-4e512542e9e8&pf_rd_r=Y80SFM425X6TYA9224FJ&pd_rd_r=2ec40c6d-e2f4-4c70-987b-3d7332c450d9&pd_rd_wg=ZDLF9&ref_=pd_gw_ci_mcx_mr_hp_atf_m"

	// os.Setenv("PIPENV_PIPFILE", pipfile)
	// cmd := exec.Command("pipenv", "run", "python", crawlerPath, fmt.Sprintf("-u %s", url))

	// var outb, errb bytes.Buffer
	// cmd.Stdout = &outb
	// cmd.Stderr = &errb

	// cmd.Run()
	// fmt.Println("output: ", outb.String())

	// product, err := parseCrawlerOutput(regexMap, outb.String())
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("%+v\n", product)
}

func parseCrawlerOutput(reMap map[string]string, out string) (*Product, error) {
	product := Product{}
	parsedData := make(map[string]interface{})

	for k, v := range reMap {
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

	err := mapstructure.Decode(parsedData, &product)
	if err != nil {
		return &Product{}, err
	}

	return &product, nil
}
