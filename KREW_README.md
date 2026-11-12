# kubectl kanvas-snapshot

`kubectl kanvas-snapshot` is a native kubectl plugin for generating visual snapshots of Kubernetes manifests, collections of manifests, and directories of resources using Meshery and Kanvas.

## Installation Using Krew

[kubectl krew](https://github.com/kubernetes-sigs/krew) is the plugin manager for kubectl.

### 1. Install the plugin

```bash
kubectl krew install kanvas-snapshot
```

### 2. Verify installation

```bash
kubectl kanvas-snapshot --help
```

### 3. Update the plugin

```bash
kubectl krew upgrade kanvas-snapshot
```

### 4. Uninstall

```bash
kubectl krew uninstall kanvas-snapshot
```

---

## Manual Installation

If you prefer not to use Krew:

1. Download the latest binary archive for your OS/architecture from [Releases](https://github.com/meshery-extensions/kubectl-kanvas-snapshot/releases).
2. Extract the archive:
   ```bash
   tar -xzf kubectl-kanvas-snapshot_*.tar.gz
   ```
3. Move `kubectl-kanvas-snapshot` into your `$PATH` (e.g. `/usr/local/bin`):
   ```bash
   chmod +x kubectl-kanvas-snapshot
   sudo mv kubectl-kanvas-snapshot /usr/local/bin/
   ```
4. Verify installation:
   ```bash
   kubectl kanvas-snapshot --help
   ```

---

## Usage & Examples

### Generate a snapshot from a single manifest:
```bash
kubectl kanvas-snapshot -f ./deployment.yaml
```

### Generate a snapshot from multiple manifests:
```bash
kubectl kanvas-snapshot -f ./deployment.yaml -f ./service.yaml
```

### Generate a snapshot from a directory of manifests:
```bash
kubectl kanvas-snapshot -f ./k8s-manifests/
```

### Generate a snapshot from a remote URL with notification:
```bash
kubectl kanvas-snapshot -f https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/deployment.yaml -e user@example.com --name my-app
```

### Generate a snapshot from stdin:
```bash
cat deployment.yaml | kubectl kanvas-snapshot -f -
```

### Download the snapshot image locally:
```bash
kubectl kanvas-snapshot -f ./deployment.yaml -o ./snapshot.png
```
