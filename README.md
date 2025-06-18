# SBOMbastic

SBOMbastic is a SBOM-centric security scanner for Kubernetes. It provides native Kubernetes resources and integrates seamlessly with Rancher and other SUSE tooling.

## Key Features

### 1. Kubernetes-Native, Event-Driven Architecture
- Generates SBOM CRs (Software Bill of Materials)
- Generates Vulnerability Report CRs

### 2. SBOM-Centric Design
- Image contents change less frequently than vulnerability definitions
- Scanning an image is more expensive compared to generating an SBOM

## Use Cases

- Visualize results in the Rancher UI
- Feed data into Kubewarden for policy enforcement
- Export metrics to SUSE Observability for centralized monitoring

## Learn More

- [Source Code](https://github.com/rancher-box/sbombastic)
# SBOMbastic Storage

The `storage` Helm chart installs the SBOMbastic storage deployment, which should be installed alongside the SBOMbastic controller and worker components.

The storage component uses SQLite as its database backend. **Note that SQLite is intended for development and testing purposes only, and should not be used in production environments.**

To ensure data persistence, the storage component requires a PersistentVolumeClaim (PVC). You can provide your own PVC to control how and where data is stored.

There are two ways to satisfy this requirement:

1. Provide a pre-created PVC and reference it in your Helm values using `persistence.storageData.existingClaim`.
2. If no PVC is provided, and your cluster supports dynamic provisioning via a `StorageClass`, a new PVC and corresponding PV will be created automatically.
