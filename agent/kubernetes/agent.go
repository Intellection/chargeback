package kubernetes

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Agent will periodically collect and store cost information about a cluster
type Agent struct {
	interval       time.Duration
	influxdbClient client.Client
	clientset      kubernetes.Interface
	costService    *CostService
}

// NewAgent creates a Kubernetes Agent
func NewAgent(influxdbClient client.Client, interval time.Duration) *Agent {
	costService := NewCostService(influxdbClient)

	return &Agent{
		interval:       interval,
		influxdbClient: influxdbClient,
		costService:    costService,
	}
}

// Run starts the main control loop of the Agent
func (agent *Agent) Run() {
	log.Info("starting the kubernetes agent...")
	quitting := false

	err := agent.init()
	if err != nil {
		log.Fatal(err)
		log.Fatal("error initializing the kubernetes agent")
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)

		if sig == syscall.SIGINT {
			// For ease of development
			log.Info("Received SIGINT")
			log.Info("immediately force stopping the kubernetes agent...")
			os.Exit(0)
		}

		if sig == syscall.SIGTERM {
			log.Info("Received SIGTERM")
			quitting = true
		}

	}()

	for {
		startTime := time.Now()

		if quitting {
			log.Info("stopping the kubernetes agent...")
			os.Exit(0)
		}

		// may take longer than 30 seconds
		agent.collect()

		endTime := time.Now()
		elapsedTime := endTime.Sub(startTime)
		sleepTime := agent.interval - elapsedTime

		log.Infof("elapsed time: %v", elapsedTime)

		if sleepTime >= 0 {
			log.Infof("sleeping for: %v at %v", sleepTime, time.Now())
			time.Sleep(sleepTime)
		} else {
			log.Warn("you should consider increasing your interval seconds to get accurate info")
		}

		log.Infof("writing at: %v", time.Now())

		agent.write()
	}
}

func (agent *Agent) init() error {

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	agent.clientset = clientset

	return nil
}

func (agent *Agent) collect() {
	agent.costService = NewCostService(agent.influxdbClient)

	nodes, err := agent.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
	}

	pods, err := agent.clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
	}

	agent.costService.processRawData(nodes.Items, pods.Items)
	agent.costService.calculatePodCosts()
}

func (agent *Agent) write() {
	agent.costService.storePodCosts()
}
