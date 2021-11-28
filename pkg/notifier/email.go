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

import "gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/core"

// EmailNotifier sends email notifications about killed pods
type EmailNotifier struct {
}

// SendWarningToMaintainers sends an advance warning
func (n EmailNotifier) SendWarningToMaintainers(pod core.CrashingPod) {
	// TODO: implement
}

// TellMaintainersAboutDeath delivers the sad news after a pods was cleaned up
func (n EmailNotifier) TellMaintainersAboutDeath(pod core.CrashingPod) {
	// TODO: implement
}
