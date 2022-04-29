package main

import (
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tre-code-analysis/signals"
	"k8s.io/klog/v2"
	"time"
)

func main() {
	klog.InitFlags(nil)
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/trecool/.kube/config")
	if err != nil {
		panic(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	stopCh := signals.SetupSignalHandler()

	informerFactory := informers.NewSharedInformerFactory(clientSet, 5 * time.Second)
	informer := informerFactory.Core().V1().Pods()

	controller := NewController(clientSet, informer)

	informerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		klog.Fatalf("Error running controller: %s", err.Error())
	}

}