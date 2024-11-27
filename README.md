# Metadata Reflector

> Reflect metadata from source objects to targets

## 📖 General Information

The Metadata Reflector controller allows you to replicate specific metadata from a source object to targets and control their state.

Example:
- **Problem**: we need to dynamically update labels on pods managed by a deployment without restarting them. Our system only manages deployments so we cannot label pods directly.
- **Solution**: we can set & update the needed labels on the deployment itself and Metadata Reflector will find pods managed by the deployment and replicate the specified metadata.

### Usage

To start using the Metadata Reflector, annotate the source `Deployment` with the [supported annotation](#supported-annotations), e.g. `labels.metadata-reflector.spaceship.com/list`:

```yaml
kind: Deployment
metadata:
  name: my-app
  annotations:
    labels.metadata-reflector.spaceship.com/list: "feature-x,feature-y"
  labels:
    feature-x: "true"
```

Metadata Reflector will find pods managed by this deployment using `.spec.selector` and replicate them to the managed pods.

Metadata Reflector will also annotate managed pods with a list of labels that were reflected, in this case `labels.metadata-reflector.spaceship.com/reflected-list: "feature-x"`.
> NOTE: This list does not contain labels that should reflected but are not present on the deployment itself.

If the label gets deleted/updated on the deployment, it will be deleted from the corresponding pods as well.

If the annotation is deleted, all managed pods will also lose the label.

Additionally, the presence of propagated labels will be checked in the background periodically.

### Configuration

It's possible to limit the watched resources and namespaces as well as configure the background job and other features. For more information, please check [environments.md](environments.md)


#### <a id="supported-annotations"></a> Supported Annotations

Below is a table of supported annotations with their purpose

| Annotation    | Description |
| ------------- | ----------- |
| `labels.metadata-reflector.spaceship.com/list`  | A comma-separated list of labels to reflect from the object that the annotation is added to |
| `labels.metadata-reflector.spaceship.com/regex`  | A regular expression to list the labels that will be reflected from the object that the annotation is added to |
| `labels.metadata-reflector.spaceship.com/reflected-list`  | A comma-separated list of labels reflected by Metadata Reflector to target objects. The annotation is only added to target objects |

### Features

Below is a list of implemented features and features that could fit into this project but are not yet implemented:

- [x] Label reflection from `Deployment`s to managed `Pod`s
- [ ] Annotation reflection from `Deployment`s to managed `Pod`s
- [ ] Label & Annotation reflection from an arbitrary source (e.g. Secret, ConfigMap, etc.) to an arbitrary target (e.g. `Deployment`, etc.)
- [x] A background job to periodically check the state of the target resources

The priority of each feature will depend on the number of relevant use cases.
