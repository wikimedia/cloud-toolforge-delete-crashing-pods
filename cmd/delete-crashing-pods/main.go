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

package main

import (
	"gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/core"
	"gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/locator"
	"gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/notifier"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type config struct {
	Debug         bool `default:"false"`
	DryRun        bool `default:"false"`
	PrometheusURL string
	SMTPServer    string `default:""`
	SMTPPort      int    `default:"25"`
	EmailFrom     string `default:""`
	EmailTo       string `default:""`
}

func main() {
	config := &config{}
	err := envconfig.Process("", config)
	if err != nil {
		logrus.Errorln("Failed to load config", err)
		return
	}

	if config.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	kubernetes, err := core.CreateKubernetesHandlerInCluster(config.DryRun)
	if err != nil {
		logrus.Errorln("Failed to connect to Kubernetes", err)
		return
	}

	var notifiers []notifier.Notifier

	if config.SMTPServer != "" {
		emailNotifier := notifier.EmailNotifier{
			SMTPServer:  config.SMTPServer,
			SMTPPort:    config.SMTPPort,
			FromAddress: config.EmailFrom,
			ToDomain:    config.EmailTo,
		}

		notifiers = append(notifiers, emailNotifier)
		logrus.Debugf("Created email notifier %v", emailNotifier)
	}

	crashLocator := locator.PrometheusCrashLocator{
		PrometheusURL: config.PrometheusURL,
	}

	pods, err := crashLocator.GetPodsToDestroy()
	if err != nil {
		logrus.Errorln("Failed to get pods to destroy", err)
		return
	}

	logrus.Infof("Found pods: %v", pods)

	for _, pod := range pods {
		if !kubernetes.PodExists(pod) {
			logrus.Warningf("Skipping non-existent pod %v/%v", pod.Namespace, pod.Pod)
			continue
		}

		if pod.Days == 7 {
			logrus.Infof("Removing pod %v/%v", pod.Namespace, pod.Pod)

			err := kubernetes.RemovePod(pod)
			if err != nil {
				logrus.Errorln("Failed to remove a pod", err)
				continue
			}

			if !config.DryRun {
				for _, not := range notifiers {
					err := not.TellMaintainersAboutDeath(pod)
					if err != nil {
						logrus.Errorln("Failed to notify maintainers", err)
					}
				}
			}
		} else if pod.Days == 4 {
			logrus.Infof("Warning maintainers of %v/%v about immiment death", pod.Namespace, pod.Pod)

			if !config.DryRun {
				for _, not := range notifiers {
					err := not.SendWarningToMaintainers(pod)
					if err != nil {
						logrus.Errorln("Failed to notify maintainers", err)
					}
				}
			}
		}
	}
}
