package kubernetes

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesAgent struct {
	interval       time.Duration
	influxdbClient client.Client
	clientset      kubernetes.Interface
}

func NewKubernetesAgent(influxdbClient client.Client, interval time.Duration) *KubernetesAgent {
	return &KubernetesAgent{
		interval:       interval,
		influxdbClient: influxdbClient,
	}
}

func (agent *KubernetesAgent) Run() {
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

	tick := time.Tick(agent.interval)
	for _ = range tick {
		log.Info("scraping info from kubernetes...")

		agent.collect()

		if quitting {
			log.Info("stopping the kubernetes agent...")
			os.Exit(0)
		}
	}
}

func (agent *KubernetesAgent) init() error {

	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	kubeconfig := home + "/.kube/custom_config/sandbox_kube_config.yml"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)

	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return err
	}

	agent.clientset = clientset

	return nil
}

func (agent *KubernetesAgent) collect() {

	costService := NewCostService()

	nodes, err := agent.clientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
	}

	pods, err := agent.clientset.CoreV1().Pods(metav1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		log.Error(err)
	}

	costService.process(nodes.Items, pods.Items, agent.influxdbClient)
}
