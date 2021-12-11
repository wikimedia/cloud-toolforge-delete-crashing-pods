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

package notifier

import (
	"fmt"
	"gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/core"
	"net/smtp"
	"strings"
)

// EmailNotifier sends email notifications about killed pods
type EmailNotifier struct {
	SMTPServer  string
	SMTPPort    int
	FromAddress string
	ToDomain    string
}

func (n EmailNotifier) sendMailToMaintainers(pod core.CrashingPod, subject string, content string) error {
	toolName := strings.Replace(pod.Namespace, "tool-", "", 1)

	body := fmt.Sprintf("From: %v\r\nSubject: %v\r\n\r\n%v", n.FromAddress, subject, content)

	return smtp.SendMail(
		fmt.Sprintf("%v:%v", n.SMTPServer, n.SMTPPort),
		nil,
		n.FromAddress,
		[]string{
			fmt.Sprintf("%v.maintainers@%v", toolName, n.ToDomain),
		},
		[]byte(body),
	)
}

// SendWarningToMaintainers sends an advance warning
func (n EmailNotifier) SendWarningToMaintainers(pod core.CrashingPod) error {
	toolName := strings.Replace(pod.Namespace, "tool-", "", 1)

	return n.sendMailToMaintainers(
		pod,
		fmt.Sprintf("[Toolforge] %v tool is having issues", toolName),
		"Hello, Tool maintainer!\r\n\r\n" +
			fmt.Sprintf("The '%v' Toolforge tool is running a Kubernetes pod\r\n", toolName) +
			"that is constantly crashing: \r\n\r\n" +
			fmt.Sprintf(" * %v\r\n\r\n", pod.Pod) +
			"This is likely an issue with the tool itself, please check the logs and\r\n" +
			"fix the problem. The crashing workload will be automatically removed\r\n" +
			"from the cluster in 3 days unless the problem is fixed.\r\n\r\n" +
			"For further support, visit <ircs://irc.libera.chat/wikimedia-cloud>\r\n" +
			"or <https://wikitech.wikimedia.org>.\r\n\r\n" +
			"Thank you.\r\n",
	)
}

// TellMaintainersAboutDeath delivers the sad news after a pods was cleaned up
func (n EmailNotifier) TellMaintainersAboutDeath(pod core.CrashingPod) error {
	toolName := strings.Replace(pod.Namespace, "tool-", "", 1)

	return n.sendMailToMaintainers(
		pod,
		fmt.Sprintf("[Toolforge] %v tool is having issues", toolName),
		"Hello, Tool maintainer!\r\n\r\n" +
			fmt.Sprintf("The '%v' Toolforge tool is running a Kubernetes pod\r\n", toolName) +
			"that is constantly crashing: \r\n\r\n" +
			fmt.Sprintf(" * %v\r\n\r\n", pod.Pod) +
			"The workload has automatically been removed to free cluster resources.\r\n" +
			"You can fix your application and re-deploy it if it is still needed.\r\n\r\n" +
			"For further support, visit <ircs://irc.libera.chat/wikimedia-cloud>\r\n" +
			"or <https://wikitech.wikimedia.org>.\r\n\r\n" +
			"Thank you.\r\n",
	)
}
