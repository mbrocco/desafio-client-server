package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if error != nil {
		log.Println("001 - main - Erro: " + error.Error())
		panic(error)
	}

	res, error := http.DefaultClient.Do(req)
	if error != nil {
		log.Println("002 - main - Erro: " + error.Error())
		panic(error)
	}

	defer res.Body.Close()

	cotacao, error := ioutil.ReadAll(res.Body)
	if error != nil {
		log.Println("003 - main - Erro ao realizar a leitura de retorno da requisição: " + error.Error())
		panic(error)
	}

	file, error := os.Create("cotacao.txt")
	if error != nil {
		log.Println("004 - main - Erro ao criar o arquivo: " + error.Error())
		panic(error)
	}

	defer file.Close()
	_, error = file.WriteString(fmt.Sprintf("Dólar: " + string(cotacao)))
	if error != nil {
		log.Println("005 - main - Erro ao escrever no arquivo as informações sobre a cotação: " + error.Error())
		panic(error)
	}

	io.Copy(os.Stdout, res.Body)
	log.Println("Dados da cotação da moeda salva com sucesso no arquivo cotacao.txt")

}
