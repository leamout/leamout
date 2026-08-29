FROM debian:bookworm-slim

ARG OPENSIPS_VERSION=4.0.1

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl \
    && curl -fsSL https://apt.opensips.org/keyring.gpg \
        -o /usr/share/keyrings/opensips.gpg \
    && printf '%s\n' \
        'deb [signed-by=/usr/share/keyrings/opensips.gpg] https://apt.opensips.org bookworm 4.0-releases' \
        > /etc/apt/sources.list.d/opensips.list \
    && printf 'Package: opensips opensips-*\nPin: version %s*\nPin-Priority: 1001\n' \
        "$OPENSIPS_VERSION" \
        > /etc/apt/preferences.d/opensips \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        opensips \
        opensips-auth-modules \
        opensips-postgres-module \
    && installed_version="$(dpkg-query -W -f='${Version}' opensips)" \
    && case "$installed_version" in \
        "$OPENSIPS_VERSION"*) ;; \
        *) echo "expected OpenSIPS $OPENSIPS_VERSION, got $installed_version" >&2; exit 1 ;; \
       esac \
    && for module in auth.so auth_db.so db_postgres.so pike.so rtpengine.so sqlops.so; do \
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
