# Leamout console

Auth pages live in `app/(auth)`, each using a dedicated form in `components/auth`. React Hook Form and Zod handle form validation. The shared AuthForm handles presentation; each dedicated form handles its own preview submission feedback.

Onboarding pages live in `app/(onboarding)`, with dedicated components in `components/onboarding`. `/create-organization` validates an organization name. `/organizations` displays an empty state by default; the selector component also supports an organization list for UI previews.

Console pages live in `app/(console)` with the sidebar.

These pages are UI previews. No API requests, transaction storage, sessions, or backend guards are integrated. Submitting forms displays a preview notice without creating accounts or organizations. Links allow reviewing the page designs.
