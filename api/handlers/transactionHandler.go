package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ilovepitsa/Transaction/api/rabbit"
	"github.com/ilovepitsa/Transaction/api/repo"
	pb "github.com/ilovepitsa/protobufForTestCase"
)

type TransactionHandler struct {
	l             *log.Logger
	repository    *repo.TransactionRepository
	rabbitHandler *rabbit.RabbitHandler
}

func NewTransactionHandler(l *log.Logger, repos *repo.TransactionRepository, rabbitHandler *rabbit.RabbitHandler) *TransactionHandler {
	return &TransactionHandler{l: l, repository: repos, rabbitHandler: rabbitHandler}
}

func (x *TransactionHandler) process(trans *pb.Transaction) {
	x.repository.Add(trans)
	err := x.rabbitHandler.PublishTrans(trans)
	if err != nil {
		x.repository.SetError(trans)
		return
	}
}

func getCurrencyProto(currency string) pb.CurrencyType {

	usd := regexp.MustCompile(`^[uU][sS][dD]$`)
	usdt := regexp.MustCompile(`^[uU][sS][dD][tT]$`)
	eur := regexp.MustCompile(`^[eE][uU][rR]$`)
	rub := regexp.MustCompile(`^[rR][uU][bB]$`)
	btc := regexp.MustCompile(`^[bB][tT][cC]$`)
	// res := rub.FindString(currency)
	fmt.Printf(".%s.", currency)
	switch {
	case usd.MatchString(currency):
		return pb.CurrencyType_CURRENCY_USD
	case usdt.MatchString(currency):
		return pb.CurrencyType_CURRENCY_USDT
	case eur.MatchString(currency):
		return pb.CurrencyType_CURRENCY_EUR
	case rub.MatchString(currency):

		return pb.CurrencyType_CURRENCY_RUB
	case btc.MatchString(currency):
		return pb.CurrencyType_CURRENCY_BTC
	}
	return pb.CurrencyType_CURRENCY_UNDEFINED
}

func (x *TransactionHandler) Invoice(w http.ResponseWriter, r *http.Request) {

	// x.l.Println(r.URL)
	if !r.URL.Query().Has("currency") {
		fmt.Fprintln(w, "add currency")
		return
	}
	currency := getCurrencyProto(r.URL.Query().Get("currency"))

	if currency == pb.CurrencyType_CURRENCY_UNDEFINED {
		fmt.Fprintln(w, "error currency")
		return
	}

	if !r.URL.Query().Has("amount") {
		fmt.Fprintln(w, "add amount")
		return
	}
	amount, err := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)
	if err != nil {
		x.l.Println("Error parsing amount: ", err)
		return
	}

	if amount < 0.0 {
		fmt.Fprintln(w, "amount must be gr than zero")
		return
	}

	if !r.URL.Query().Has("account") {
		fmt.Fprintln(w, "add account")
		return
	}

	account := r.URL.Query().Get("account")
	trans := pb.Transaction{
		Currency:       currency,
		Number_Invoice: account,
		Action:         pb.ActionType_ACTION_ADD,
		Amount:         amount}
	x.process(&trans)
	x.l.Println(trans.String())
}

func (x *TransactionHandler) Withdraw(w http.ResponseWriter, r *http.Request) {

	x.l.Println(r.URL.Query())
	if !r.URL.Query().Has("currency") {
		fmt.Fprintln(w, "add currency")
		return
	}
	currency := getCurrencyProto(r.URL.Query().Get("currency"))

	if currency == pb.CurrencyType_CURRENCY_UNDEFINED {
		fmt.Fprintln(w, "error currency")
		return
	}

	if !r.URL.Query().Has("amount") {
		fmt.Fprintln(w, "add amount")
		return
	}
	amount, err := strconv.ParseFloat(r.URL.Query().Get("amount"), 64)
	if err != nil {
		fmt.Println("Error parsing amount: ", err)
		return
	}

	if amount < 0.0 {
		fmt.Fprintln(w, "amount must be gr than zero")
		return
	}

	if !r.URL.Query().Has("account") {
		fmt.Fprintln(w, "add account")
		return
	}
	account := r.URL.Query().Get("account")

	if !r.URL.Query().Has("accountTo") {
		fmt.Fprintln(w, "add accountTo")
		return
	}

	accountTo := r.URL.Query().Get("accountTo")

	trans := pb.Transaction{
		Currency:         currency,
		Number_Invoice:   account,
		Action:           pb.ActionType_ACTION_SUB,
		Amount:           amount,
		Number_InvoiceTo: accountTo,
	}
	x.process(&trans)
	x.l.Println(trans.String())
}
