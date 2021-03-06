/*
Copyright 2021 Flant CJSC

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

package modules

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/deckhouse/deckhouse/testing/matrix/linter/rules/errors"
)

func skipModuleImageNameIfNeeded(filePath string) bool {
	return filePath == "/deckhouse/modules/040-control-plane-manager/images/kube-apiserver/werf.inc.yaml"
}

var regexPatterns = map[string]string{
	`$BASE_ALPINE`:           imageRegexp(`alpine:[\d.]+`),
	`$BASE_DEBIAN`:           imageRegexp(`debian:[\d.]+`),
	`$BASE_GOLANG_ALPINE`:    imageRegexp(`golang:1.15.[\d.]+-alpine3.12`),
	`$BASE_GOLANG_16_ALPINE`: imageRegexp(`golang:1.16.[\d.]+-alpine3.12`),
	`$BASE_GOLANG_BUSTER`:    imageRegexp(`golang:1.15.[\d.]+-buster`),
	`$BASE_GOLANG_16_BUSTER`: imageRegexp(`golang:1.16.[\d.]+-buster`),
	`$BASE_NGINX_ALPINE`:     imageRegexp(`nginx:[\d.]+-alpine`),
	`$BASE_PYTHON_ALPINE`:    imageRegexp(`python:[\d.]+-alpine`),
	`$BASE_SHELL_OPERATOR`:   imageRegexp(`shell-operator:v[\d.]+`),
	`$BASE_UBUNTU`:           imageRegexp(`ubuntu:[\d.]+`),
}

func imageRegexp(s string) string {
	return fmt.Sprintf("^(from:|FROM)(\\s+)(%s)", s)
}

func isImageNameUnacceptable(imageName string) (bool, string) {
	for ciVariable, pattern := range regexPatterns {
		matched, _ := regexp.MatchString(pattern, imageName)
		if matched {
			return true, ciVariable
		}
	}
	return false, ""
}

func checkImageNamesInDockerAndWerfFiles(lintRuleErrorsList *errors.LintRuleErrorsList, name, path string) {
	var filePaths []string
	imagesPath := filepath.Join(path, imagesDir)

	if !isExistsOnFilesystem(imagesPath) {
		return
	}

	err := filepath.Walk(imagesPath, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch filepath.Base(fullPath) {
		case "werf.inc.yaml",
			"Dockerfile":
			filePaths = append(filePaths, fullPath)
		}
		return nil
	})

	if err != nil {
		lintRuleErrorsList.Add(errors.NewLintRuleError(
			"MODULE001",
			moduleLabel(name),
			imagesPath,
			"Cannot read directory structure:%s",
			err,
		))
		return
	}
	for _, filePath := range filePaths {
		if skipModuleImageNameIfNeeded(filePath) {
			continue
		}
		lintRuleErrorsList.Add(lintOneDockerfileOrWerfYAML(name, filePath, imagesPath))
	}
}

func lintOneDockerfileOrWerfYAML(name, filePath, imagesPath string) errors.LintRuleError {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.NewLintRuleError(
			"MODULE001",
			moduleLabel(name),
			filePath,
			"Error opening file:%s",
			err,
		)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	linePos := 0
	relativeFilePath, err := filepath.Rel(imagesPath, filePath)
	if err != nil {
		return errors.NewLintRuleError(
			"MODULE001",
			moduleLabel(name),
			filePath,
			"Error calculating relative file path:%s",
			err,
		)
	}

	for scanner.Scan() {
		line := scanner.Text()
		linePos++
		result, ciVariable := isImageNameUnacceptable(line)
		if result {
			return errors.NewLintRuleError(
				"MODULE001",
				fmt.Sprintf("module = %s, image = %s, line = %d", name, relativeFilePath, linePos),
				line,
				"Please use %s as an image name", ciVariable,
			)
		}
	}

	return errors.EmptyRuleError
}
