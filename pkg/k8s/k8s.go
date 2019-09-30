package k8s

import (
	"flag"
	"github.com/manifoldco/promptui"
	corev1 "k8s.io/api/core/v1"
	crdv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

func New() (*K8s, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8s{client: clientset, config: config}, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows

}

type K8s struct {
	client *kubernetes.Clientset
	config *restclient.Config
}

func (k *K8s) SelectKind() (string, error) {
	kinds := []string{"node", "pod", "crd"}
	prompt := promptui.Select{
		Label:             "Kinds",
		Items:             kinds,
		StartInSearchMode: true,
	}

	_, item, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return item, nil
}

func (k *K8s) SelectNode() (*corev1.Node, error) {
	nodes, err := k.client.CoreV1().Nodes().List(metav1.ListOptions{})
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
		node := nodes.Items[index]
		name := strings.Replace(strings.ToLower(node.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Nodes",
		Items:             nodes.Items,
		Searcher:          searcher,
		StartInSearchMode: true,
		Templates:         templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &nodes.Items[i], nil
}

func (k *K8s) SelectPod() (*corev1.Pod, error) {
	pods, err := k.client.CoreV1().Pods("kube-system").List(metav1.ListOptions{})
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
		pod := pods.Items[index]
		name := strings.Replace(strings.ToLower(pod.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Pods",
		Items:             pods.Items,
		Searcher:          searcher,
		StartInSearchMode: true,
		Templates:         templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &pods.Items[i], nil
}

func (k *K8s) SelectCRD() (*crdv1beta1.CustomResourceDefinition, error) {
	client, err := clientset.NewForConfig(k.config)
	if err != nil {
		return nil, err
	}
	crds, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().List(metav1.ListOptions{})
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
		crd := crds.Items[index]
		name := strings.Replace(strings.ToLower(crd.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Pods",
		Items:             crds.Items,
		Searcher:          searcher,
		StartInSearchMode: true,
		Templates:         templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &crds.Items[i], nil
}
