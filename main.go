package main

import (
	"github.com/pkg/errors"
	"github.com/takaishi/kelm/pkg/actions"
	"github.com/takaishi/kelm/pkg/k8s"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("%+v", err)
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows

}

func run() error {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "kubeconfig",
			Value: filepath.Join(homeDir(), ".kube", "config"),
		},
		cli.StringFlag{
			Name: "namespace, n",
		},
		cli.StringFlag{
			Name: "kind, k",
		},
		cli.StringFlag{
			Name: "action, a",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: filepath.Join(homeDir(), ".kelm"),
		},
	}

	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	return app.Run(os.Args)
}

func action(c *cli.Context) error {
	k8s, err := k8s.New(c.String("kubeconfig"))
	if err != nil {
		return err
	}

	namespace := c.String("namespace")
	if namespace == "" {
		namespace, err := k8s.SelectNamespace()
		if err != nil {
			return err
		}
		k8s.SetNamespace(namespace)
	} else {
		k8s.SetNamespace(namespace)
	}

	kind := c.String("kind")
	if kind == "" {
		kind, err = k8s.SelectKind()
		if err != nil {
			return err
		}
		k8s.SetKind(kind)
	} else {
		k8s.SetKind(kind)
	}

	obj, err := k8s.SelectObjects(kind)
	if err != nil {
		return err
	}

	runner, err := actions.NewActionRunner(k8s.GetNamespace(), c.String("config"))
	if err != nil {
		return err
	}

	actionString := c.String("action")
	action, err := runner.Select(k8s.GetKind(), actionString)
	if err != nil {
		return errors.Wrap(err, "failed to runner.Select()")
	}

	cmdText, err := runner.GenerateCommand(*obj, kind, action)
	if err != nil {
		return errors.Wrap(err, "failed to runner.GenerateCommand()")
	}

	cmd := exec.Command(cmdText[0], cmdText[1:]...)
	if err != nil {
		return err
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
