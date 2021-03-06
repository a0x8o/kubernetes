/*
Copyright 2017 The Kubernetes Authors.

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

package deviceplugin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
	"k8s.io/kubernetes/pkg/kubelet/lifecycle"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// ActivePodsFunc is a function that returns a list of pods to reconcile.
type ActivePodsFunc func() []*v1.Pod

// Handler defines the functions used to manage and access device plugin resources.
type Handler interface {
	// Start starts device plugin registration service.
	Start(activePods ActivePodsFunc) error
	// Devices returns all of registered devices keyed by resourceName.
	Devices() map[string][]pluginapi.Device
	// Allocate scans through containers in the pod spec
	// If it finds the container requires device plugin resource, it:
	// 1. Checks whether it already has this information in its cached state.
	// 2. If not, it calls Allocate and populate its cached state afterwards.
	// 3. If there is no cached state and Allocate fails, it returns an error.
	// 4. Otherwise, it updates allocatableResource in nodeInfo if necessary,
	// to make sure it is at least equal to the pod's requested capacity for
	// any registered device plugin resource
	Allocate(node *schedulercache.NodeInfo, attrs *lifecycle.PodAdmitAttributes) error
	// GetDeviceRunContainerOptions checks whether we have cached containerDevices
	// for the passed-in <pod, container> and returns its DeviceRunContainerOptions
	// for the found one. An empty struct is returned in case no cached state is found.
	GetDeviceRunContainerOptions(pod *v1.Pod, container *v1.Container) *DeviceRunContainerOptions
}

// HandlerImpl implements the actual functionality to manage device plugin resources.
type HandlerImpl struct {
	// TODO: consider to change this to RWMutex.
	sync.Mutex
	// devicePluginManager is an implementation of deviceplugin.Manager interface.
	devicePluginManager Manager
	// activePods is a method for listing active pods on the node
	// so the amount of pluginResources requested by existing pods
	// could be counted when updating allocated devices
	activePods ActivePodsFunc
	// devicePluginManagerMonitorCallback is used for updating devices' states in one time call.
	// e.g. a new device is advertised, two old devices are deleted and a running device fails.
	devicePluginManagerMonitorCallback MonitorCallback
	// allDevices contains all of registered resourceNames and their exported device IDs.
	allDevices map[string]sets.String
	// allocatedDevices contains allocated deviceIds, keyed by resourceName.
	allocatedDevices map[string]sets.String
	// podDevices contains pod to allocated device mapping.
	podDevices podDevices
}

// NewHandlerImpl creates a HandlerImpl to manage device plugin resources.
// updateCapacityFunc is called to update ContainerManager capacity when
// device capacity changes.
func NewHandlerImpl(updateCapacityFunc func(v1.ResourceList)) (*HandlerImpl, error) {
	glog.V(2).Infof("Creating Device Plugin Handler")
	handler := &HandlerImpl{
		allDevices:       make(map[string]sets.String),
		allocatedDevices: make(map[string]sets.String),
		podDevices:       make(podDevices),
	}

	deviceManagerMonitorCallback := func(resourceName string, added, updated, deleted []pluginapi.Device) {
		var capacity = v1.ResourceList{}
		kept := append(updated, added...)
<<<<<<< HEAD
=======

		handler.Lock()
		defer handler.Unlock()

>>>>>>> axbaretto
		if _, ok := handler.allDevices[resourceName]; !ok {
			handler.allDevices[resourceName] = sets.NewString()
		}
		// For now, Handler only keeps track of healthy devices.
		// We can revisit this later when the need comes to track unhealthy devices here.
		for _, dev := range kept {
			if dev.Health == pluginapi.Healthy {
				handler.allDevices[resourceName].Insert(dev.ID)
			} else {
				handler.allDevices[resourceName].Delete(dev.ID)
			}
		}
		for _, dev := range deleted {
			handler.allDevices[resourceName].Delete(dev.ID)
		}
		capacity[v1.ResourceName(resourceName)] = *resource.NewQuantity(int64(handler.allDevices[resourceName].Len()), resource.DecimalSI)
		updateCapacityFunc(capacity)
	}

	mgr, err := NewManagerImpl(pluginapi.KubeletSocket, deviceManagerMonitorCallback)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize device plugin manager: %+v", err)
	}

	handler.devicePluginManager = mgr
	handler.devicePluginManagerMonitorCallback = deviceManagerMonitorCallback

	return handler, nil
}

// Start initializes podDevices and allocatedDevices information from checkpoint-ed state
// and starts device plugin registration service.
func (h *HandlerImpl) Start(activePods ActivePodsFunc) error {
	h.activePods = activePods

	// Loads in allocatedDevices information from disk.
	err := h.readCheckpoint()
	if err != nil {
		glog.Warningf("Continue after failing to read checkpoint file. Device allocation info may NOT be up-to-date. Err: %v", err)
	}

	return h.devicePluginManager.Start()
}

// Devices returns all of registered devices keyed by resourceName.
func (h *HandlerImpl) Devices() map[string][]pluginapi.Device {
	return h.devicePluginManager.Devices()
}

// Returns list of device Ids we need to allocate with Allocate rpc call.
// Returns empty list in case we don't need to issue the Allocate rpc call.
func (h *HandlerImpl) devicesToAllocate(podUID, contName, resource string, required int) (sets.String, error) {
	h.Lock()
	defer h.Unlock()
	needed := required
	// Gets list of devices that have already been allocated.
	// This can happen if a container restarts for example.
	devices := h.podDevices.containerDevices(podUID, contName, resource)
	if devices != nil {
		glog.V(3).Infof("Found pre-allocated devices for resource %s container %q in Pod %q: %v", resource, contName, podUID, devices.List())
		needed = needed - devices.Len()
		// A pod's resource is not expected to change once admitted by the API server,
		// so just fail loudly here. We can revisit this part if this no longer holds.
		if needed != 0 {
			return nil, fmt.Errorf("pod %v container %v changed request for resource %v from %v to %v", podUID, contName, resource, devices.Len(), required)
		}
	}
	if needed == 0 {
		// No change, no work.
		return nil, nil
	}
	devices = sets.NewString()
	// Needs to allocate additional devices.
	if h.allocatedDevices[resource] == nil {
		h.allocatedDevices[resource] = sets.NewString()
	}
	// Gets Devices in use.
	devicesInUse := h.allocatedDevices[resource]
	// Gets a list of available devices.
	available := h.allDevices[resource].Difference(devicesInUse)
	if int(available.Len()) < needed {
		return nil, fmt.Errorf("requested number of devices unavailable for %s. Requested: %d, Available: %d", resource, needed, available.Len())
	}
	allocated := available.UnsortedList()[:needed]
	// Updates h.allocatedDevices with allocated devices to prevent them
	// from being allocated to other pods/containers, given that we are
	// not holding lock during the rpc call.
	for _, device := range allocated {
		h.allocatedDevices[resource].Insert(device)
		devices.Insert(device)
	}
	return devices, nil
}

// allocateContainerResources attempts to allocate all of required device
// plugin resources for the input container, issues an Allocate rpc request
// for each new device resource requirement, processes their AllocateResponses,
// and updates the cached containerDevices on success.
func (h *HandlerImpl) allocateContainerResources(pod *v1.Pod, container *v1.Container) error {
	podUID := string(pod.UID)
	contName := container.Name
	allocatedDevicesUpdated := false
	for k, v := range container.Resources.Limits {
		resource := string(k)
		needed := int(v.Value())
		glog.V(3).Infof("needs %d %s", needed, resource)
		if _, registeredResource := h.allDevices[resource]; !registeredResource {
			continue
		}
		// Updates allocatedDevices to garbage collect any stranded resources
		// before doing the device plugin allocation.
		if !allocatedDevicesUpdated {
			h.updateAllocatedDevices(h.activePods())
			allocatedDevicesUpdated = true
		}
		allocDevices, err := h.devicesToAllocate(podUID, contName, resource, needed)
		if err != nil {
			return err
		}
		if allocDevices == nil || len(allocDevices) <= 0 {
			continue
		}
		// devicePluginManager.Allocate involves RPC calls to device plugin, which
		// could be heavy-weight. Therefore we want to perform this operation outside
		// mutex lock. Note if Allcate call fails, we may leave container resources
		// partially allocated for the failed container. We rely on updateAllocatedDevices()
		// to garbage collect these resources later. Another side effect is that if
		// we have X resource A and Y resource B in total, and two containers, container1
		// and container2 both require X resource A and Y resource B. Both allocation
		// requests may fail if we serve them in mixed order.
		// TODO: may revisit this part later if we see inefficient resource allocation
		// in real use as the result of this. Should also consider to parallize device
		// plugin Allocate grpc calls if it becomes common that a container may require
		// resources from multiple device plugins.
		resp, err := h.devicePluginManager.Allocate(resource, allocDevices.UnsortedList())
		if err != nil {
			// In case of allocation failure, we want to restore h.allocatedDevices
			// to the actual allocated state from h.podDevices.
			h.Lock()
			h.allocatedDevices = h.podDevices.devices()
			h.Unlock()
			return err
		}

		// Update internal cached podDevices state.
		h.Lock()
		h.podDevices.insert(podUID, contName, resource, allocDevices, resp)
		h.Unlock()
	}

	// Checkpoints device to container allocation information.
	return h.writeCheckpoint()
}

// Allocate attempts to allocate all of required device plugin resources,
// and update Allocatable resources in nodeInfo if necessary
func (h *HandlerImpl) Allocate(node *schedulercache.NodeInfo, attrs *lifecycle.PodAdmitAttributes) error {
	pod := attrs.Pod
	// TODO: Reuse devices between init containers and regular containers.
	for _, container := range pod.Spec.InitContainers {
		if err := h.allocateContainerResources(pod, &container); err != nil {
			return err
		}
	}
	for _, container := range pod.Spec.Containers {
		if err := h.allocateContainerResources(pod, &container); err != nil {
			return err
		}
	}

	// quick return if no pluginResources requested
	if _, podRequireDevicePluginResource := h.podDevices[string(pod.UID)]; !podRequireDevicePluginResource {
		return nil
	}

	h.sanitizeNodeAllocatable(node)

	return nil
}

// sanitizeNodeAllocatable scans through allocatedDevices in DevicePluginHandler
// and if necessary, updates allocatableResource in nodeInfo to at least equal to
// the allocated capacity. This allows pods that have already been scheduled on
// the node to pass GeneralPredicates admission checking even upon device plugin failure.
func (h *HandlerImpl) sanitizeNodeAllocatable(node *schedulercache.NodeInfo) {
	var newAllocatableResource *schedulercache.Resource
	allocatableResource := node.AllocatableResource()
	if allocatableResource.ScalarResources == nil {
		allocatableResource.ScalarResources = make(map[v1.ResourceName]int64)
	}
	for resource, devices := range h.allocatedDevices {
		needed := devices.Len()
		quant, ok := allocatableResource.ScalarResources[v1.ResourceName(resource)]
		if ok && int(quant) >= needed {
			continue
		}
		// Needs to update nodeInfo.AllocatableResource to make sure
		// NodeInfo.allocatableResource at least equal to the capacity already allocated.
		if newAllocatableResource == nil {
			newAllocatableResource = allocatableResource.Clone()
		}
		newAllocatableResource.ScalarResources[v1.ResourceName(resource)] = int64(needed)
	}
	if newAllocatableResource != nil {
		node.SetAllocatableResource(newAllocatableResource)
	}
}

// GetDeviceRunContainerOptions checks whether we have cached containerDevices
// for the passed-in <pod, container> and returns its DeviceRunContainerOptions
// for the found one. An empty struct is returned in case no cached state is found.
func (h *HandlerImpl) GetDeviceRunContainerOptions(pod *v1.Pod, container *v1.Container) *DeviceRunContainerOptions {
	h.Lock()
	defer h.Unlock()
	return h.podDevices.deviceRunContainerOptions(string(pod.UID), container.Name)
}

// updateAllocatedDevices gets a list of active pods and then frees any Devices that are bound to
// terminated pods. Returns error on failure.
func (h *HandlerImpl) updateAllocatedDevices(activePods []*v1.Pod) {
	h.Lock()
	defer h.Unlock()
	activePodUids := sets.NewString()
	for _, pod := range activePods {
		activePodUids.Insert(string(pod.UID))
	}
	allocatedPodUids := h.podDevices.pods()
	podsToBeRemoved := allocatedPodUids.Difference(activePodUids)
	if len(podsToBeRemoved) <= 0 {
		return
	}
	glog.V(5).Infof("pods to be removed: %v", podsToBeRemoved.List())
	h.podDevices.delete(podsToBeRemoved.List())
	// Regenerated allocatedDevices after we update pod allocation information.
	h.allocatedDevices = h.podDevices.devices()
}

// Checkpoints device to container allocation information to disk.
func (h *HandlerImpl) writeCheckpoint() error {
	h.Lock()
	data := h.podDevices.toCheckpointData()
	h.Unlock()

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}
	filepath := h.devicePluginManager.CheckpointFile()
	return ioutil.WriteFile(filepath, dataJSON, 0644)
}

// Reads device to container allocation information from disk, and populates
// h.allocatedDevices accordingly.
func (h *HandlerImpl) readCheckpoint() error {
	filepath := h.devicePluginManager.CheckpointFile()
	content, err := ioutil.ReadFile(filepath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read checkpoint file %q: %v", filepath, err)
	}
	glog.V(2).Infof("Read checkpoint file %s\n", filepath)
	var data checkpointData
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint data: %v", err)
	}

	h.Lock()
	defer h.Unlock()
	h.podDevices.fromCheckpointData(data)
	h.allocatedDevices = h.podDevices.devices()
	return nil
}
