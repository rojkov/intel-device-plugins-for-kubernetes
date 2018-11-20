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
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Thomasdezeeuw/ini"
	"github.com/pkg/errors"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	utilsexec "k8s.io/utils/exec"

	dpapi "github.com/intel/intel-device-plugins-for-kubernetes/internal/deviceplugin"
	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/debug"
)

const (
	namespace = "qat.intel.com"
)

var (
	uioRegex = regexp.MustCompile(`^uio[0-9]+$`)
)

type endpoint struct {
	id        string
	processes int
}

type section struct {
	endpoints          []endpoint
	cryptoEngines      int
	compressionEngines int
	pinned             bool
}

func getDevTree(config map[string]section) dpapi.DeviceTree {
	devTree := dpapi.NewDeviceTree()

	devFiles, err := ioutil.ReadDir("/dev")
	if err != nil {
		debug.Print("Can't read /dev. Do nothing.")
		return devTree
	}

	devs := []string{
		"/dev/qat_adf_ctl",
		"/dev/qat_dev_processes",
		"/dev/usdm_drv",
	}
	for _, devFile := range devFiles {
		fname := devFile.Name()

		if uioRegex.MatchString(fname) {
			devs = append(devs, path.Join("/dev", fname))
		}
	}

	for sname, svalue := range config {
		var devType string

		devType = fmt.Sprintf("cy%d_dc%d", svalue.cryptoEngines, svalue.compressionEngines)
		for k, ep := range svalue.endpoints {
			for i := 0; i < ep.processes; i++ {
				devTree.AddDevice(devType, fmt.Sprintf("%s_%s_%d", sname, ep.id, i), dpapi.DeviceInfo{
					State: pluginapi.Healthy,
					Nodes: devs,
					Envs: map[string]string{
						fmt.Sprintf("QAT_SECTION_NAME_%s_%d", devType, i*(k+1)): sname,
						// This env variable may get overriden if a container requests more than one QAT process.
						// But we keep this code since the majority of pod workloads run only one QAT process.
						// The rest should use QAT_SECTION_NAME_XXX variables.
						"QAT_SECTION_NAME": sname,
					},
				})
			}

			if !svalue.pinned {
				break
			}
		}
	}

	return devTree
}

type devicePlugin struct {
	execer    utilsexec.Interface
	configDir string
}

func newDevicePlugin(configDir string, execer utilsexec.Interface) *devicePlugin {
	return &devicePlugin{
		execer:    execer,
		configDir: configDir,
	}
}

func (dp *devicePlugin) parseConfigs() (map[string]section, error) {
	outputBytes, err := dp.execer.Command("adf_ctl", "status").CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "Can't get driver status")
	}
	output := string(outputBytes[:])

	devNum := 0
	driverConfig := make(map[string]section)
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

		f, err := os.Open(path.Join(dp.configDir, fmt.Sprintf("%s_%s.conf", devType, devID)))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		defer f.Close()

		// Parse the configuration.
		config, err := ini.Parse(f)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		devNum++

		debug.Print(ln, devID, line)

		for sectionName, data := range config {
			if sectionName == "GENERAL" || sectionName == "KERNEL" || sectionName == "" {
				continue
			}
			debug.Print(sectionName)

			numProcesses, err := strconv.Atoi(data["NumProcesses"])
			if err != nil {
				return nil, errors.Wrapf(err, "Can't convert NumProcesses in %s", sectionName)
			}
			cryptoEngines, err := strconv.Atoi(data["NumberCyInstances"])
			if err != nil {
				return nil, errors.Wrapf(err, "Can't convert NumberCyInstances in %s", sectionName)
			}
			compressionEngines, err := strconv.Atoi(data["NumberDcInstances"])
			if err != nil {
				return nil, errors.Wrapf(err, "Can't convert NumberDcInstances in %s", sectionName)
			}
			pinned := false
			if limitDevAccess, ok := data["LimitDevAccess"]; ok {
				if limitDevAccess != "0" {
					pinned = true
				}
			}

			if old, ok := driverConfig[sectionName]; ok {
				// first check the sections are consistent across endpoints
				if old.pinned != pinned {
					return nil, errors.Errorf("Value of LimitDevAccess must be consistent across all devices in %s", sectionName)
				}
				if !pinned && old.endpoints[0].processes != numProcesses {
					return nil, errors.Errorf("For not pinned section \"%s\" NumProcesses must be equal for all devices", sectionName)
				}
				if old.cryptoEngines != cryptoEngines || old.compressionEngines != compressionEngines {
					return nil, errors.Errorf("NumberCyInstances and NumberDcInstances must be consistent across all devices in %s", sectionName)
				}

				// then add a new endpoint to the section
				old.endpoints = append(old.endpoints, endpoint{
					id:        devID,
					processes: numProcesses,
				})
				driverConfig[sectionName] = old
			} else {
				driverConfig[sectionName] = section{
					endpoints: []endpoint{
						{
							id:        devID,
							processes: numProcesses,
						},
					},
					cryptoEngines:      cryptoEngines,
					compressionEngines: compressionEngines,
					pinned:             pinned,
				}
			}
		}

	}

	// check if the number of sections with LimitDevAccess=1 is equal to the number of endpoints
	for sname, svalue := range driverConfig {
		if svalue.pinned && len(svalue.endpoints) != devNum {
			return nil, errors.Errorf("Section [%s] must be defined for all QAT devices since it contains LimitDevAccess=1", sname)
		}
	}

	return driverConfig, nil
}

func (dp *devicePlugin) Scan(notifier dpapi.Notifier) error {
	for {
		driverConfig, err := dp.parseConfigs()
		if err != nil {
			return err
		}

		notifier.Notify(getDevTree(driverConfig))

		time.Sleep(5 * time.Second)
	}
}

func (dp *devicePlugin) PostAllocate(response *pluginapi.AllocateResponse) error {
	for _, containerResponse := range response.GetContainerResponses() {
	}

	return nil
}

func main() {
	debugEnabled := flag.Bool("debug", false, "enable debug output")
	flag.Parse()

	if *debugEnabled {
		debug.Activate()
	}

	plugin := newDevicePlugin("/etc", utilsexec.New())

	manager := dpapi.NewManager(namespace, plugin)
	manager.Run()
}
