// Copyright 2018 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/coreos/prometheus-operator/pkg/prometheus"

	"k8s.io/api/core/v1"
	k8sYAML "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/ghodss/yaml"
)

func main() {
	var ruleConfigMapName = flag.String("rule-config-map", "", "path to rule config map")
	var ruleCRDSDestination = flag.String("rule-crds-destination", "", "destination new crds should be created in")
	flag.Parse()

	if *ruleConfigMapName == "" {
		log.Print("please specify 'rule-config-map' flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *ruleCRDSDestination == "" {
		log.Print("please specify 'rule-crds-destination' flag")
		flag.PrintDefaults()
		os.Exit(1)
	}

	destPath, err := filepath.Abs(*ruleCRDSDestination)
	if err != nil {
		log.Fatalf("failed to get absolut path of '%v': %v", ruleCRDSDestination, err.Error())
	}
	ruleCRDSDestination = &destPath

	file, err := os.Open(*ruleConfigMapName)
	if err != nil {
		log.Fatalf("failed to read file '%v': %v", ruleConfigMapName, err.Error())
	}

	configMap := v1.ConfigMap{}

	err = k8sYAML.NewYAMLOrJSONDecoder(file, 100).Decode(&configMap)
	if err != nil {
		log.Fatalf("failed to decode manifest: %v", err.Error())
	}

	ruleFiles, err := prometheus.CMToRuleFiles(&configMap)
	if err != nil {
		log.Fatalf("failed to transform config map to rule file crds: %v", err.Error())
	}

	for _, ruleFile := range ruleFiles {
		encodedRuleFile, err := yaml.Marshal(ruleFile)
		if err != nil {
			log.Fatalf("failed to encode ruleFile '%v': %v", ruleFile.Name, err.Error())
		}

		err = ioutil.WriteFile(path.Join(*ruleCRDSDestination, ruleFile.Name), encodedRuleFile, 0644)
		if err != nil {
			log.Fatalf("failed to write yaml manifest for rule file '%v': %v", ruleFile.Name, err.Error())
		}
	}
}
