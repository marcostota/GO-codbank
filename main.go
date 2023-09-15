package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/marcostota/codebank/infrastructure/grpc/server"
	"github.com/marcostota/codebank/infrastructure/kafka"
	"github.com/marcostota/codebank/infrastructure/repository"
	"github.com/marcostota/codebank/usecase"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
}

func main() {
	db := setupDb()
	defer db.Close()
	producer := setupKafkaProducer()
	processTransactionUseCase := setupTransactionUseCase(db, producer)
	fmt.Println("Rodando gRPC server")
	serveGrpc(processTransactionUseCase)
}

func setupDb() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("host"),
		os.Getenv("port"),
		os.Getenv("user"),
		os.Getenv("password"),
		os.Getenv("dbname"),
	)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("error connection to database")
	}
	return db
}

func setupTransactionUseCase(db *sql.DB, producer kafka.KafkaProducer) usecase.UseCaseTransaction {
	transactionRepository := repository.NewTransactionRepositoryDb(db)
	usecase := usecase.NewUseCaseTransaction(transactionRepository)
	usecase.KafkaProducer = producer
	return *usecase
}

func setupKafkaProducer() kafka.KafkaProducer {
	producer := kafka.NewKafkaProducer()
	producer.SetupProducer(os.Getenv("KafkaBootstrapServers"))
	return producer
}

func serveGrpc(processTransactionUseCase usecase.UseCaseTransaction) {
	grpcServer := server.NewGRPCServer()
	grpcServer.ProcessTransactionUseCase = processTransactionUseCase
	grpcServer.Serve()
}

// Create credit card
// cc := domain.NewCreditCard()
// cc.Number = "1234"
// cc.Name = "tota"
// cc.ExpirationYear = 2030
// cc.ExpirationMonth = 4
// cc.CVV = 123
// cc.Limit = 1000
// cc.Balance = 0

// repo := repository.NewTransactionRepositoryDb(db)
// repo.CreateCreditCard(*cc)
