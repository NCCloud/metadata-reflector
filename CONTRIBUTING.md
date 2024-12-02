# Contributing Guidelines

We are grateful for your willingness to contribute to this project! We are interested in any features, bug fixes, new usage examples, etc.

## How to Contribute

1. Fork this repository, develop, and test your changes.
2. Submit a pull request.
3. Make sure all tests are passing.

***NOTE***: In order to make testing and merging of PRs easier, please submit changes for different fixes/features/improvements in separate PRs.

### Technical Requirements

* Must follow [Golang best practices](https://go.dev/doc/effective_go)
* Must pass CI jobs for linting. Please run `make lint` in the root of the project to know if the project complies with the requirements.
* All changes require reviews from the responsible organization members before merge.

Once changes have been merged, the release will be done by the responsible organization members.

## 🛠 Development

You can easily run the controller by following these steps:

1) Create a Kubernetes cluster or change the `kubectl` context for the existing one.

```bash
make cluster
```

2) Make changes.
3) Build the controller image and load it into the cluster.

```bash
make docker-build
make docker-load
```

4) Deploy the helm chart with your image. If your changes require updating the Helm chart itself, please open a corresponding PR in this repository https://github.com/NCCloud/charts

```
helm repo add nccloud https://nccloud.github.io/charts
helm install metadata-reflector nccloud/metadata-reflector --set image.tag="0.1.0-dev" # override the tag with the built image
```

Alternatively, configure the environment variables according to [environments.md](environments.md) and run the project.

```bash
export NAMESPACES="default,namespace1" # Metadata Reflector will be limited to these namespaces
go run cmd/manager/main.go
```

5) Destroy the cluster once you are doing testing your changes.

```bash
make cluster-delete
```
