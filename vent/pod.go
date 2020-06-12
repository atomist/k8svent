// Copyright © 2020 Atomist
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

package vent

import (
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
)

// processPods iterates through the provided pods and processes those
// that do not have an identical pod in lastPods.  It returns a map of
// successfully processed pods.
func (v *Venter) processPods(pods []v1.Pod, lastPods map[string]v1.Pod) map[string]v1.Pod {
	newPods := map[string]v1.Pod{}
	for _, pod := range pods {
		slug := podSlug(pod)
		log := logger.WithField("pod", slug)
		newPods[slug] = pod
		if lastPod, ok := lastPods[slug]; ok {
			if podHealthy(pod) && cmp.Diff(pod, lastPod) == "" {
				log.Debug("Pod is healthy and state is unchanged")
				continue
			}
		}
		if err := v.processPod(pod); err != nil {
			log.Errorf("Failed to process pod: %v", err)
			delete(newPods, slug)
			continue
		}
	}
	return newPods
}

// podSlug returns a string uniquely identifying a pod in a Kubernetes
// cluster.
func podSlug(pod v1.Pod) string {
	return pod.ObjectMeta.Namespace + "/" + pod.ObjectMeta.Name
}

// podHealthy determines if a pod is healthy.
func podHealthy(pod v1.Pod) bool {
	if pod.Status.Phase != v1.PodRunning {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Status != v1.ConditionTrue {
			return false
		}
	}
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if !containerHealthy(containerStatus, true) {
			return false
		}
	}
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerHealthy(containerStatus, false) {
			return false
		}
	}
	return true
}

// containerHealthy interrogates the container status to determine if
// the container is fully healthy.
func containerHealthy(containerStatus v1.ContainerStatus, init bool) bool {
	if !containerStatus.Ready {
		return false
	}
	if containerStatus.State.Waiting != nil {
		return false
	}
	if init {
		if containerStatus.State.Terminated != nil {
			if containerStatus.State.Terminated.ExitCode != 0 {
				return false
			}
		} else if containerStatus.State.Running == nil {
			return false
		}
	} else {
		if containerStatus.State.Terminated != nil {
			return false
		}
		if containerStatus.State.Running == nil {
			return false
		}
	}
	return true
}

// ProcessPods iterates through the pods and calls PostToWebhooks for
// each.
func (v *Venter) processPod(pod v1.Pod) error {
	podEnv := k8sPodEnv{
		Pod: pod,
		Env: v.env,
	}
	postToWebhooks(v.urls, &podEnv, v.secret)
	return nil
}
