package main

import (
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	clientset, err := kubeClient()
	if err != nil {
		return err
	}

	pod, err := selectPod(clientset)
	if err != nil {
		return err
	}

	actionPrompt := promptui.Select{
		Label: "actions",
		Items: []string{"get", "describe"},
	}

	_, actionName, err := actionPrompt.Run()
	if err != nil {
		return err
	}
	out, err := exec.Command("kubectl", "-n", "kube-system", actionName, "pod", pod.Name).CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func kubeClient() (*kubernetes.Clientset, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func selectPod(clientset *kubernetes.Clientset) (*corev1.Pod, error) {
	pods, err := clientset.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
		Details: `
--------- Pepper ----------
{{ "Name:" | faint }}	{{ .Name }}`,
	}

	searcher := func(input string, index int) bool {
		pepper := pods.Items[index]
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Pods",
		Items:     pods.Items,
		Searcher:  searcher,
		Templates: templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &pods.Items[i], nil
}
