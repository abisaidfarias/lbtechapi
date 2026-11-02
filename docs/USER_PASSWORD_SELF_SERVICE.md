# User password self-service (backend + frontend)

## Backend API (base `/api/v1`)

| Method | Path | Auth | Body / query |
|--------|------|------|----------------|
| POST | `/users` | Bearer | User JSON **without** `password` |
| POST | `/users/:id/resend-invitation` | Bearer | — (pending users only; 1/min per user) |
| POST | `/auth/forgot-password` | Public | `{ "email": "user@example.com" }` |
| POST | `/auth/set-password` | Public | `{ "token": "...", "new_password": "..." }` |
| GET | `/auth/password-token/validate?token=` | Public | — |
| PUT | `/users/me/password` | Bearer | `{ "old_password", "new_password" }` |
| PUT | `/users/:id/changePassword` | Bearer | **410 Gone** (deprecated) |
| POST | `/sign-in` | Public | unchanged; `403` + `code: password_not_set` if pending |

## Env (backend)

- `FRONTEND_BASE_URL` — SPA origin, no trailing slash (email links: `{FRONTEND_BASE_URL}/auth/set-password?token=...`)
- `PASSWORD_TOKEN_TTL_HOURS` — default `48`

## Frontend routes (Angular)

- `/auth/sign-in` — link to forgot password; show `password_not_set` message
- `/auth/forgot-password` — email form → `POST auth/forgot-password`
- `/auth/set-password?token=` — validate token, set password → `POST auth/set-password`

## Users module

- Remove password fields from create user form
- Remove lock / change-password action from user list
- Success copy: invitation email sent

## Invitation email not arriving

If `POST /users` or `resend-invitation` returns **201/200** but nothing shows in the inbox:

1. **API side is OK** when SMTP returns no error (check API logs for `password email accepted by SMTP`).
2. Check **Spam / Promotions** (Gmail) and search for **LB OneTrack** or **Activate your account**.
3. In **Microsoft 365** (mailbox `EMAIL_FROM`), open **Sent Items** or run **Message trace** to the recipient.
4. Links use `FRONTEND_BASE_URL`; `http://localhost:4200` in dev can increase spam scoring — use a public dev URL when possible.
5. Local redirect: `USE_LOCAL_ENV=true` + `DEV_EMAIL_TO=you@gmail.com` sends all invite/reset mail to one address for testing.

Test send from repo root: `go run ./tools/test_invite_email you@gmail.com`

## Manual test checklist

1. Admin creates user (no password) → 201 + email received
2. Open link → set password → sign-in works
3. Forgot password from login → email → reset → sign-in
4. Pending user sign-in → `password_not_set` message
5. Legacy `PUT .../changePassword` → 410
