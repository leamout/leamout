#!/usr/bin/env python3
"""Read-only Polar sandbox configuration checks; never creates a purchase."""

import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid


API = "https://sandbox-api.polar.sh"
API_VERSION = "2026-10"
MAX_RESPONSE_BYTES = 1 << 20
PLANS = (
    ("developer", "POLAR_SANDBOX_DEVELOPER_PRODUCT_ID", 7900),
    ("pro", "POLAR_SANDBOX_PRO_PRODUCT_ID", 29900),
    ("scale", "POLAR_SANDBOX_SCALE_PRODUCT_ID", 79900),
)


class CheckFailure(RuntimeError):
    pass


class NoRedirects(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise CheckFailure("Sandbox API returned a redirect; refusing to forward credentials")


def required(name):
    value = os.getenv(name, "").strip()
    if not value:
        raise CheckFailure(f"{name} is required")
    return value


def identifier(value, name):
    try:
        parsed = uuid.UUID(value)
    except ValueError as error:
        raise CheckFailure(f"{name} must be a nonempty UUID") from error
    if parsed.int == 0:
        raise CheckFailure(f"{name} must be a nonempty UUID")
    return str(parsed)


def get(opener, token, path):
    request = urllib.request.Request(
        API + path,
        headers={
            "Authorization": f"Bearer {token}",
            "Polar-Version": API_VERSION,
            "Accept": "application/json",
        },
        method="GET",
    )
    try:
        with opener.open(request, timeout=20) as response:
            body = response.read(MAX_RESPONSE_BYTES + 1)
    except urllib.error.HTTPError as error:
        # Do not print customer payloads, authorization headers, or provider errors.
        raise CheckFailure(f"Sandbox GET failed with HTTP {error.code}") from error
    except urllib.error.URLError as error:
        raise CheckFailure("Cannot reach the Polar sandbox; check network access and TLS trust") from error
    if len(body) > MAX_RESPONSE_BYTES:
        raise CheckFailure("Sandbox response exceeds the maximum allowed size")
    try:
        value = json.loads(body)
    except (json.JSONDecodeError, UnicodeDecodeError) as error:
        raise CheckFailure("Sandbox response is not valid JSON") from error
    if not isinstance(value, dict):
        raise CheckFailure("Sandbox response must be a JSON object")
    return value


def validate_product(product, product_id, plan, amount):
    if product.get("id") != product_id:
        raise CheckFailure(f"{plan}: product ID does not match the requested product")
    if product.get("is_archived") is not False:
        raise CheckFailure(f"{plan}: product is archived or its availability is unknown")
    if product.get("is_recurring") is not True:
        raise CheckFailure(f"{plan}: product must have recurring pricing")
    if product.get("recurring_interval") != "month":
        raise CheckFailure(f"{plan}: expected monthly billing")
    if product.get("recurring_interval_count", 1) != 1:
        raise CheckFailure(f"{plan}: expected one-month billing intervals")
    prices = product.get("prices")
    if not isinstance(prices, list):
        raise CheckFailure(f"{plan}: product prices are missing")
    matches = [
        price
        for price in prices
        if isinstance(price, dict)
        and price.get("amount_type") == "fixed"
        and price.get("price_amount") == amount
        and price.get("price_currency", "").lower() == "usd"
        and price.get("is_archived", False) is False
    ]
    if not matches:
        raise CheckFailure(f"{plan}: expected an active USD price of {amount} cents")
    organization_id = product.get("organization_id")
    if not organization_id:
        raise CheckFailure(f"{plan}: Polar organization is missing")
    return organization_id


def main():
    token = required("POLAR_SANDBOX_ACCESS_TOKEN")
    products = [
        (plan, identifier(required(name), name), amount)
        for plan, name, amount in PLANS
    ]
    if len({product_id for _, product_id, _ in products}) != len(products):
        raise CheckFailure("Sandbox product IDs must be distinct")

    opener = urllib.request.build_opener(NoRedirects())
    organizations = set()
    for plan, product_id, amount in products:
        product = get(opener, token, "/v1/products/" + product_id)
        organizations.add(validate_product(product, product_id, plan, amount))
        print(f"PASS {plan}: sandbox product, monthly USD price, availability")
    if len(organizations) != 1:
        raise CheckFailure("Sandbox products must belong to the same Polar organization")

    external_id = os.getenv("POLAR_SANDBOX_CUSTOMER_EXTERNAL_ID", "").strip()
    if external_id:
        external_id = identifier(external_id, "POLAR_SANDBOX_CUSTOMER_EXTERNAL_ID")
        path = "/v1/customers/external/" + urllib.parse.quote(external_id, safe="") + "/state"
        state = get(opener, token, path)
        if state.get("external_id") != external_id:
            raise CheckFailure("Customer State external ID does not match the Leamout organization")
        subscriptions = state.get("active_subscriptions")
        if not isinstance(subscriptions, list):
            raise CheckFailure("Customer State active subscriptions are missing")
        product_ids = {product_id for _, product_id, _ in products}
        known = [
            subscription
            for subscription in subscriptions
            if isinstance(subscription, dict)
            and subscription.get("product_id") in product_ids
        ]
        print(f"PASS Customer State: identity matches; {len(known)} Leamout subscriptions")
    else:
        print("SKIP Customer State: POLAR_SANDBOX_CUSTOMER_EXTERNAL_ID was not supplied")
    print("Read-only sandbox checks complete. Purchase and payment lifecycle checks remain manual.")


if __name__ == "__main__":
    try:
        main()
    except CheckFailure as error:
        print(f"FAIL {error}", file=sys.stderr)
        sys.exit(1)
