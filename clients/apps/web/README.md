# Leamout website

The marketing layout provides the shared header, footer, and main landmark. Pages inside `(marketing)` render their content within that landmark.

Run `bun install` from `clients`, then `bun run dev` from this directory. The website runs on port 3000 and links to the console on port 3001 in development. Production console links default to `https://console.leamout.com`.

For another console hostname, set `CONSOLE_URL` in this app's `.env.local` before building. Marketing pages are prerendered, so changing the URL requires a rebuild. Documentation links use the existing `/docs` redirect.
