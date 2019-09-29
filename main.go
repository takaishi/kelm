package main

import (
	"bytes"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
	"github.com/takaishi/ik/pkg/actions"
	"github.com/takaishi/ik/pkg/k8s"
	"html/template"
	"log"
	"os/exec"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run() error {
	k8s, err := k8s.New()
	if err != nil {
		return err
	}

	pod, err := k8s.SelectPod()
	if err != nil {
		return err
	}

	actions, err := actions.New()
	if err != nil {
		return err
	}
	action, err := actions.Select()
	if err != nil {
		return errors.Wrap(err, "failed to actions.Select()")
	}

	cmd, err := action.GenerateCommand(pod)
	if err != nil {
		return errors.Wrap(err, "failed to actions.GenerateCommand()")
	}

	out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

type Config struct {
	Actions Actions `yaml:"actions"`
}

type Actions []Action

type Action struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

func selectActions(actions Actions) (*Action, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "> {{ .Name | cyan }}",
		Inactive: "  {{ .Name | cyan }}",
		Selected: "  {{ .Name | red | cyan }}",
	}

	searcher := func(input string, index int) bool {
		pepper := actions[index]
		name := strings.Replace(strings.ToLower(pepper.Name), " ", "", -1)
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

func generateCommand(tmplStr string, data interface{}) ([]string, error) {
	tmpl, err := template.New("command").Parse(tmplStr)
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
