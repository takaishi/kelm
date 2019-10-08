package k8s

import (
	"github.com/manifoldco/promptui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

func New(kubeconfigPath string) (*K8s, error) {

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &K8s{client: clientset, config: config, namespace: "default"}, nil
}

type K8s struct {
	client    *kubernetes.Clientset
	config    *restclient.Config
	namespace string
	kind      string
}

func (k *K8s) SetNamespace(n string) {
	k.namespace = n
}

func (k *K8s) GetNamespace() string {
	return k.namespace
}

func (k *K8s) SetKind(s string) {
	k.kind = s
}

func (k *K8s) GetKind() string {
	return k.kind
}

func (k *K8s) SelectNamespace() (string, error) {
	r, err := k.client.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
		Details: `
--------- Namespace ----------
{{ "Name:" | faint }}	{{ .Name }}`,
	}

	searcher := func(input string, index int) bool {
		namespace := r.Items[index]
		name := strings.Replace(strings.ToLower(namespace.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Namespace?",
		Items:             r.Items,
		StartInSearchMode: true,
		Templates:         templates,
		Searcher:          searcher,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return r.Items[i].Name, nil
}

func (k *K8s) SelectKind() (string, error) {
	kinds := []metav1.APIResource{}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(k.config)
	if err != nil {
		return "", err
	}
	lists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return "", err
	}

	for _, list := range lists {
		for _, resource := range list.APIResources {
			kinds = append(kinds, resource)
		}
	}

	searcher := func(input string, index int) bool {
		kind := kinds[index]
		name := strings.Replace(strings.ToLower(kind.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
		Details: `
--------- Kind ----------
{{ "Name:" | faint }}	{{ .Name }}`,
	}

	prompt := promptui.Select{
		Label:             "Kinds",
		Items:             kinds,
		Searcher:          searcher,
		Templates:         templates,
		StartInSearchMode: true,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return kinds[i].Name, nil
}

func (k *K8s) SelectObjects(kind string) (*runtime.Object, error) {
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	r := f.NewBuilder().
		Unstructured().
		NamespaceParam(k.namespace).DefaultNamespace().AllNamespaces(false).
		ResourceTypeOrNameArgs(true, []string{kind}...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()
	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
		Details: `
--------- Object ----------
{{ "Name:" | faint }}	{{ .Name }}`,
	}

	searcher := func(input string, index int) bool {
		info := infos[index]
		name := strings.Replace(strings.ToLower(info.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:             "Objects",
		Items:             infos,
		Searcher:          searcher,
		StartInSearchMode: true,
		Templates:         templates,
	}

	i, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	return &infos[i].Object, nil
}
