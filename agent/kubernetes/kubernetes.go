package kubernetes

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesAgent struct {
	interval           time.Duration
	kubeClient         *kubernetes.Clientset
	nodeSharedInformer *cache.SharedInformer
	podSharedInformer  *cache.SharedInformer
}

func NewKubernetesAgent(interval time.Duration) *KubernetesAgent {
	return &KubernetesAgent{
		interval: interval,
	}
}

func (agent *KubernetesAgent) Run() {
	log.Info("starting kubernetes agent...")

	err := agent.init()
	if err != nil {
		log.Fatal("error initializing agent")
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	done := false

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done = true
	}()

	tick := time.Tick(agent.interval)
	for _ = range tick {
		log.Info("collecting...")

		agent.collect()

		if done {
			log.Info("stopping kubernetes agent...")
			os.Exit(0)
		}
	}
}

func (agent *KubernetesAgent) init() error {
	kubeconfig := "/Users/zacblazic/.kube/custom_config/sandbox_kube_config.yml"

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	agent.kubeClient = clientset
	agent.nodeSharedInformer = agent.createNodeSharedInformer()
	agent.podSharedInformer = agent.createPodSharedInformer()

	return nil
}

func (agent *KubernetesAgent) createNodeSharedInformer() *cache.SharedInformer {
	lw := cache.NewListWatchFromClient(agent.kubeClient.CoreV1().RESTClient(), "nodes", metav1.NamespaceAll, fields.Everything())
	si := cache.NewSharedInformer(lw, &v1.Node{}, 5*time.Minute)

	go si.Run(context.Background().Done())

	return &si
}

func (agent *KubernetesAgent) createPodSharedInformer() *cache.SharedInformer {
	lw := cache.NewListWatchFromClient(agent.kubeClient.CoreV1().RESTClient(), "pods", metav1.NamespaceAll, fields.Everything())
	si := cache.NewSharedInformer(lw, &v1.Pod{}, 5*time.Minute)

	go si.Run(context.Background().Done())

	return &si
}

// // cloud agnostic
// type NodeCostInfo struct {
// 	CostPerHour decimal.Decimal
// }
//
// func (nci *NodeCostInfo) CPUCostPerHour() decimal.Decimal {
//
// }
//
// func (nci *NodeCostInfo) MemoryCostPerHour() decimal.Decimal {
//
// }
//
// type PodCostInfo struct{}

func (agent *KubernetesAgent) collect() {

	// initial list of nodes
	nodeList := (*agent.nodeSharedInformer).GetStore().List()

	// // enrich node info with costs
	// for node := range nodeList {
	// 	// get node info from cloud provider
	// 	cloudprovider.Pricing().GetNodeCost(node)
	// }

	podList := (*agent.podSharedInformer).GetStore().List()

	log.Infof("There are %d nodes in the cluster", len(nodeList))
	log.Infof("There are %d pods in the cluster", len(podList))
}
