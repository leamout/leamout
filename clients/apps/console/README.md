# Leamout console

Auth pages live in `app/(auth)`, each using a dedicated form in `components/auth`. The shared AuthForm handles layout and preview-only submission. No API requests, transaction storage, or sessions are integrated. Submitting shows a preview notice. Links allow reviewing the page designs.

Console pages live in `app/(console)` with the sidebar. Backend integration and organization selection are deferred.
