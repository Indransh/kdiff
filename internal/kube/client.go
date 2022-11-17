package kube

import (
	"fmt"
	"kdiff/internal/helpers"
	"sort"

	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// Using same values as hardcoded in kubectl
	// https://github.com/kubernetes/kubernetes/blob/e39a0af5ce0a836b30bd3cce237778fb4557f0cb/staging/src/k8s.io/kubectl/pkg/cmd/cmd.go#L94
	clientQPS   = 50
	clientBurst = 300
)

var clientSets = make(map[string]*kubernetes.Clientset)

type KubeResources struct {
	AppsV1Resource *appsv1.AppsV1Interface
	CoreV1Resource *corev1.CoreV1Interface
}

type KubeConfig struct {
	configConfig   *clientcmd.ClientConfig
	kubeConfigPath string
	Client         Client
}

type Client interface {
	GetNamespaces(ctx string)
	GetDeployments(ctx, ns string)
	GetDaemonSets(ctx, ns string)
	GetStatefulSets(ctx, ns string)
}

type KubeContext struct {
	Name             string
	Reachable        bool
	UnreachableError error
	ServerVersion    string
}

// ParseConfig loads a kubeconfig file.
func (k *KubeConfig) ParseConfig(kubeConfigPath string) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfigPath},
		&clientcmd.ConfigOverrides{})
	k.configConfig = &clientConfig
	k.kubeConfigPath = kubeConfigPath
}

// GetContextNames returns a list of contexts from the kubeconfig file.
func (k *KubeConfig) GetContextNames() []string {
	var listContexts []string

	rawConfig, err := (*k.configConfig).RawConfig()
	helpers.HandleError(err)
	for context := range rawConfig.Contexts {
		listContexts = append(listContexts, context)
	}

	sort.Strings(listContexts)
	return listContexts
}

// InitializeClients initializes clients for each context found in kubeconfig file.
func (k *KubeConfig) InitializeClients() {
	for _, ctx := range k.GetContextNames() {
		clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.kubeConfigPath},
			&clientcmd.ConfigOverrides{
				CurrentContext: ctx,
			}).ClientConfig()
		helpers.HandleError(err)

		// Configure rate limits (default value is 5 QPS which is too low)
		clientConfig.QPS = clientQPS
		clientConfig.Burst = clientBurst

		clientSets[ctx], err = kubernetes.NewForConfig(clientConfig)
		helpers.HandleError(err)
	}
}

// GetContextInfo returns information about all contexts.
func (k *KubeConfig) GetContextInfo() []*KubeContext {
	var (
		allContexts []*KubeContext
		chanVersion = make(chan *KubeContext)
	)
	for _, ctx := range k.GetContextNames() {
		// Although clients are already initialized, we initialize again with a lower timeout.
		clientConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.kubeConfigPath},
			&clientcmd.ConfigOverrides{
				CurrentContext: ctx,
				Timeout:        "2",
			}).ClientConfig()
		helpers.HandleError(err)

		clientSet, err := kubernetes.NewForConfig(clientConfig)
		helpers.HandleError(err)
		go getServerVersion(ctx, clientSet, chanVersion)
	}

	for range k.GetContextNames() {
		allContexts = append(allContexts, <-chanVersion)
	}
	close(chanVersion)
	return allContexts
}

// getServerVersion performs an API call to get the server version for the provided
// clientSet and sends the result back to a channel.
func getServerVersion(ctx string, clientSet *kubernetes.Clientset, out chan<- *KubeContext) {
	version, err := clientSet.Discovery().ServerVersion()

	var context = KubeContext{
		Name:             ctx,
		UnreachableError: err,
	}

	if err == nil {
		context.Reachable = true
		context.ServerVersion = fmt.Sprintf("%s.%s", version.Major, version.Minor)
	} else {
		context.Reachable = false
	}
	out <- &context
}
