// Copyright 2018 Intel Corporation. All Rights Reserved.
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
	"fmt"
	"os"
	"strings"

	"github.com/Thomasdezeeuw/ini"
	utilsexec "k8s.io/utils/exec"

	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/debug"
)

func main() {
	debugEnabled := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *debugEnabled {
		debug.Activate()
	}

	execer := utilsexec.New()

	outputBytes, err := execer.Command("adf_ctl", "status").CombinedOutput()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
	output := string(outputBytes[:])

	for ln, line := range strings.Split(output, "\n") {
		if !strings.HasPrefix(line, " qat_") {
			continue
		}

		devstr := strings.SplitN(line, "-", 2)
		if len(devstr) != 2 {
			continue
		}

		devprops := strings.Split(devstr[1], ",")
		devType := ""
		for _, propstr := range devprops {
			if strings.TrimSpace(propstr) == "type: c6xx" {
				devType = "c6xx"
			}
		}

		if devType == "" {
			continue
		}

		devID := strings.TrimPrefix(strings.TrimSpace(devstr[0]), "qat_")

		f, err := os.Open(fmt.Sprintf("/etc/%s_%s.conf", devType, devID))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		// Parse the configuration.
		config, err := ini.Parse(f)
		if err != nil {
			panic(err)
		}

		fmt.Println(ln, devID, line)

		for k, v := range config["SHIM"] {
			fmt.Println(k, v)
		}

	}
}
