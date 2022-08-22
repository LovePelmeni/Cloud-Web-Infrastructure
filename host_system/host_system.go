package host_system

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/exp/slices"
)

// Package for Managing Host System of the Virtual Machine Server

type HostSystemCredentials struct {
	Bit        int64  `json:"Bit"`
	SystemName string `json:"SystemName"`
}

func NewHostSystemCredentials(SystemName string, Bit int64) *HostSystemCredentials {

	return &HostSystemCredentials{
		SystemName: strings.ToLower(SystemName),
		Bit:        Bit,
	}
}

type VirtualMachineHostSystemManager struct{}

func NewVirtualMachineHostSystemManager() *VirtualMachineHostSystemManager {
	return &VirtualMachineHostSystemManager{}
}
func (this *VirtualMachineHostSystemManager) GetHostSystemLocalPath(SystemName string) (string, error) {
	// Picking up default Local Path, depending on the Operational System
	if SystemName == "ubuntu" {
		return "", nil
	}
	if SystemName == "windows" {
		return "", nil
	}
	return "", errors.New("Invalid System Name")
}

func (this *VirtualMachineHostSystemManager) GetDefaultCustomizationOptions(SystemName string) (types.BaseCustomizationOptions, error) {
	// Returns Customization Options, based on the Operational System passed
	LinuxDistributions := []string{}
	WindowsDistrubitions := []string{}
	if Contains := slices.Contains(LinuxDistributions, strings.ToLower(SystemName)); Contains {
		return &types.CustomizationLinuxOptions{}, nil
	}
	if Contains := slices.Contains(WindowsDistrubitions, strings.ToLower(SystemName)); Contains {
		return &types.CustomizationWinOptions{}, nil
	}
	return nil, errors.New("Invalid Host System Name")
}

// Returns Default Operational System Options, depending on the System Name.

func (this *VirtualMachineHostSystemManager) SetupHostSystem(VirtualMachine *object.VirtualMachine, HostSystemCredentials HostSystemCredentials) (*types.VirtualMachineConfigSpec, error) {

	TimeoutContext, CancelFunc := context.WithTimeout(context.Background(), time.Minute*1)
	defer CancelFunc()

	DefaultCustomizationOptions, OptionsError := this.GetDefaultCustomizationOptions(HostSystemCredentials.SystemName)
	if OptionsError != nil {
		return nil, errors.New("Invalid Operational System Name")
	}

	VirtualMachineConfigSpecification := types.CustomizationSpec{
		Options:  *DefaultCustomizationOptions,
		Identity: &types.CustomizationIdentitySettings{},
	}

	V := types.VirtualMachineConfigSpec{
		ExtraConfig: []types.BaseOptionValue{&types.OptionValue{}},
		BootOptions: &types.VirtualMachineBootOptions{
			BootDelay:        10,
			BootRetryEnabled: types.NewBool(true),
			BootRetryDelay:   10,
		},
	}

	BootDevice, DeviceError := VirtualMachine.Device(TimeoutContext)
	if DeviceError != nil {
		ErrorLogger.Printf("Failed to Retrieve List of Boot Devices for the VM, Error: %s",
			DeviceError)
		return nil, errors.New("Failed to Setup HostSystem")
	}

	HostLocalFileSystemConfiguration := types.HostLocalFileSystemVolumeSpec{
		Device:    BootDevice.PrimaryMacAddress(),
		LocalPath: this.GetHostSystemLocalPath(HostSystemCredentials.SystemName),
	}
	HostSystemReconnectConfiguration := types.HostSystemReconnectSpec{
		SyncState: types.NewBool(true),
	}
	return VirtualMachineConfigSpecification, nil
}
