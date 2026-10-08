# Leamout console

Run `bun run --filter @leamout/console dev` from `clients/`.

Authentication pages live in `app/(auth)` without the sidebar. Console routes live in `app/(console)` with the shared shell.

The Next.js server proxies `/api/v1` to the Go API. Set `API_URL` for deployment; it defaults to `http://localhost:8080`. Configure the Go session cookie domain and development setting for the console hostname. Only transaction ID, email, navigation step, and resend time are stored in sessionStorage. Passwords and codes remain in memory.

Signup and recovery use the existing OTP login flow, which issues a session, followed by authenticated password enrollment. Navigation purpose is not a backend authorization boundary. The backend does not yet distinguish signup/recovery transaction intents or revoke all sessions after recovery. OTP can create an account for an unknown email, including during recovery. Organization selection and console session guards remain future work; the console shell is still a preview.
