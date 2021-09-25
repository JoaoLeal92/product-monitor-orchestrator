package entities

type ProcessingChannels struct {
	CrawlerJobsChan      chan ProductRelations
	CrawlerResultsChan   chan CrawlerChanRestult
	ProcessingResultChan chan []ProductSearchResult
}
