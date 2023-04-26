package main

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"time"
)

func main() {
	// Need to package as a container image and run in a cluster

	// 从 InClusterConfig() 方法中获取到 pod 中挂载的 rbac 信息
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	// 使用认证信息从 cluster 中拿到 clientSet 对象
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// 通过几层嵌套方法调用 获取到 []pod 切片
		pods, err := clientSet.CoreV1().Pods("default").
			List(context.TODO(), v1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("There are %d pods is the cluster \n", len(pods.Items))
		for i, pod := range pods.Items {
			log.Printf("%d -> %s/%s", i+1, pod.Namespace, pod.Name)
		}

		// 通过tick方法产生一个管道 5s后管道放入值 等待结束
		<-time.Tick(5 * time.Second)
	}

}
