# TODO

- in-browser player
- Pagination for search
- Pagination for list podcasts
- Support parsing of authenticated feeds
- Disclaimer in readme
- Internationalisation
- Font
- Message queue for downloading / other jobs
- Remove podcast
- Metrics (prometheus?)
- backoff on download failures
- Permission for feed access vs app access
- Auth housekeeping task
- Auth rate limiting
- Refactor background jobs to use scheduler
- Logout button / show active username in UI
- User management UI
- Collapse cmds into single script
- Min password length / policy
- Rate limiting or other login protections

Priority

- Tests
- Docs
- Disclaimer in readme

-- OPEN SOURCE, then...

- Dependabot (go.mod, web/package.json, website/package.json)
- Support authenticated podcasts
- Browser player

Docs

- Introduction
  - What is it?
  - Why does it exist / Motivation?
- Getting started
  - Installation
  - Deployment (Docker vs Binary)
  - Configuration file / env vars
  - Database config
  - Storage config
  - Creating users
  - Updates
- Usage
  - Subscribing to podcasts
    - Search vs feed URL
    - Authenticated feeds
    - Feed worker
    - Common errors
  - Managing podcasts
    - View podcasts
    - Checking status
  - Using CastKeeper feeds in your podcast app of choice
    - What is this, why?
    - Auth
  - Troubleshooting
    - Logs
    - Alerting/metrics etc
    - Raising issues and feedback
- Dev/contributing
  - Local dev env
  - Running tests
