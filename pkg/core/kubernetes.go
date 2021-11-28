//   Copyright 2021 Taavi Väänänen <hi@taavi.wtf>
//
// Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package core

import (
	"context"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesHandler deletes things inside a Kubernetes cluster
type KubernetesHandler struct {
	kubeClient *kubernetes.Clientset
	dryRun bool
}

// ObjectAPI represents something that can delete things
type ObjectAPI interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// CreateKubernetesHandlerInCluster will use credentials mounted
// by Kubernetes to create a KubernetesHandler
func CreateKubernetesHandlerInCluster(dryRun bool) (KubernetesHandler, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return KubernetesHandler{}, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return KubernetesHandler{}, err
	}

	return KubernetesHandler{
		kubeClient: client,
		dryRun: dryRun,
	}, nil
}

func (k KubernetesHandler) tryEliminateOwners(
	crashingPod CrashingPod,
	object metav1.ObjectMeta,
	kind string,
	deleter ObjectAPI,
) error {
	foundOwners := false
	for _, nextOwner := range object.GetOwnerReferences() {
		err := k.eliminateOwner(crashingPod, nextOwner)
		if err != nil {
			return err
		}
		foundOwners = true
	}

	if !foundOwners {
		logrus.Debugf("Deleting %v/%v", kind, object.Name)

		if !k.dryRun {
			err := deleter.Delete(context.TODO(), object.Name, metav1.DeleteOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (k KubernetesHandler) eliminateOwner(crashingPod CrashingPod, owner metav1.OwnerReference) error {
	if owner.Kind == "ReplicaSet" {
		api := k.kubeClient.AppsV1().ReplicaSets(crashingPod.Namespace)

		replicaSet, err := api.Get(context.TODO(), owner.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		err = k.tryEliminateOwners(
			crashingPod,
			replicaSet.ObjectMeta,
			owner.Kind,
			api,
		)
		if err != nil {
			return err
		}
	} else if owner.Kind == "Deployment" {
		api := k.kubeClient.AppsV1().Deployments(crashingPod.Namespace)

		deployment, err := api.Get(context.TODO(), owner.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		err = k.tryEliminateOwners(
			crashingPod,
			deployment.ObjectMeta,
			owner.Kind,
			api,
		)
		if err != nil {
			return err
		}
	} else {
		// TODO: support for Job, CronJob, StatefulSet at least
		logrus.Warningf("Encountered owner of unknown kind %v", owner.Kind)
	}

	return nil
}

// PodExists check if the pod in question is a real pod
func (k KubernetesHandler) PodExists(crashingPod CrashingPod) bool {
	api := k.kubeClient.CoreV1().Pods(crashingPod.Namespace)
	_, err := api.Get(context.TODO(), crashingPod.Pod, metav1.GetOptions{})

	if err != nil {
		// not found or something else, in any case doesn't exist for our purposes
		return false
	}

	return true
}

// RemovePod will eliminate the given pod and any controllers creating it
func (k KubernetesHandler) RemovePod(crashingPod CrashingPod) error {
	api := k.kubeClient.CoreV1().Pods(crashingPod.Namespace)
	pod, err := api.Get(context.TODO(), crashingPod.Pod, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Warningf("Pod %v/%v not found", crashingPod.Namespace, crashingPod.Pod)
			return nil
		}

		return err
	}

	return k.tryEliminateOwners(crashingPod, pod.ObjectMeta, "Pod", api)
}
