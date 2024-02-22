package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/ilovepitsa/Transaction/api/handlers"
	"github.com/ilovepitsa/Transaction/api/rabbit"
	"github.com/ilovepitsa/Transaction/api/repo"
	pb "github.com/ilovepitsa/protobufForTestCase"
	_ "github.com/lib/pq"
)

func fillTest(transRepo repo.TransactionRepository, l *log.Logger) {

	test := pb.Transaction{
		Id:               1,
		Currency:         pb.CurrencyType_CURRENCY_BTC,
		Number_Invoice:   "123123123",
		Amount:           10.0,
		Action:           pb.ActionType_ACTION_ADD,
		Number_InvoiceTo: "",
	}

	err := transRepo.Add(&test)
	if err != nil {
		l.Print(err)
	}
}
func main() {
	l := log.New(os.Stdout, "Transaction ", log.LstdFlags)
	l.SetFlags(log.LstdFlags | log.Lshortfile)

	connStr := "user=postgres password=123 dbname=TransactionSystem sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		l.Print(err)
		return
	}
	defer db.Close()

	transRepo := repo.NewTransactionRepository(db, l)
	// fillTest(*transRepo, l)

	rabbitHandler := rabbit.NewRabbitHandler(l, transRepo)
	err = rabbitHandler.Init(rabbit.RabbitParameters{
		Login:    "transaction",
		Password: "transaction",
		Ip:       "localhost",
		Port:     "5672"})

	if err != nil {
		l.Println("Cant create rabbitHandler", err)
	}
	defer rabbitHandler.Close()

	go rabbitHandler.Consume()

	transHandler := handlers.NewTransactionHandler(l, transRepo, rabbitHandler)

	sm := mux.NewRouter()
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	// postRouter := sm.Methods(http.MethodPost).Subrouter()

	getRouter.HandleFunc("/invoice", transHandler.Invoice)
	getRouter.HandleFunc("/withdraw", transHandler.Withdraw)

	l.Printf("Starting  on port %v.... \n", 8081)
	srv := http.Server{
		Addr:         ":8081",
		Handler:      sm,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	srv.ListenAndServe()

}
