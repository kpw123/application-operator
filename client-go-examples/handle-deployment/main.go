package main

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
	"log"
	"path/filepath"
	"time"
)

func main() {

	homePath := homedir.HomeDir()
	if homePath == "" {
		log.Fatal("failed to get the home directory")
	}

	kubeconfig := filepath.Join(homePath, ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	dpClient := clientSet.AppsV1().Deployments(metav1.NamespaceDefault)

	// 调用函数创建
	log.Println("create Deployment")
	if err := createDeployment(dpClient); err != nil {
		log.Fatal(err)
	}

	// 等待一分钟
	<-time.Tick(1 * time.Minute)

	// 调用函数更新
	log.Println("update Deployment")
	if err := updateDeployment(dpClient); err != nil {
		log.Fatal(err)
	}

	<-time.Tick(1 * time.Minute)

	// 调用函数删除
	log.Println("delete Deployment")
	if err := deleteDeployment(dpClient); err != nil {
		log.Fatal(err)
	}

	<-time.Tick(1 * time.Minute)

	log.Println("end")

}

func createDeployment(dpClient v1.DeploymentInterface) error {

	// 定义副本数
	var replicas int32 = 3

	// 初始化一个 deployment 对象
	newDp := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "nginx-deploy"},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx:1.19", Ports: []corev1.ContainerPort{
							{Name: "http", Protocol: "TCP", ContainerPort: 80},
						}},
					},
				},
			},
		},
	}
	// 创建 deployment
	_, err := dpClient.Create(context.TODO(), newDp, metav1.CreateOptions{})
	return err
}

func updateDeployment(dpClient v1.DeploymentInterface) error {
	// 判断是否存在 存在就拿到当前 deployment 对象
	dp, err := dpClient.Get(context.TODO(), "nginx-deploy", metav1.GetOptions{})
	if err != nil {
		return err
	}
	// 修改第一个 container 的镜像信息
	dp.Spec.Template.Spec.Containers[0].Image = "nginx:1.20"
	// 使用 RetryOnConflict 防止同时更新
	return retry.RetryOnConflict(
		retry.DefaultRetry, func() error {
			// 更新 deployment
			_, err = dpClient.Update(context.TODO(), dp, metav1.UpdateOptions{})
			return err
		})
}

func deleteDeployment(dpClient v1.DeploymentInterface) error {
	deletePolicy := metav1.DeletePropagationForeground
	return dpClient.Delete(
		context.TODO(), "nginx-deploy",
		metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
}
