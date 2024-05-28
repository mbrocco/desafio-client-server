package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

func main() {

	http.HandleFunc("/cotacao", BuscaCotacaoHandler)
	log.Println("Servidor iniciado na porta: 8080")
	http.ListenAndServe(":8080", nil)
}

func buscarBancoDadosCon() (*sql.DB, error) {
	os.Remove("./base_dados.db")

	db, err := sql.Open("sqlite3", "./base_dados.db")
	if err != nil {
		log.Println("buscarBancoDadosCon - Erro ao abrir banco de dados - Erro: " + err.Error())
		return nil, err
	}
	//defer db.Close()

	sqlStmt := `create table cotacao (code text not null, codein text, name text, high text, low text, varBid text, pctChange text, bid text, ask text, timestamp text not null, create_date text, PRIMARY KEY (code, timestamp));`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Println("buscarBancoDadosCon - Erro ao criar a tabela cotação no banco de dados - Erro: " + err.Error())
		return nil, err
	}

	return db, nil
}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Requisição da cotação da moeda iniciada!")

	db, err := buscarBancoDadosCon()
	if err != nil {
		log.Println("main - Erro ao abrir banco de dados - Erro: " + err.Error())
	}

	/*tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into cotacao(code, name, bid) values(?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec("USD", "Dollar", "5.75")
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}


	rows, err := db.Query("select code, name, bid from cotacao")
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var code string
		var name string
		var bid string
		err = rows.Scan(&code, &name, &bid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(code, name, bid)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}*/

	ctx := r.Context()

	defer log.Println("Requisição da cotação da moeda finalizada!")

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		log.Println("001 - BuscaCotacaoHandler - Erro - Path inválido: " + r.URL.Path)
		return
	}

	// Buscar cotação na API
	cotacao, error := BuscaCotacao(ctx)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("002 - BuscaCotacaoHandler - Erro: " + error.Error())
		return
	}

	// Salvar os dados da cotacão recebidos da API
	_, error = SalvarCotacao(ctx, cotacao, db)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("003 - BuscaCotacaoHandler - Erro: " + error.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao.Usdbrl.Bid)

}

func SalvarCotacao(ctx context.Context, cotacao *Cotacao, db *sql.DB) (*Cotacao, error) {
	ctxBancoDados, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into cotacao(code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) values(?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		log.Println("001 - SalvarCotacao - Erro: " + err.Error())
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctxBancoDados, cotacao.Usdbrl.Code, cotacao.Usdbrl.Codein, cotacao.Usdbrl.Name, cotacao.Usdbrl.High, cotacao.Usdbrl.Low, cotacao.Usdbrl.VarBid, cotacao.Usdbrl.PctChange, cotacao.Usdbrl.Bid, cotacao.Usdbrl.Ask, cotacao.Usdbrl.Timestamp, cotacao.Usdbrl.CreateDate)
	if err != nil {
		log.Println("002 - SalvarCotacao - Erro: " + err.Error())
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Registro da cotação salvo com sucesso!")
	return cotacao, nil
}

func BuscaCotacao(ctx context.Context) (*Cotacao, error) {

	ctxCotacao, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctxCotacao, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if error != nil {
		log.Println("001 - BuscaCotacao - Erro - " + error.Error())
		return nil, error
	}

	resp, error := http.DefaultClient.Do(req)
	if error != nil {
		log.Println("002 - BuscaCotacao - Erro - " + error.Error())
		return nil, error
	}

	defer resp.Body.Close()
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		log.Println("003 - BuscaCotacao - Erro - " + error.Error())
		return nil, error
	}

	var c Cotacao
	error = json.Unmarshal(body, &c)
	if error != nil {
		log.Println("004 - BuscaCotacao - Erro - " + error.Error())
		return nil, error
	}

	return &c, nil
}
