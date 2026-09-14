# Leamout Self-Hosted deployment

This composition runs the customer-operated BYOC communications runtime. It does not include Backoffice, wallet/payment configuration, managed-provider jobs, or managed SIP/CDR endpoints.

For normal lifecycle operations, install and use the `leamout` CLI. The shell entry points in this directory are thin delegates and do not implement a second deployment lifecycle.

```sh
./deploy/self-hosted/install.sh
./deploy/self-hosted/update.sh
./deploy/self-hosted/uninstall.sh
```

Container image sources are shared under `containers/`. The Self-Hosted composition selects a static BYOC-only OpenSIPS configuration and the Self-Hosted server image target.
