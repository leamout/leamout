#!/bin/sh
set -eu

real_cli=/usr/local/freeswitch/bin/fs_cli
command=""
previous=""

for argument in "$@"; do
    if [ "$previous" = "-x" ]; then
        command=$argument
        break
    fi
    previous=$argument
done

case "$command" in
    originate\ *)
        channel_uuid=$(cat /proc/sys/kernel/random/uuid)
        dial_string=${command#originate }
        case "$dial_string" in
            \{*) dial_string="{origination_uuid=$channel_uuid,${dial_string#\{}" ;;
            *) dial_string="{origination_uuid=$channel_uuid}$dial_string" ;;
        esac

        output=$(
            "$real_cli" \
                -H 127.0.0.1 \
                -P 8021 \
                -p "$FREESWITCH_ESL_PASSWORD" \
                -x "bgapi originate $dial_string"
        )
        case "$output" in
            *"+OK Job-UUID:"*)
                printf '+OK %s\n' "$channel_uuid"
                ;;
            *)
                printf '%s\n' "$output"
                exit 1
                ;;
        esac
        ;;
    *)
        exec "$real_cli" "$@"
        ;;
esac
