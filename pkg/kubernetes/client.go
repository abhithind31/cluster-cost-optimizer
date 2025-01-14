package kubernetes

import (
    "flag"
    "os"
    "path/filepath"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

func GetClientSet() (*kubernetes.Clientset, *rest.Config, error) {
    var config *rest.Config
    var err error

    if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
        // Running inside the cluster
        config, err = rest.InClusterConfig()
    } else {
        var kubeconfig *string
        home, err := os.UserHomeDir()
        if err != nil {
            return nil, nil, err
        }

        if home != "" {
            kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
        } else {
            kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
        }
        flag.Parse()

        config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
    }

    if err != nil {
        return nil, nil, err
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        return nil, nil, err
    }

    return clientset, config, nil
}