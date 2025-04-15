# 🧙‍♂️ MATT Daemon – Modular Automation Trigger Tool

> “He’s not just a daemon… he’s the MATT-daemon.”

`matt-daemon` is a no-nonsense, file-based task automation daemon for Linux. Think of it as your personal task butler: silently watching your directories and springing into action when a trigger file appears.

Whether you're moving freshly downloaded media, kicking off backups, or launching scripts like a DevOps ninja, `matt-daemon` is here for it. Built for security, speed, and full systemd integration — because crons are boring and inotify is beautiful.
 
> "One small file for man, one giant task for matt-daemon."

## 🔍 Key Features

- **File-Based Triggers**: Monitor designated directories for the creation of specific files to initiate tasks.
- **Flexible Task Execution**: Configure tasks to move files/directories, set ownership and permissions, or execute arbitrary scripts.
- **Safe and Idempotent**: Automatically removes trigger files upon task initiation to prevent duplicate executions.
- **Systemd Integration**: Seamlessly integrates with systemd for easy management, startup, and logging.
- **Configurable Behavior**: Customize task parameters, including overwrite policies and permission settings, through a YAML configuration file.
- **Dry run and Stats**: Supports dry-run and stats modes.
- **Efficient**: Uses inotify on Linux to efficiently monitor file system events.

> "Drop a file. Trigger a task. Deploy with style."

## 🚀 Use Cases

- Automating deployment processes by dropping a trigger file to initiate scripts.
- Organizing files by moving them to designated directories with specific permissions.
- Running maintenance tasks or backups triggered by simple file creation.

> "Your friendly file-watching daemon — mission-ready."

## 📦 Installation

Just clone this repo and run the install/upgrade script.

```bash
install_or_upgrade.sh
```

The install/upgrade script installs the binary to /usr/local/sbin/matt-daemon, upgrades it if necessary, installs the systemd service, enables the service, and installs a default configuration file to /etc/matt-daemon/config.yaml if one does not already exist. It won't overwrite your existing configuration file.

> "He watches. He waits. He runs your scripts."

## 🎯 Task Types

You can configure different types of tasks.

- **Move Tasks** are focused on moving files and directories with options for permissions and ownership.
- **Generic Tasks** provide greater flexibility and can run any executable task, making them suitable for more complex workflows.

## 🧾 Configuration Example (`/etc/matt-daemon/config.yaml`)

The configuration file is a YAML file consisting of a list of different kinds of tasks including `move_tasks` and `generic_tasks`.

Each **move task** defines:

| Field       | Type       | Required | Description                                                                |
|-------------|------------|----------|----------------------------------------------------------------------------|
| `trigger`   | string     | ✅       | Path to a file that triggers the move task                                  |
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

Each **generic task** defines:

| Field       | Type       | Required | Description                                                                |
|-------------|------------|----------|----------------------------------------------------------------------------|
| `trigger`   | string     | ✅       | Path to a file that triggers the task                                       |
| `run`       | string[]   | ✅       | One or more shell commands to run                                           |

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

generic_tasks:
  - trigger: /triggers/run_backup.now
    run:
      - "/usr/local/bin/full-backup.sh"
      - "/usr/local/bin/some-other-script.sh"

  - trigger: /var/www/html/stats.csv
    run:
      - "/usr/local/bin/generate-report.sh"

```

## 📄 Viewing Logs

matt-daemon logs everything to `stderr`, which is captured by `systemd` and viewable using the journal.

To follow logs in real time:

```bash
journalctl -u matt-daemon.service -f
```

## 🛠 Use Case Example

You run:
- **Transmission** as one user (e.g., `transmission`)
- **Samba** to manage and copy files to a staging folder as a different user (e.g., `chris`)
- **Plex** with its own user (e.g., `plex`)

`matt-daemon` keeps a close watch on your source folders, patiently waiting for a file named move.now to appear. When the user `chris` drops a `move.now` file into a movie or TV staging folder, the daemon springs into action — automatically moving all media files from the staging area into the proper Plex library folder, applying the correct permissions along the way.

---


