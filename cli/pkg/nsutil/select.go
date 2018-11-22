package nsutil

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"gopkg.in/AlecAivazis/survey.v1"
)

func ChooseMesh(nsr options.NsResourceMap) (string, string, error) {
	meshOptions, meshMap := generateMeshSelectOptions(nsr)
	if len(meshOptions) == 0 {
		return "", "", fmt.Errorf("No meshs found. Please create a mesh")
	}
	question := &survey.Select{
		Message: "Select a mesh",
		Options: meshOptions,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", "", err
	}
	return meshMap[choice].name, meshMap[choice].namespace, nil
}

func ChooseSecret(nsr options.NsResourceMap) (string, string, error) {
	secretOptions, secretMap := generateSecretSelectOptions(nsr)
	if len(secretOptions) == 0 {
		return "", "", fmt.Errorf("No secrets found. Please create a secret")
	}
	question := &survey.Select{
		Message: "Select a secret",
		Options: secretOptions,
	}
	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", "", err
	}
	return secretMap[choice].name, secretMap[choice].namespace, nil
}
