package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/takaishi/ik/pkg/actions"
	"github.com/takaishi/ik/pkg/k8s"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"os/exec"
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

	kind, err := k8s.SelectKind()
	if err != nil {
		return err
	}

	var obj runtime.Object
	switch kind {
	case "node":
		obj, err = k8s.SelectNode()
		if err != nil {
			return err
		}
	case "pod":
		obj, err = k8s.SelectPod()
		if err != nil {
			return err
		}
	case "crd":
		obj, err = k8s.SelectCRD()
		if err != nil {
			return err
		}
	}

	actions, err := actions.New()
	if err != nil {
		return err
	}
	action, err := actions[kind].Select()
	if err != nil {
		return errors.Wrap(err, "failed to actions.Select()")
	}

	cmd, err := action.GenerateCommand(obj)
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
