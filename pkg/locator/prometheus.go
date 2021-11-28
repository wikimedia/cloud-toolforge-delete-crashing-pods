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

package locator

import (
	"context"
	"errors"
	"gerrit.wikimedia.org/r/cloud/toolforge/delete-crashing-pods/pkg/core"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	errorWrongType = errors.New("got back wrong value type")

	query = "count_over_time((sum(rate(kube_pod_container_status_restarts_total{namespace=~\"tool-[a-z0-9\\\\-]+\"}[60m])) by (namespace, pod) > 0.002)[7d:1d]) > 4"
)

// PrometheusCrashLocator finds crashing pods from Prometheus
type PrometheusCrashLocator struct {
	PrometheusURL string
}

// GetPodsToDestroy does a Prometheus query to find out any crashing pods
func (c PrometheusCrashLocator) GetPodsToDestroy() ([]core.CrashingPod, error) {
	client, err := api.NewClient(api.Config{
		Address: c.PrometheusURL,
	})

	if err != nil {
		return nil, err
	}

	v1api := v1.NewAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query, time.Now())

	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		logrus.Warningf("Got warnings while querying Prometheus for data: %v", warnings)
	}

	if result.Type() != model.ValVector {
		return nil, errorWrongType
	}

	vectorResult := result.(model.Vector)
	var pods []core.CrashingPod

	for _, sample := range vectorResult {
		pods = append(pods, core.CrashingPod{
			Namespace: string(sample.Metric["namespace"]),
			Pod:       string(sample.Metric["pod"]),
			Days:      int(sample.Value),
		})
	}

	return pods, nil
}
