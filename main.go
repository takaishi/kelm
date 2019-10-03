package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/takaishi/kelm/pkg/actions"
	"github.com/takaishi/kelm/pkg/k8s"
	"github.com/urfave/cli"
	"k8s.io/apimachinery/pkg/runtime"
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
	}

	app.Action = func(c *cli.Context) error {

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
			log.Printf("[DEBUG] namespace: %s\n", namespace)
			k8s.SetNamespace(namespace)
		} else {
			k8s.SetNamespace(namespace)
		}

		kind, err := k8s.SelectKind()
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] kind: %+v\n", kind)
		var obj runtime.Object
		switch kind {
		case "node":
			obj, err = k8s.SelectNode()
			if err != nil {
				return err
			}
		case "pods":
			log.Println("[DEBUG] select pod")
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
		runner, err := actions.NewActionRunner()
		if err != nil {
			return err
		}
		action, err := runner.Select(kind)
		if err != nil {
			return errors.Wrap(err, "failed to runner.Select()")
		}

		//kind := obj.GetObjectKind().GroupVersionKind().Kind
		cmd, err := action.GenerateCommand(obj, kind)
		if err != nil {
			return errors.Wrap(err, "failed to runner.GenerateCommand()")
		}
		log.Printf("[DEBUG] cmd: %+v\n", cmd)

		out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			return err
		}

		fmt.Println(string(out))
		return nil
	}

	return app.Run(os.Args)
}
