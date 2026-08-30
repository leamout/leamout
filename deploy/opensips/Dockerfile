FROM debian:bookworm-slim

ARG OPENSIPS_VERSION=4.0.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && curl -fsSL https://apt.opensips.org/keyring.gpg \
        -o /usr/share/keyrings/opensips.gpg \
    && printf '%s\n' \
        'deb [signed-by=/usr/share/keyrings/opensips.gpg] https://apt.opensips.org bookworm 4.0-releases' \
        > /etc/apt/sources.list.d/opensips.list \
    && apt-get update \
    && package_version="$(apt-cache madison opensips | awk -v wanted="$OPENSIPS_VERSION" '$3 ~ ("^" wanted "([.-]|$)") {print $3; exit}')" \
    && test -n "$package_version" || { \
        echo "OpenSIPS $OPENSIPS_VERSION is unavailable in the official 4.0 release repository" >&2; \
        exit 1; \
       } \
    && apt-get install -y --no-install-recommends \
        "opensips=$package_version" \
        "opensips-auth-modules=$package_version" \
        "opensips-postgres-module=$package_version" \
    && installed_version="$(dpkg-query -W -f='${Version}' opensips)" \
    && test "$installed_version" = "$package_version" \
    && for module in \
        auth.so \
        auth_db.so \
        db_postgres.so \
        pike.so \
        rtpengine.so \
        sqlops.so \
        uac.so \
        uac_auth.so; do \
        find /usr/lib -type f -name "$module" -print -quit | grep -q . || { \
            echo "required OpenSIPS module missing: $module" >&2; \
            exit 1; \
        }; \
       done \
    && apt-get purge -y --auto-remove curl \
    && rm -rf /var/lib/apt/lists/*

COPY opensips.cfg /etc/opensips/opensips.cfg
COPY tls.cfg /etc/opensips/tls.cfg

RUN opensips -C -f /etc/opensips/opensips.cfg

EXPOSE 5060/udp 5060/tcp

CMD ["opensips", "-F", "-f", "/etc/opensips/opensips.cfg"]
