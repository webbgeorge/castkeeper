---
sidebar_position: 3
---

# Managing Users

CastKeeper requires username and password authentication to access the web UI
and to use [CastKeeper feeds](/usage/listening-to-podcasts#castkeeper-feed).
To support this, CastKeeper includes a built-in user management system, giving
admins fine-grained control over access.

## Access levels

Each user is assigned one of the following access levels:

- **Read-Only** – can browse and listen to podcasts and episodes, but cannot make
  changes.
- **Manage Podcasts** – can add, edit, and remove podcasts, but cannot manage users
  or system settings.
- **Administrator** – Full access, including user management and system
  settings.

## Managing users via the web interface

Admins (i.e. users with the Admin access level) can manage users through the
**Manage Users** interface of the web UI:

- **Create a User** – enter a username, password, and select an access level.
- **Edit a User** – update details such as username, password, or access level.
- **Delete a User** – permanently remove a user.

The Manage Users section is accessed by clicking the user icon in the top right
of the web UI and selecting "Manage Users". This is only visible for users with
the Admin access level.

## Managing users via the CLI

CastKeeper also provides CLI commands for managing users:

- `castkeeper users list` – list all existing users.
- `castkeeper users create` – create a new user.
- `castkeeper users edit` – edit details such as username or access level.
- `castkeeper users change-password` – change the user's password.
- `castkeeper users delete` – delete a user.

Run any command with `--help` to see full usage details.

## Updating your own password

All users, regardless of access level, are able to update their own passwords
in the web UI, by clicking on the user icon in the top right and selecting
update password.

## CastKeeper feed authentication

Users must also authenticate to listen to podcasts using [CastKeeper feeds](/usage/listening-to-podcasts#castkeeper-feed).
These use the same username and passwords as above, except they are provided
to your podcast player instead of being entered in the CastKeeper web UI.

All users are able to access CastKeeper feeds, including users with the
Read-Only access level. Therefore, it is recommended that a specific Read-Only
user is created for this purpose to avoid exposing admin credentials.
