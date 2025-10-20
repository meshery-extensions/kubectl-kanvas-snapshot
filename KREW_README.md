## Installation Using Krew

[kubectl krew](https://github.com/kubernetes-sigs/krew) is the package manager for kubectl plugins.

Install the plugin:
```bash
kubectl krew install plugin-name
```

Use the plugin:
```bash
kubectl plugin-name [command] [flags]
```

Update to the latest version:
```bash
kubectl krew upgrade plugin-name
```

Uninstall:
```bash
kubectl krew uninstall plugin-name
```

### Manual Installation

If you prefer not to use Krew:

1. Download the appropriate binary from [Releases](https://github.com/your-org/plugin-name/releases)
2. Extract the archive
3. Copy the binary to somewhere in your PATH, or rename it to `kubectl-plugin-name` and add the directory to your PATH
4. Make it executable: `chmod +x kubectl-plugin-name`
5. Verify installation: `kubectl plugin-name --version`