// Copyright 2019 Intel Corporation. All Rights Reserved.
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
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/fpga"
	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/fpga/bitstream"
)

const (
	fpgaBitStreamDirectory = "/srv/intel.com/fpga"
)

func main() {
	var err error
	var bitstream, device string
	var dryRun, force, quiet bool
	flag.StringVar(&bitstream, "b", "", "Path to bitstream file (GBS or AOCX)")
	flag.StringVar(&device, "d", "", "Path to device node (FME or Port)")
	flag.BoolVar(&dryRun, "dry-run", false, "Don't write/program, just validate and log")
	flag.BoolVar(&force, "force", false, "Force overwrite operation for installing bitstreams")
	flag.BoolVar(&quiet, "q", false, "Quiet mode. Only errors will be reported")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("Please provide command: info, fpgainfo, install, list, fmeinfo, portinfo, list-fme, list-port, pr")
	}

	cmd := flag.Arg(0)
	err = validateFlags(cmd, bitstream, device)
	if err != nil {
		log.Fatalf("Invalid arguments: %+v", err)
	}

	switch cmd {
	case "info":
		err = printBitstreamInfo(bitstream, quiet)
	case "pr":
		err = doPR(device, bitstream, dryRun, quiet)
	case "fpgainfo":
		err = fpgaInfo(device, quiet)
	case "fmeinfo":
		err = fmeInfo(device, quiet)
	case "portinfo":
		err = portInfo(device, quiet)
	case "install":
		err = installBitstream(bitstream, dryRun, force, quiet)
	case "list":
		err = listDevices(true, true, quiet)
	case "list-fme":
		err = listDevices(true, false, quiet)
	case "list-port":
		err = listDevices(false, true, quiet)
	default:
		err = errors.Errorf("unknown command %+v", flag.Args())
	}
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func validateFlags(cmd, bitstream, device string) error {
	switch cmd {
	case "info", "install":
		// bitstream must not be empty
		if bitstream == "" {
			return errors.Errorf("bitstream filename is missing")
		}
	case "fpgainfo", "fmeinfo", "portinfo":
		// device must not be empty
		if device == "" {
			return errors.Errorf("FPGA device name is missing")
		}
	case "pr":
		// device and bitstream can't be empty
		if bitstream == "" {
			return errors.Errorf("bitstream filename is missing")
		}
		if device == "" {
			return errors.Errorf("FPGA device name is missing")
		}
	}
	return nil
}

func installBitstream(fname string, dryRun, force, quiet bool) (err error) {
	info, err := bitstream.Open(fname)
	if err != nil {
		return
	}
	defer info.Close()

	installPath := info.InstallPath(fpgaBitStreamDirectory)

	if !quiet {
		fmt.Printf("Installing bitstream %q as %q\n", fname, installPath)
		if dryRun {
			fmt.Println("Dry-run: no copying performed")
			return
		}
	}
	err = os.MkdirAll(filepath.Dir(installPath), 0755)
	if err != nil {
		return errors.Wrap(err, "unable to create destination directory")
	}
	src, err := os.Open(fname)
	if err != nil {
		return errors.Wrap(err, "can't open bitstream file")
	}
	defer src.Close()
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if !force {
		flags = flags | os.O_EXCL
	}
	dst, err := os.OpenFile(installPath, flags, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("destination file %q already exist. Use --force to overwrite it", installPath)
		}
		return errors.Wrap(err, "can't create destination file")
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	return err
}

func printBitstreamInfo(fname string, quiet bool) (err error) {
	info, err := bitstream.Open(fname)
	if err != nil {
		return
	}
	defer info.Close()
	fmt.Printf("Bitstream file        : %q\n", fname)
	fmt.Printf("Interface UUID        : %q\n", info.InterfaceUUID())
	fmt.Printf("Accelerator Type UUID : %q\n", info.AcceleratorTypeUUID())
	fmt.Printf("Unique UUID           : %q\n", info.UniqueUUID())
	fmt.Printf("Installation Path     : %q\n", info.InstallPath(fpgaBitStreamDirectory))
	extra := info.ExtraMetadata()
	if len(extra) > 0 && !quiet {
		fmt.Println("Extra:")
		for k, v := range extra {
			fmt.Printf("\t%s : %q\n", k, v)
		}
	}
	return
}

func fpgaInfo(fname string, quiet bool) error {
	switch {
	case fpga.IsFpgaFME(fname):
		return fmeInfo(fname, quiet)
	case fpga.IsFpgaPort(fname):
		return portInfo(fname, quiet)
	}
	return errors.Errorf("unknown FPGA device file %s", fname)
}

