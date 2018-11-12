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
	"testing"

	"k8s.io/utils/exec"
	fakeexec "k8s.io/utils/exec/testing"

	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/debug"
)

const (
	adfCtlOutput = `Checking status of all devices.
There is 3 QAT acceleration device(s) in the system:
 qat_dev0 - type: c6xx,  inst_id: 0,  node_id: 0,  bsf: 3b:00.0,  #accel: 5 #engines: 10 state: up
 qat_dev1 - type: c6xx,  inst_id: 1,  node_id: 0,  bsf: 3d:00.0,  #accel: 5 #engines: 10 state: up
 qat_dev2 - type: c6xx,  inst_id: 2,  node_id: 3,  bsf: d8:00.0,  #accel: 5 #engines: 10 state: up
`
)

func init() {
	debug.Activate()
}

func TestParseConfigs(t *testing.T) {
	fcmd := fakeexec.FakeCmd{
		CombinedOutputScript: []fakeexec.FakeCombinedOutputAction{
			func() ([]byte, error) {
				return []byte(adfCtlOutput), nil
			},
		},
	}
	execer := fakeexec.FakeExec{
		CommandScript: []fakeexec.FakeCommandAction{
			func(cmd string, args ...string) exec.Cmd {
				return fakeexec.InitFakeCmd(&fcmd, cmd, args...)
			},
		},
	}
	tcases := []struct {
		name        string
		testData    string
		expectedErr bool
	}{
		{
			name:     "All is goog",
			testData: "all_is_good",
		},
	}
	for _, tc := range tcases {
		dp := &devicePlugin{
			execer:    &execer,
			configDir: "./test_data/" + tc.testData,
		}
		_, err := dp.parseConfigs()
		if tc.expectedErr && err == nil {
			t.Errorf("Test case '%s': expected error hasn't been triggered", tc.name)
		}
		if !tc.expectedErr && err != nil {
			t.Errorf("Test case '%s': Unexpcted error: %+v", tc.name, err)
		}
	}
}
