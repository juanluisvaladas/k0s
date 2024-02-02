/*
Copyright 2024 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package helm

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"

	"helm.sh/helm/v3/cmd/helm/app"

	_ "unsafe"
)

func NewHelmCmd() *cobra.Command {
	// This function is an adaptation of teh main function of helm in
	// github.com/helm/helm/cmd/helm/helm.go

	// Setting the name of the app for managedFields in the Kubernetes client.
	// It is set here to the full name of "helm" so that renaming of helm to
	// another name (e.g., helm2 or helm3) does not change the name of the
	// manager as picked up by the automated name detection.
	kube.ManagedFieldsManager = "helm"

	actionConfig := new(action.Configuration)
	args := extractHelmCommand(os.Args)
	fmt.Println(args)

	cmd, err := app.NewRootCmd(actionConfig, os.Stdout, args)
	fmt.Println()
	if err != nil {
		//warning("%+v", err)
		os.Exit(1)
	}

	return cmd

}

func extractHelmCommand(osArgs []string) []string {
	var args []string
	helmArgFound := false
	for _, arg := range osArgs {
		if arg == "helm" {
			helmArgFound = true
		}
		if helmArgFound {
			args = append(args, arg)
		}
	}
	return args
}
