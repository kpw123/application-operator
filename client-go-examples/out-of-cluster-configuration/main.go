package main

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main() {
	// 获取本地用户家目录
	homePath := homedir.HomeDir()
	if homePath == "" {
		log.Fatal("failed to get the home directory")
	}
	// 拼接 家目录 .kube config 拿到完整的 k8s cluster 连接配置地址
	kubeconfig := filepath.Join(homePath, ".kube", "config")

	//从家目录集群连接配置地址生成 config
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		pods, err := clientSet.CoreV1().Pods("default").
			List(context.TODO(), v1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("There are %d pods is the cluster \n", len(pods.Items))
		for i, pod := range pods.Items {
			log.Printf("%d -> %s/%s", i+1, pod.Namespace, pod.Name)
		}
		<-time.Tick(5 * time.Second)
	}

}
