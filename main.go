package main

import (
	"log"
	"os"
	"os/signal"
	"products/internal/cache"
	"products/internal/config"
	"products/internal/database"
	"products/internal/migrations"
	"products/internal/server"
	"syscall"
	"time"

	_ "products/docs"

	"github.com/hudl/fargo"
	"github.com/op/go-logging"
)

func heartBeat(conn fargo.EurekaConnection, instance fargo.Instance, l *logging.Logger) {
	for {
		err := conn.HeartBeatInstance(&instance)
		if err != nil {
			l.Errorf("Heartbeat failed:", err)
		} else {
			l.Info("Heartbeat sent")
		}

		time.Sleep(30 * time.Second)
	}
}

// @title products microservice API
// @version 1.0
// @description This is a products server for ecomm micro project
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:42069
// @BasePath /
func main() {
	config.Init()

	f, err := os.OpenFile(config.Config().LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatalf("unable to open log file %v", config.Config().LogFile)
	}
	defer f.Close()

	backend := logging.NewLogBackend(f, "", 0)
	logging.SetBackend(backend)

	serviceRegistry := config.Config().ServiceRegistry
	c := fargo.NewConn(serviceRegistry)
	instance := fargo.Instance{
		InstanceId:       "products-service",
		HostName:         config.Config().EurekaHostname,
		App:              "PRODUCTS-SERVICE",
		IPAddr:           "localhost",
		VipAddress:       "PRODUCTS-SERVICE",
		SecureVipAddress: "PRODUCTS-SERVICE",
		Status:           fargo.UP,
		Port:             42069,
		PortEnabled:      true,
		DataCenterInfo: fargo.DataCenterInfo{
			Name: fargo.MyOwn,
		},
		LeaseInfo: fargo.LeaseInfo{
			RenewalIntervalInSecs: 30,
			DurationInSecs:        90,
		},
	}

	// Register with Eureka
	err = c.RegisterInstance(&instance)
	if err != nil {
		log.Fatal("Failed to register:", err)
	}

	l := logging.MustGetLogger("products")
	go heartBeat(c, instance, l)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("deregistering from eureka")
		if err := c.DeregisterInstance(&instance); err != nil {
			log.Printf("Unable to deregister from eureka : %v\n", err)
		} else {
			log.Println("Deregistered from eureka")
		}

		log.Println("Disconnecting from DB")
		if err := database.Disconnect(); err != nil {
			log.Printf("unable to disconnect from db :%v\n", err)
		}
		log.Println("Disconnected from DB")

		os.Exit(0)
	}()

	database.Connect()
	cache.Connect()
	migrations.AutoMigrate()

	server.SetUp()

	app := server.New()

	port := config.Config().Port
	if err := app.Listen(port); err != nil {
		log.Fatalf("err : %v", err)
	}
}
