package rabbit

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ilovepitsa/Transaction/api/repo"
	pb "github.com/ilovepitsa/protobufForTestCase"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
)

type RabbitParameters struct {
	Login    string
	Password string
	Ip       string
	Port     string
}

type RabbitHandler struct {
	l                    *log.Logger
	cr                   *repo.TransactionRepository
	connection           *amqp.Connection
	channel              *amqp.Channel
	requestQueue         amqp.Queue
	responceQueue        amqp.Queue
	accountResponceQueue amqp.Queue
}

func NewRabbitHandler(l *log.Logger, cr *repo.TransactionRepository) *RabbitHandler {
	return &RabbitHandler{l: l, cr: cr}
}

func (rb *RabbitHandler) Init(param RabbitParameters) error {
	var err error
	rb.connection, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", param.Login, param.Password, param.Ip, param.Port))
	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.channel, err = rb.connection.Channel()
	if err != nil {
		// rb.l.Println(err)
		return err
	}

	err = rb.channel.ExchangeDeclare("transaction", "topic", false, false, false, false, amqp.Table{})
	if err != nil {
		// rb.l.Println(err)
		return err
	}
	err = rb.channel.ExchangeDeclare("account", "topic", false, false, false, false, amqp.Table{})
	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.requestQueue, err = rb.channel.QueueDeclare("transactionRequest", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}
	err = rb.channel.QueueBind(rb.requestQueue.Name, "request", "transaction", false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.responceQueue, err = rb.channel.QueueDeclare("transactionReqsponce", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	err = rb.channel.QueueBind(rb.responceQueue.Name, "responce", "transaction", false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	rb.accountResponceQueue, err = rb.channel.QueueDeclare("accountResponceT", false, false, false, false, amqp.Table{})

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	err = rb.channel.Qos(
		1,
		0,
		false,
	)

	if err != nil {
		// rb.l.Println(err)
		return err
	}

	return nil
}

func (rb *RabbitHandler) Consume() {
	consumeRequestChan, err := rb.channel.Consume(rb.requestQueue.Name, "", true, false, false, false, amqp.Table{})

	if err != nil {
		rb.l.Println(err)
		return
	}

	consumeAccountResponce, err := rb.channel.Consume(rb.accountResponceQueue.Name, "", true, false, false, false, amqp.Table{})

	if err != nil {
		rb.l.Println(err)
		return
	}
	var forever chan struct{}

	go func() {
		for d := range consumeRequestChan {
			rb.l.Println("Recieve new request: ", d.Body)
			rb.l.Println("I dont have any methods for this request")
		}
	}()

	respAccount := pb.ResponceAccount{}
	go func() {
		for resp := range consumeAccountResponce {
			// rb.l.Println("Recieve new responce from account: ", resp.Body)
			err := proto.Unmarshal(resp.Body, &respAccount)
			if err != nil {
				rb.l.Println("Error parsing responce account: ")
				continue
			}
			rb.l.Println(respAccount.String())
		}
	}()

	rb.l.Println("Waiting commands")
	<-forever
}

func (rb *RabbitHandler) Close() {
	rb.channel.Close()
	rb.connection.Close()
}

func (rb *RabbitHandler) PublishTrans(trans *pb.Transaction) error {
	var requestTrans pb.RequestTransaction = pb.RequestTransaction{Transaction: trans}
	reqAccount := pb.RequestAccount{Req: &pb.RequestAccount_ReqTrans{ReqTrans: &requestTrans}}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := proto.Marshal(&reqAccount)

	if err != nil {
		return err
	}

	err = rb.channel.PublishWithContext(ctx,
		"account",
		"request",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        req,
		})
	if err != nil {
		// rb.l.Println(err)
		return err
	}
	return nil
}
