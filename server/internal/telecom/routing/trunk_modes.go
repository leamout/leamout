package routing

// Trunks retain an internal provisioning mode because Cloud Managed uses a
// Leamout-operated trunk while organization-scoped customer trunks are BYOC.
// Phone numbers do not have this mode.
const (
	provisioningModeBYOC    = "byoc"
	provisioningModeManaged = "managed"
)
