package main

import (
	// golang package
	"context"
	"encoding/json"
	"fmt"
	"paymentfc/cmd/payment/handler"
	"paymentfc/cmd/payment/repository"
	"paymentfc/cmd/payment/resource"
	"paymentfc/cmd/payment/service"
	"paymentfc/cmd/payment/usecase"
	"paymentfc/config"
	"paymentfc/grpc"
	"paymentfc/infrastructure/constant"
	"paymentfc/infrastructure/log"
	"paymentfc/kafka"
	"paymentfc/models"
	"paymentfc/routes"

	// external package
	"github.com/gin-gonic/gin"
)

// main main.
func main() {
	cfg := config.LoadConfig()
	cfg = config.LoadSecretConfig(cfg)
	// redis := resource.InitRedis(&cfg)
	db := resource.InitDB(&cfg)
	kafkaWriter := kafka.NewWriter(cfg.Kafka.Broker, cfg.Kafka.KafkaTopics[constant.KafkaTopicPaymentSuccess])

	log.SetupLogger()

	grpcUserClient := grpc.NewUserClient()
	databaseRepository := repository.NewPaymentDatabase(db)
	publisherRepository := repository.NewKafkaPublisher(kafkaWriter)
	paymentService := service.NewPaymentService(databaseRepository, publisherRepository)
	paymentUsecase := usecase.NewPaymentUsecase(paymentService)
	paymentHandler := handler.NewPaymentHandler(paymentUsecase, cfg.Xendit.XenditWebhookToken)

	xenditRepository := repository.NewXenditClient(cfg.Xendit.XenditAPIKey)
	xenditService := service.NewXenditService(grpcUserClient, databaseRepository, xenditRepository)
	xenditUsecase := usecase.NewXenditUsecase(xenditService)

	// scheduler
	scheduler := service.SchedulerService{
		Database:       databaseRepository,
		Xendit:         xenditRepository,
		Publisher:      publisherRepository,
		PaymentService: paymentService,
	}

	scheduler.StartCheckPendingInvoices()
	scheduler.StartProcessPendingPaymentRequests()
	scheduler.StartProcessFailedPaymentRequests()
	scheduler.StartSweepingExpiredPendingPayments()

	// REMOVE ME - DEBUG PURPOSE ONLY
	tempDebug69, _ := json.Marshal(cfg.Kafka)
	fmt.Printf("\n==== DEBUG main.go - Line: 69 ===== \n\n%s\n\n=====================\n\n\n", string(tempDebug69))
	// END OF REMOVE ME

	// kafka consumer
	// potential less efficient --> traffic gede
	kafka.StartOrderConsumer(cfg.Kafka.Broker, cfg.Kafka.KafkaTopics[constant.KafkaTopicOrderCreated],
		func(event models.OrderCreatedEvent) {
			if cfg.Toggle.DisableCreateInvoiceDirectly {
				err := paymentUsecase.ProcessPaymentRequests(context.Background(), event)
				if err != nil {
					log.Logger.Println("Failed Handling Order Created Event: ", err.Error())
				}
			} else {
				err := xenditUsecase.CreateInvoice(context.Background(), event)
				if err != nil {
					log.Logger.Println("Failed Handling Order Created Event: ", err.Error())
				}
			}
		})

	// current condition
	/*
		- user checkout order
		- order execute checkout --> publish event order.created
		- payment service akan memproses create invoice
	*/

	// new condition
	/*
		- user checkout order
		- order publish event order.created
		- payment service akan simpan event yang dari order.created
		- payment akan menyediakan background process utk create invoice per batch
	*/

	// cons: - data tidak di-execute secara real time
	// pertimbangan: transactional especially payment processes --> harus lebih fokus ke consistency dan stability

	// pro:
	/*
		sample scenario:
			- xendit team informed there will be maintenance for 5 minutes (12:00 - 12:05)
			- kita bisa hold execute payment_requests sampai xendit stable
			- data dari order service (order.created) tidak menumpuk
	*/

	port := cfg.App.Port
	router := gin.Default()
	routes.SetupRoutes(router, paymentHandler)
	router.Run(":" + port)

	log.Logger.Printf("Server running on port: %s", port)