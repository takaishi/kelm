package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/takaishi/ik/pkg/actions"
	"github.com/takaishi/ik/pkg/k8s"
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
