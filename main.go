package main

import (
	"flag"
	"fmt"
	pytorchversioned "github.com/kubeflow/training-operator/pkg/client/clientset/versioned"
	"github.com/shuangkun/argo-workflows-pytorch-plugin/controller"
	"k8s.io/klog/v2"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type option struct {
	port int
}

func main() {
	opt := &option{}
	cmd := &cobra.Command{
		Use:  "argo-pytorch-plugin",
		RunE: opt.runE,
	}
	flags := cmd.Flags()
	flags.IntVarP(&opt.port, "port", "", 3008, "The port of the HTTP server")
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func (o *option) runE(c *cobra.Command, args []string) (err error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		if config, err = rest.InClusterConfig(); err != nil {
			panic(err.Error())
		}
	}

	ct := &controller.PytorchJobController{}
	pytorchClient := getPytorchClient(config)

	ct.PytorchClient = pytorchClient
	router := gin.Default()
	router.POST("/api/v1/template.execute", ct.ExecutePytorchJob)
	if err := router.Run(fmt.Sprintf(":%d", o.port)); err != nil {
		klog.Fatal("Failed to start server:", err)
	}
	return
}

// GetPytorchClient get a clientset for Pytorch Job.
func getPytorchClient(restConfig *rest.Config) *pytorchversioned.Clientset {
	clientset, err := pytorchversioned.NewForConfig(restConfig)
	klog.Info(clientset.ServerVersion())
	if err != nil {
		klog.Fatal(err)
	}
	return clientset
}
