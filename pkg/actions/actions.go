package actions

import (
	"bytes"
	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"text/template"
)

func New() (map[string]*Actions, error) {
	d, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(d, cfg)
	if err != nil {
		return nil, err
	}

	return cfg.ActionsMap, nil
}

type Config struct {
	ActionsMap map[string]*Actions `yaml:"actions"`
}

type Actions []Action

type Action struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

func (a *Actions) Select() (*Action, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
	}

	searcher := func(input string, index int) bool {
		action := (*a)[index]
		name := strings.Replace(strings.ToLower(action.Name), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input)
	}

	actionPrompt := promptui.Select{
		Label:             "actions",
		Items:             *a,
		Templates:         templates,
		Searcher:          searcher,
		StartInSearchMode: true,
	}

	i, _, err := actionPrompt.Run()
	if err != nil {
		return nil, err
	}
	cmdTmpl := (*a)[i]

	return &cmdTmpl, nil
}

func (a *Action) GenerateCommand(data interface{}) ([]string, error) {
	tmpl, err := template.New("command").Parse(a.Command)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, err
	}

	return strings.Split(buf.String(), " "), nil
}
