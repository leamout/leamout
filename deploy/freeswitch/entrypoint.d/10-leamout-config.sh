#!/bin/sh
set -eu

install -d \
  /etc/freeswitch/autoload_configs \
  /etc/freeswitch/sip_profiles \
  /etc/freeswitch/dialplan

install -m 0644 \
  /leamout-config/autoload_configs/event_socket.conf.xml \
  /etc/freeswitch/autoload_configs/event_socket.conf.xml

install -m 0644 \
  /leamout-config/sip_profiles/internal.xml \
  /etc/freeswitch/sip_profiles/internal.xml

install -m 0644 \
  /leamout-config/dialplan/leamout.xml \
  /etc/freeswitch/dialplan/leamout.xml
