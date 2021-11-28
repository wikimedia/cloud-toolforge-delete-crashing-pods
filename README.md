# delete-crashing-pods

delete-crashing-pods hunts for crashing pods and applies your favourite
destructive API calls to get completely rid of them!

It is implemented using a Kubernetes CronJob that runs a Golang
program, which in turn locates pods with a Prometheus query.

## Deploying

The best way to deploy delete-crashing-pods is to use Helm. First,
create a `values.yaml` file with values overriding the default ones in
`charts/delete-crashing-pods/values.yaml`. Then, run the following
command to deploy it to your Kubernetes cluster:

```shell
kubectl create ns delete-crashing-pods
helm install -f values.yaml -n delete-crashing-pods delete-crashing-pods ./charts/delete-crashing-pods/
```

## License

```
Copyright 2021 Taavi Väänänen <hi@taavi.wtf>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
