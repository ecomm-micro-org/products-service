package main

import (
	"log"
	"os"
	"os/signal"
	"products/cmd/app"
	"products/internal/config"
	"products/internal/database"
	"syscall"
	"time"

	"github.com/hudl/fargo"
)

func heartBeat(conn fargo.EurekaConnection, instance fargo.Instance) {
	for {
		err := conn.HeartBeatInstance(&instance)
		if err != nil {
			log.Println("Heartbeat failed:", err)
		} else {
			log.Println("Heartbeat sent")
		}

		time.Sleep(30 * time.Second)
	}
}

// TODO : implement swagger documentation
func main() {
	config.Init()
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
	err := c.RegisterInstance(&instance)
	if err != nil {
		log.Fatal("Failed to register:", err)
	}

	go heartBeat(c, instance)

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

	app.SetUp()
}
