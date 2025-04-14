# 🧰 MATT Daemon – Modular Automation Trigger Tool

**`matt-daemon`** is a lightweight Linux daemon designed to automate tasks by monitoring for the presence of specific trigger files. Upon detecting these files, it executes predefined actions such as moving files or directories, adjusting permissions, or running custom scripts. This tool is ideal for system administrators and developers seeking a straightforward method to initiate tasks without the complexity of full-fledged automation frameworks. It's designed for security, performance, and full Linux systemd integration.

## 🔍 Key Features

- **File-Based Triggers**: Monitor designated directories for the creation of specific files to initiate tasks.
- **Flexible Task Execution**: Configure tasks to move files/directories, set ownership and permissions, or execute arbitrary scripts.
- **Safe and Idempotent**: Automatically removes trigger files upon task initiation to prevent duplicate executions.
- **Systemd Integration**: Seamlessly integrates with systemd for easy management, startup, and logging.
- **Configurable Behavior**: Customize task parameters, including overwrite policies and permission settings, through a YAML configuration file.
- **Dry run and Stats**: Supports dry-run and stats modes.
- **Efficient**: Uses inotify on Linux to efficiently monitor file system events.

## 🚀 Use Cases

- Automating deployment processes by dropping a trigger file to initiate scripts.
- Organizing files by moving them to designated directories with specific permissions.
- Running maintenance tasks or backups triggered by simple file creation.

## 📦 Installation

Just clone this repo and run the install/upgrade script.

```bash
install_or_upgrade.sh
```

The install/upgrade script installs the binary to /usr/local/sbin/matt-daemon, upgrades it if necessary, installs the systemd service, enables the service, and installs a default configuration file to /etc/matt-daemon/config.yaml if one does not already exist. It won't overwrite your existing configuration file.

## 🧾 Configuration Example (`/etc/matt-daemon/config.yaml`)

The configuration file is a YAML file consisting of a list of different kinds of tasks including `move_tasks` and `generic_tasks`.

Each move task defines:

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

Each generic task defines:

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
  - trigger: /home/chris/some-file.txt
    run:
      - "/usr/local/bin/some-script.sh"
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

matt-daemon monitors the source folders and waits for files named `move.now` to exist. Once the user chris creates the `move.now` file in movie or tv staging folders, the task for that folder is triggered to move all the media files in the source folder to the destination Plex library with the correct permissions.

---


