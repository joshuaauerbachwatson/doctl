/*
Copyright 2021 The Doctl Authors All rights reserved.
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

// TODO perhaps this doesn't belong in commands
package commands

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/digitalocean/godo"
)

// Highly experimental hack to use 'nim' in conjunction with doctl.

// Top level function to deploy serverless actions in the form of Nimbella projects
func deployServerless(projects []*godo.AppServerlessSpec) (string, error) {
	var nimProjects = []string{}
	for _, project := range projects {
		nimProject, err := convertToNimProject(project)
		if err != nil {
			return "", err
		}
		nimProjects = append(nimProjects, nimProject)
	}
	args := append([]string{"project", "deploy", "--exclude", "web"}, nimProjects...)
	return runNim(args)
}

// Function to convert a godo.AppServerlessSpec to a concrete project location that 'nim' can deploy
// (either GitHub resident or local)
func convertToNimProject(spec *godo.AppServerlessSpec) (string, error) {
	if spec.Local == nil {
		if spec.GitHub == nil {
			return "", errors.New("one of `Local` or `GitHub` must appear in a `serverless` spec")
		} else {
			return githubProject(spec.GitHub, spec.SourceDir)
		}
	} else if spec.GitHub != nil {
		return "", errors.New("you cannot specify both `Local` and `GitHub` in a `serverless` spec")
	} else {
		return localProject(spec.Local, spec.SourceDir)
	}
}

// Function to convert godo.GitHubSourceSpec + source directory path to an appropriate project argument
// for 'nim project deploy'
func githubProject(spec *godo.GitHubSourceSpec, sourceDir string) (string, error) {
	if spec.DeployOnPush {
		return "", errors.New("the `deploy on push` feature is not currently supported for serverless")
	}
	if spec.Repo == "" {
		return "", errors.New("The `repo` field is required")
	}
	project := "github:" + spec.Repo
	if sourceDir != "" {
		project = project + "/" + sourceDir
	}
	if spec.Branch != "" {
		project = project + "#" + spec.Branch
	}
	return project, nil
}

// Function to convert godo.LocalSourceSpec + source directory path to an appropriate project argument
// for 'nim project deploy'
func localProject(spec *godo.LocalSourceSpec, sourceDir string) (string, error) {
	if spec.Path != "" {
		if sourceDir != "" {
			return spec.Path + "/" + sourceDir, nil
		} else {
			return spec.Path, nil
		}
	} else if sourceDir != "" {
		return sourceDir, nil
	} else {
		return "", errors.New("If `local` is used, either the path or the sourceDir or both must be specified")
	}
}

// For use with doctl, we currently assume a special installation of 'nim' in ~/.nimbella/cli.
// The install procedure for this is unclear in the long run.  Basically an install tarball for
// nimbella-cli needs to be unpacked there and then a 'nim login' run to set minimal credentials
// and establish a current namespace.
// This has the advantage of being findable on all OSs without having to be in $PATH. Among many
// plausible alternatives is to keep it in any directory that doctl uses for other purposes.  I'm
// currently less sure how to manage that across supported OSs so I'm using ~/.nimbella, which
// currently has to exist for any user logged into a Nimbella stack.
func getNimPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(homeDir, ".nimbella", "cli", "bin", "nim")
	return path, nil
}

// Function to run any 'nim' command.  Currently used by deployServerless but can be used to expose parts
// of the 'nim' command tree as subcommands of doctl.
func runNim(args []string) (string, error) {
	nim, err := getNimPath()
	if err != nil {
		return "", err
	}
	output, err := exec.Command(nim, args...).CombinedOutput()
	return string(output), err
}
