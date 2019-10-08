package actions

import (
	"bytes"
	"encoding/json"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/jsonpath"
	"os"
	"strings"

	"text/template"
)

func NewActionRunner(namespace string, confPath string) (*ActionRunner, error) {
	defaultActions := []Action{
		{
			Name:    "get",
			Command: "kubectl -n {{ .Namespace }} get {{ .Kind }} {{ .Obj.metadata.name }}",
		},
		{
			Name:    "describe",
			Command: "kubectl -n {{ .Namespace }} describe {{ .Kind }} {{ .Obj.metadata.name }}",
		},
	}
	cfg := &Config{}
	if exists(confPath) {
		d, err := ioutil.ReadFile(confPath)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(d, cfg)
		if err != nil {
			return nil, err
		}
	}

	runner := &ActionRunner{
		Namespace:      namespace,
		ActionsMap:     cfg.ActionsMap,
		DefaultActions: defaultActions,
	}

	return runner, nil
}

type Config struct {
	ActionsMap map[string][]Action `yaml:"actions"`
}

type ActionRunner struct {
	Namespace      string
	ActionsMap     map[string][]Action
	DefaultActions []Action
}

type Action struct {
	Name      string     `yaml:"name"`
	Variables []Variable `yaml:"variables,omitempty"`
	Command   string     `yaml:"command"`
}

type Variable struct {
	Name     string `yaml:"name"`
	JSONPath string `yaml:jsonpath`
}

func (a *ActionRunner) Select(kind string) (*Action, error) {
	actions := a.ActionsMap[kind]
	actions = append(actions, a.DefaultActions...)

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
	}

	searcher := func(input string, index int) bool {
		action := actions[index]
		name := strings.Replace(strings.ToLower(action.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	actionPrompt := promptui.Select{
		Label:             "actions",
		Items:             actions,
		Templates:         templates,
		Searcher:          searcher,
		StartInSearchMode: true,
	}

	i, _, err := actionPrompt.Run()
	if err != nil {
		return nil, err
	}
	cmdTmpl := actions[i]

	return &cmdTmpl, nil
}

func (a *ActionRunner) GenerateCommand(obj runtime.Object, kind string, action *Action) ([]string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	d := map[string]interface{}{
		"Obj":       out,
		"Namespace": a.Namespace,
		"Kind":      kind,
	}

	for _, variable := range action.Variables {
		j := jsonpath.New(variable.Name)
		j.Parse(variable.JSONPath)
		tmp := new(bytes.Buffer)
		err = j.Execute(tmp, out)
		if err != nil {
			return nil, err
		}
		d[variable.Name] = tmp.String()
	}

	tmpl, err := template.New("command").Parse(action.Command)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, d)
	if err != nil {
		return nil, err
	}

	return strings.Split(buf.String(), " "), nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
