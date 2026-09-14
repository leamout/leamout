# Leamout Cloud deployment

This composition runs the Leamout-operated Cloud API, worker, Backoffice, prepaid Commercial surface, and managed-carrier boundary.

```sh
docker compose --env-file .env -f deploy/cloud/compose.yaml up -d --build
```

Container image sources are shared under `containers/`. The Cloud composition selects the Cloud OpenSIPS configuration and the Cloud server image target.
