# Leamout Cloud-first architecture

Leamout Cloud is the first commercial deployment of one proprietary infrastructure
platform. The runtime has three entrypoints: `server`, `worker`, and `media`.
It has no deployment-mode, deployment-ID, Community-edition or signed-license
branches. Enterprise private deployments are a future commercial expansion.

The server and worker initialize subscription plan policies regardless of
whether Polar checkout credentials are configured. Subscription allowance and
resource quotas are enforced through the existing service callbacks and database
entitlements; missing plan data must never grant unrestricted Cloud access.
The media process does not consult billing on its realtime audio path.

For future private deployments, design a separate explicit commercial policy
and deployment packaging after there is a paying customer and an agreed
contract. Do not restore edition flags to the hot path.

**Licensing:** Existing Apache-2.0 publications remain subject to that
license. Making the repository private would require a separate authorized
repository administration action, and does not revoke existing grants.
Review rights to all contributed code before relicensing future distributions.