func fmeInfo(fname string, quiet bool) error {
	var f fpga.FME
	var err error
	f, err = fpga.NewFME(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return printFpgaFME(f, quiet)
}

func printFpgaFME(f fpga.FME, quiet bool) (err error) {
	fmt.Println("//****** FME ******//")
	fmt.Printf("Name                             : %s\n", f.GetName())
	fmt.Printf("Device Node                      : %s\n", f.GetDevPath())
	fmt.Printf("SysFS Path                       : %s\n", f.GetSysFsPath())
	pci, err := f.GetPCIDevice()
	if err != nil {
		return
	}
	printPCIeInfo(pci, quiet)
	fmt.Printf("Interface UUID                   : %s\n", f.GetInterfaceUUID())
	if !quiet {
		if apiVer, err := f.GetAPIVersion(); err == nil {
			fmt.Printf("Kernet API Version               : %d\n", apiVer)
		}
		fmt.Printf("Ports Num                        : %d\n", f.GetPortsNum())
		if id, err := f.GetSocketID(); err == nil {
			fmt.Printf("Socket Id                        : %d\n", id)
		}
		fmt.Printf("Bitstream Id                     : %s\n", f.GetBitstreamID())
		fmt.Printf("Bitstream Metadata               : %s\n", f.GetBitstreamMetadata())
	}

	return
}

func portInfo(fname string, quiet bool) error {
	var f fpga.Port
	var err error
	f, err = fpga.NewPort(fname)
	if err != nil {
		return err
	}
	defer f.Close()
	return printFpgaPort(f, quiet)
}

func printFpgaPort(f fpga.Port, quiet bool) (err error) {
	fmt.Println("//****** PORT ******//")
	fmt.Printf("Name                             : %s\n", f.GetName())
	fmt.Printf("Device Node                      : %s\n", f.GetDevPath())
	fmt.Printf("SysFS Path                       : %s\n", f.GetSysFsPath())
	pci, err := f.GetPCIDevice()
	if err != nil {
		return
	}
	printPCIeInfo(pci, quiet)
	fme, err := f.GetFME()
	if err != nil {
		return
	}
	fmt.Printf("FME Name                         : %s\n", fme.GetName())
	num, err := f.GetPortID()
	if err != nil {
		return
	}
	fmt.Printf("Port Id                          : %d\n", num)
	fmt.Printf("Interface UUID                   : %s\n", f.GetInterfaceUUID())
	fmt.Printf("Accelerator UUID                 : %s\n", f.GetAcceleratorTypeUUID())
	if !quiet {
		if apiVer, err2 := f.GetAPIVersion(); err2 == nil {
			fmt.Printf("Kernet API Version               : %d\n", apiVer)
			pi, err2 := f.PortGetInfo()
			if err2 == nil {
				fmt.Printf("Port Regions                     : %d\n", pi.Regions)
				for idx := 0; uint32(idx) < pi.Regions; idx++ {
					if ri, err3 := f.PortGetRegionInfo(uint32(idx)); err3 == nil {
						fmt.Printf("Port Region (Index/Size/Offset)  : %d / %d / %d\n", ri.Index, ri.Size, ri.Offset)
					}
				}
			}
		}
	}
	return err
}

func printPCIeInfo(pci *fpga.PCIDevice, quiet bool) {
	fmt.Printf("PCIe s:b:d:f                     : %s\n", pci.BDF)
	if pci.PhysFn != nil && !quiet {
		fmt.Printf("Physical Function PCIe s:b:d:f   : %s\n", pci.PhysFn.BDF)
	}
	fmt.Printf("Device Id                        : %s:%s\n", pci.Vendor, pci.Device)
	if !quiet {
		fmt.Printf("Device Class                     : %s\n", pci.Class)
		fmt.Printf("Local CPUs                       : %s\n", pci.CPUs)
		fmt.Printf("NUMA                             : %s\n", pci.NUMA)
		if pci.VFs != "" {
			fmt.Printf("SR-IOV Virtual Functions         : %s\n", pci.VFs)
		}
		if pci.TotalVFs != "" {
			fmt.Printf("SR-IOV maximum Virtual Functions : %s\n", pci.TotalVFs)
		}
	}
}

func doPR(dev, fname string, dryRun, quiet bool) (err error) {
	fp, err := fpga.NewPort(dev)
	if err != nil {
		return
	}
	defer fp.Close()
	bs, err := bitstream.Open(fname)
	if err != nil {
		return
	}
	defer bs.Close()

	if !quiet {
		fmt.Printf("Before: Interface ID: %q AFU ID: %q\n", fp.GetInterfaceUUID(), fp.GetAcceleratorTypeUUID())
		fmt.Printf("Programming %q to port %q: ", fname, dev)
	}
	err = fp.PR(bs, dryRun)
	if !quiet {
		if err != nil {
			fmt.Println("FAILED")
		} else {
			fmt.Println("OK")
		}
		fmt.Printf("After : Interface ID: %q AFU ID: %q\n", fp.GetInterfaceUUID(), fp.GetAcceleratorTypeUUID())
	}
	return
}

func listDevices(listFMEs, listPorts, quiet bool) error {
	fmes, ports := fpga.ListFpgaDevices()
	if listFMEs {
		if !quiet {
			fmt.Printf("Detected FPGA FMEs: %d\n", len(fmes))
		}
		for _, v := range fmes {
			fmt.Println(v)
		}
	}
	if listPorts {
		if !quiet {
			fmt.Printf("Detected FPGA Ports: %d\n", len(ports))
		}
		for _, v := range ports {
			fmt.Println(v)
		}
	}
	return nil
}
