#!/usr/bin/env python3
import json
import pathlib
import sys
import xml.etree.ElementTree as ET


def fail(message: str) -> None:
    raise SystemExit(f"network-zone validation failed: {message}")


def service_networks(service: dict) -> set[str]:
    networks = service.get("networks", {})
    if isinstance(networks, dict):
        return set(networks)
    if isinstance(networks, list):
        return set(networks)
    return set()


def subnet(config: dict, network: str) -> str | None:
    entries = config.get("networks", {}).get(network, {}).get("ipam", {}).get("config", [])
    for entry in entries:
        value = entry.get("subnet")
        if value:
            return value
    return None


def ipv4(config: dict, service: str, network: str) -> str | None:
    networks = config["services"][service].get("networks", {})
    if not isinstance(networks, dict):
        return None
    attachment = networks.get(network)
    if not isinstance(attachment, dict):
        return None
    return attachment.get("ipv4_address")


config = json.load(sys.stdin)
services = config.get("services", {})

expected = {
    "postgres": {"private-control"},
    "migrate": {"private-control"},
    "redis": {"private-control"},
    "nats": {"private-control"},
    "server": {"private-control"},
    "worker": {"private-control"},
    "freeswitch": {"private-control"},
    "opensips": {"private-control", "public-signaling"},
    "rtpengine": {"private-control", "public-media"},
    "coturn": {"public-media"},
}

for service, wanted in expected.items():
    if service not in services:
        fail(f"missing service {service}")
    actual = service_networks(services[service])
    if actual != wanted:
        fail(f"{service} networks are {sorted(actual)}, expected {sorted(wanted)}")

expected_subnets = {
    "public-signaling": "172.30.0.0/24",
    "public-media": "172.31.0.0/24",
    "private-control": "172.32.0.0/24",
}
for network, wanted in expected_subnets.items():
    actual = subnet(config, network)
    if actual != wanted:
        fail(f"{network} subnet is {actual!r}, expected {wanted!r}")

expected_addresses = {
    ("opensips", "public-signaling"): "172.30.0.10",
    ("rtpengine", "public-media"): "172.31.0.10",
    ("rtpengine", "private-control"): "172.32.0.10",
    ("freeswitch", "private-control"): "172.32.0.20",
}
for (service, network), wanted in expected_addresses.items():
    actual = ipv4(config, service, network)
    if actual != wanted:
        fail(f"{service} on {network} has {actual!r}, expected {wanted!r}")

acl_root = ET.parse("deploy/freeswitch/autoload_configs/acl.conf.xml").getroot()
for list_name in ("leamout-esl", "leamout-sip"):
    acl_list = acl_root.find(f"./network-lists/list[@name='{list_name}']")
    if acl_list is None or acl_list.get("default") != "deny":
        fail(f"FreeSWITCH {list_name} ACL is missing or does not default-deny")
    cidrs = {node.get("cidr") for node in acl_list.findall("node") if node.get("type") == "allow"}
    if "172.32.0.0/24" not in cidrs:
        fail(f"FreeSWITCH {list_name} ACL does not allow the private-control subnet")
    if "172.30.0.0/24" in cidrs:
        fail(f"FreeSWITCH {list_name} ACL allows the public-signaling subnet")

profile_root = ET.parse("deploy/freeswitch/sip_profiles/internal.xml").getroot()
settings = {
    param.get("name"): param.get("value")
    for param in profile_root.findall("./settings/param")
}
for setting in ("apply-inbound-acl", "local-network-acl"):
    if settings.get(setting) != "leamout-sip":
        fail(f"FreeSWITCH {setting} is not pinned to leamout-sip")

print("network-zone topology is valid")
