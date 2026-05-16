package integration

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/beevik/etree"
)

type CBRClient struct{}

func NewCBRClient() *CBRClient {
	return &CBRClient{}
}

func (c *CBRClient) GetCentralBankRate() (float64, error) {
	soapRequest := buildSOAPRequest()
	rawBody, err := sendRequest(soapRequest)
	if err != nil {
		return 0, err
	}
	rate, err := parseXMLResponse(rawBody)
	if err != nil {
		return 0, err
	}
	rate += 5 // маржа банка
	return rate, nil
}

func buildSOAPRequest() string {
	fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	toDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
		<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
			<soap12:Body>
				<KeyRate xmlns="http://web.cbr.ru/">
					<fromDate>%s</fromDate>
					<ToDate>%s</ToDate>
				</KeyRate>
			</soap12:Body>
		</soap12:Envelope>`, fromDate, toDate)
}

func sendRequest(soapRequest string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(
		"POST",
		"https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx",
		bytes.NewBuffer([]byte(soapRequest)),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса: %v", err)
	}
	defer resp.Body.Close()
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %v", err)
	}
	return rawBody, nil
}

func parseXMLResponse(rawBody []byte) (float64, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(rawBody); err != nil {
		return 0, fmt.Errorf("ошибка парсинга XML: %v", err)
	}
	krElements := doc.FindElements("//diffgram/KeyRate/KR")
	if len(krElements) == 0 {
		return 0, errors.New("данные по ставке не найдены")
	}
	latestKR := krElements[0]
	rateElement := latestKR.FindElement("./Rate")
	if rateElement == nil {
		return 0, errors.New("тег Rate отсутствует")
	}
	rateStr := rateElement.Text()
	var rate float64
	if _, err := fmt.Sscanf(rateStr, "%f", &rate); err != nil {
		return 0, fmt.Errorf("ошибка конвертации ставки: %v", err)
	}
	return rate, nil
}
