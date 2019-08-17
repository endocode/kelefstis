package main

import (
	"bytes"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecuteRemoteCommand executes a remote shell command on the given pod
// and returns the output from stdout and stderr
func ExecuteRemoteCommand(namespace, name, container, command string) (string, string, error) {

	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	client, err := rest.RESTClientFor(cfg)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed creating RESTClientfor %v command %s on %s/%s",
			cfg, command, namespace, name)
	}
	glog.Infof("execing into %s/%s %s", namespace, name, command)
	request := client.
		Post().
		Namespace(namespace).
		Resource("pods").
		Name(name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Command:   []string{"/bin/sh", "-c", command},
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
			Container: container,
		}, scheme.ParameterCodec)
	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", request.URL())
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: outBuf,
		Stderr: errBuf,
	})

	if err != nil {
		return "", "", errors.Wrapf(err, "failed executing command %s on %s/%s container %s", command, namespace, name, container)
	}

	return outBuf.String(), errBuf.String(), nil
}
