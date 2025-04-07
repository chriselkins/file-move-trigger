# file-move-trigger

file-move-trigger is an automation daemon for safely moving files based on triggers. It does this based on the existence of particular files to trigger actions. It can be configured to monitor multiple directories and when a certain file is created, a task is launched to perform moving files and setting correct destination ownership and permissions. The trigger file can be created to launch a task and the task removes the trigger file once it begins. Just create the trigger file to launch it again. It's designed for security, performance, and full Linux systemd integration.

## ✨ Features

- Tasks triggered by presence of files
- Moves individual files or whole directories
- Configurable user/group ownership and file permissions
- Optional overwrite behavior
- Safe across partitions (handles cross-device moves)
- Recursively sets permissions on directories
- Supports dry-run and stats modes
- Integrates cleanly with systemd for automation
- Uses inotify on Linux to efficiently monitor file system events

## 📦 Installation

```bash
install_or_upgrade.sh
```

The install/upgrade script installs the binary to /usr/local/sbin/file-move-trigger, upgrades it if necessary, installs the systemd service, enables the service, and installs a default configuration file to /etc/file-move-trigger/config.yaml if one does not already exist.

## 🧾 Configuration Example (`/etc/file-move-trigger/config.yaml`)

The configuration file is a YAML file consisting of a list of `move_tasks`. Each task defines:

| Field       | Type       | Required | Description                                                                 |
|-------------|------------|----------|-----------------------------------------------------------------------------|
| `trigger`   | string     | ✅       | Path to a file that triggers the move (typically named `move.now`)         |
| `source`    | string     | ✅       | Directory to move files *from*                                              |
| `target`    | string     | ✅       | Directory to move files *to*                                                |
| `user`      | string     | optional | Set the owner of the moved files/folders                                   |
| `group`     | string     | optional | Set the group for the moved files/folders                                  |
| `file_mode` | string     | optional | File permission mode (e.g., `"0640"`)                                      |
| `dir_mode`  | string     | optional | Directory permission mode (e.g., `"0750"`)                                 |
| `overwrite` | boolean    | optional | If `true`, replaces files/folders that already exist in the `target` dir   |
| `pre`       | string[]   | optional | One or more shell commands to run **before** moving begins                 |
| `post`      | string[]   | optional | One or more shell commands to run **after** move is complete               |

pre hooks must succeed (exit code 0), or the move task is aborted. post hooks are executed even if file moves fail, and their errors are only logged.

### ✅ Example:

```yaml
move_tasks:
  - trigger: /storage2/Temp/Ready/Movies/move.now
    source: /storage2/Temp/Ready/Movies/
    target: /storage1/Movies/
    user: plex
    group: plexreaders
    file_mode: "0640"
    dir_mode: "0750"
    overwrite: true
    pre:
      - "/usr/local/bin/pre-movie-hook.sh"
    post:
      - "/usr/local/bin/post-movie-hook.sh"
      - "logger 'Finished moving movies'"

  - trigger: /storage2/Temp/Ready/TV/move.now
    source: /storage2/Temp/Ready/TV/
    target: /storage1/TV/
    user: plex
    group: plexreaders
    file_mode: "0640"
    dir_mode: "0750"
    overwrite: false
```

## 📄 Viewing Logs

file-move-trigger logs everything to `stderr`, which is captured by `systemd` and viewable using the journal.

To follow logs in real time:

```bash
journalctl -u file-move-trigger.service -f
```

## 🛠 Use Case Example

You run:
- **Transmission** as one user (e.g., `transmission`)
- **Samba** to manage and copy files to a staging folder as a different user (e.g., `chris`)
- **Plex** with its own user (e.g., `plex`)

file-move-trigger monitors the source folders and waits for files named `move.now` to exist. Once the user chris creates the `move.now` file in movie or tv staging folders, the task for that folder is triggered to move all the media files in the source folder to the destination Plex library with the correct permissions.

---


