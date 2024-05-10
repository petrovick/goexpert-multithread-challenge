package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type AddressViaCep struct {
	Zip      string `json:"cep"`
	State    string `json:"uf"`
	City     string `json:"localidade"`
	Neighbor string `json:"bairro"`
	Street   string `json:"logradouro"`
}

type AddressBrasilAPI struct {
	Zip      string `json:"cep"`
	State    string `json:"state"`
	City     string `json:"city"`
	Neighbor string `json:"neighborhood"`
	Street   string `json:"street"`
}

func main() {
	cep := "38705458"
	ch1 := make(chan *AddressViaCep)
	ch2 := make(chan *AddressBrasilAPI)

	go func() {
		addressViaCep := getAPIContent("http://viacep.com.br/ws/"+cep+"/json", new(AddressViaCep))
		ch1 <- addressViaCep
	}()

	go func() {
		addressBrasilAPI := getAPIContent("https://brasilapi.com.br/api/cep/v1/"+cep, new(AddressBrasilAPI))
		ch2 <- addressBrasilAPI
	}()

	select {
	case msg := <-ch1: // ViaCEP
		fmt.Printf("Received from Via CEP: CEP: %s, City: %s, Street: %s\n", msg.Zip, msg.City, msg.Street)

	case msg := <-ch2: // BrasilAPI
		fmt.Printf("Received from Brasil API: %s, City: %s, Street: %s\n", msg.Zip, msg.City, msg.Street)

	case <-time.After(time.Second):
		println("timeout")
	}
}

func getAPIContent[T any](url string, model T) T {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	body := getRequestBody(req, err)

	return convertJsonContentToModel(body, model)
}

func convertJsonContentToModel[T any](body []byte, model T) T {
	err := json.Unmarshal(body, &model)

	if err != nil {
		log.Println("Error parsing JSON data")
		panic(err)
	}

	return model
}

func getRequestBody(req *http.Request, err error) []byte {
	if err != nil {
		log.Println("Error creating request")
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Error retrieving data")
		panic(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading API data")
		panic(err)
	}

	return body
}
