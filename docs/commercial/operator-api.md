# Commercial operator API

The commercial operator API is a privileged machine-to-machine surface for administering Leamout-owned commercial state. It is intentionally separate from the customer API under `/internal/v1/commercial` and must not be exposed directly to the public internet.

## Authentication

Set `OPERATOR_API_KEY` to a high-entropy secret and send it as a bearer credential:

```http
Authorization: Bearer <OPERATOR_API_KEY>
```

When `OPERATOR_API_KEY` is empty, every operator request fails closed with `401 Unauthorized`. Use a secret manager in production, rotate the credential operationally, and restrict the route at the network layer as an additional defense.

## Resources

The first operator surface exposes:

- read-only product, plan, and price discovery;
- subscription acquisition, inspection, term changes, provider references, and lifecycle transitions;
- plan entitlement defaults and organization entitlement overrides;
- resolved organization commercial state;
- license creation, inspection, lifecycle transitions, and deployment inspection.

Catalog discovery defaults to active records. Pass `?active_only=false` to include inactive and historical records.

## Route summary

```text
GET    /internal/v1/commercial/products
GET    /internal/v1/commercial/products/{product_id}
GET    /internal/v1/commercial/products/{product_id}/plans
GET    /internal/v1/commercial/plans/{plan_id}
GET    /internal/v1/commercial/plans/{plan_id}/prices
GET    /internal/v1/commercial/prices/{price_id}
POST   /internal/v1/commercial/plans/{plan_id}/entitlements
GET    /internal/v1/commercial/plans/{plan_id}/entitlements
DELETE /internal/v1/commercial/plans/{plan_id}/entitlements/{entitlement_id}

POST   /internal/v1/commercial/organizations/{organization_id}/subscriptions
GET    /internal/v1/commercial/organizations/{organization_id}/subscriptions
GET    /internal/v1/commercial/organizations/{organization_id}/subscriptions/current
GET    /internal/v1/commercial/organizations/{organization_id}/subscriptions/{subscription_id}
PUT    /internal/v1/commercial/organizations/{organization_id}/subscriptions/{subscription_id}/status
PUT    /internal/v1/commercial/organizations/{organization_id}/subscriptions/{subscription_id}/price
PUT    /internal/v1/commercial/organizations/{organization_id}/subscriptions/{subscription_id}/period
PUT    /internal/v1/commercial/organizations/{organization_id}/subscriptions/{subscription_id}/provider
GET    /internal/v1/commercial/organizations/{organization_id}/state
POST   /internal/v1/commercial/organizations/{organization_id}/entitlements
GET    /internal/v1/commercial/organizations/{organization_id}/entitlements
DELETE /internal/v1/commercial/organizations/{organization_id}/entitlements/{entitlement_id}
POST   /internal/v1/commercial/organizations/{organization_id}/licenses
GET    /internal/v1/commercial/organizations/{organization_id}/licenses
GET    /internal/v1/commercial/organizations/{organization_id}/licenses/{license_id}
PUT    /internal/v1/commercial/organizations/{organization_id}/licenses/{license_id}/status
GET    /internal/v1/commercial/organizations/{organization_id}/licenses/{license_id}/deployments
```

The operator API does not activate deployments or produce deployment-bound signed artifacts. Those operations belong to the separate self-hosted deployment protocol.
