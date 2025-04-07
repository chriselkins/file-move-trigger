package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"gopkg.in/yaml.v3"
)

type MoveTask struct {
	Trigger   string `yaml:"trigger"`
	Source    string `yaml:"source"`
	Target    string `yaml:"target"`
	User      string `yaml:"user"`
	Group     string `yaml:"group"`
	FileMode  string `yaml:"file_mode"` // e.g. "0640"
	DirMode   string `yaml:"dir_mode"`  // e.g. "0750"
	Overwrite bool   `yaml:"overwrite"`
}

type Config struct {
	MoveTasks []MoveTask `yaml:"move_tasks"`
}

type MoveStats struct {
	FilesMoved   int
	FilesSkipped int
	TotalFiles   int
}

var (
	configPath = flag.String("config", "/etc/file-move-trigger/config.yaml", "Path to YAML config file")
	showStats  = flag.Bool("stats", false, "Show move statistics")
	dryRun     = flag.Bool("dry-run", false, "Simulate the file moves")
	help       = flag.Bool("help", false, "Show usage")
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	flag.Parse()

	if *help {
		fmt.Println("Usage: file-move-trigger [options]")
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("Exit codes:")
		fmt.Println("  0 = success")
		fmt.Println("  2 = config error")
		fmt.Println("  3 = file operation failure")
		os.Exit(0)
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Printf("Error loading config: %v", err)
		os.Exit(2)
	}

	var totalStats MoveStats
	hadError := false

	for _, task := range cfg.MoveTasks {
		stats, err := processMoveNow(task, *dryRun)

		totalStats.FilesMoved += stats.FilesMoved
		totalStats.FilesSkipped += stats.FilesSkipped
		totalStats.TotalFiles += stats.TotalFiles

		if err != nil {
			log.Printf("Error processing task [%s → %s]: %v", task.Source, task.Target, err)
			hadError = true
		}
	}

	if *showStats {
		fmt.Printf("Move complete. Stats:\n")
		fmt.Printf("  Moved:   %d\n", totalStats.FilesMoved)
		fmt.Printf("  Skipped: %d\n", totalStats.FilesSkipped)
		fmt.Printf("  Total:   %d\n", totalStats.TotalFiles)
	}

	if hadError {
		os.Exit(3)
	}

	os.Exit(0)
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func processMoveNow(task MoveTask, dryRun bool) (MoveStats, error) {
	var stats MoveStats

	if _, err := os.Stat(task.Trigger); err != nil {
		if os.IsNotExist(err) {
			return stats, nil
		}

		return stats, fmt.Errorf("checking trigger file: %w", err)
	}

	if dryRun {
		log.Printf("[dry-run] Would remove: %s", task.Trigger)
	} else {
		if err := os.Remove(task.Trigger); err != nil {
			return stats, fmt.Errorf("removing trigger file: %w", err)
		}
	}

	entries, err := os.ReadDir(task.Source)

	if err != nil {
		return stats, fmt.Errorf("reading source directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() == filepath.Base(task.Trigger) {
			stats.FilesSkipped++
			continue
		}

		stats.TotalFiles++

		srcPath := filepath.Join(task.Source, entry.Name())
		destPath := filepath.Join(task.Target, entry.Name())

		if dryRun {
			log.Printf("[dry-run] Would move: %s → %s", srcPath, destPath)
			stats.FilesMoved++
			continue
		}

		if info, err := os.Stat(destPath); err == nil {
			if !task.Overwrite {
				log.Printf("Skipped: %s → %s (exists, overwrite disabled)", srcPath, destPath)
				stats.FilesSkipped++
				continue
			}

			if info.IsDir() {
				if err := os.RemoveAll(destPath); err != nil {
					log.Printf("Failed to remove existing dir %s: %v", destPath, err)
					stats.FilesSkipped++
					continue
				}
			} else {
				if err := os.Remove(destPath); err != nil {
					log.Printf("Failed to remove existing file %s: %v", destPath, err)
					stats.FilesSkipped++
					continue
				}
			}
		}

		moveErr := os.Rename(srcPath, destPath)
		if moveErr != nil {
			if linkErr, ok := moveErr.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
				// Cross-device move
				log.Printf("Cross-device move detected, copying: %s → %s", srcPath, destPath)
				var copyErr error
				if entry.IsDir() {
					copyErr = copyDir(srcPath, destPath)
				} else {
					copyErr = copyFile(srcPath, destPath)
				}
				if copyErr == nil {
					moveErr = os.RemoveAll(srcPath)
				} else {
					moveErr = copyErr
				}
			}
		}

		if moveErr != nil {
			log.Printf("Failed to move %s to %s: %v", srcPath, destPath, moveErr)
			stats.FilesSkipped++
			continue
		}

		log.Printf("Moved: %s → %s", srcPath, destPath)
		stats.FilesMoved++

		isDir := entry.IsDir()
		if isDir {
			if err := applyRecursivePermissions(destPath, task); err != nil {
				log.Printf("Warning: failed to apply recursive perms on %s: %v", destPath, err)
			}
		} else {
			if err := applyOwnershipAndPermissions(destPath, false, task); err != nil {
				log.Printf("Warning: failed to apply perms on %s: %v", destPath, err)
			}
		}
	}

	return stats, nil
}

func applyOwnershipAndPermissions(path string, isDir bool, task MoveTask) error {
	var mode os.FileMode
	var modeStr string

	if isDir {
		modeStr = task.DirMode
	} else {
		modeStr = task.FileMode
	}

	if modeStr != "" {
		parsed, err := strconv.ParseUint(modeStr, 8, 32)
		if err != nil {
			return fmt.Errorf("invalid mode %q: %v", modeStr, err)
		}
		mode = os.FileMode(parsed)
		if err := os.Chmod(path, mode); err != nil {
			return fmt.Errorf("chmod failed: %w", err)
		}
	}

	if task.User != "" || task.Group != "" {
		uid := -1
		gid := -1

		if task.User != "" {
			u, err := user.Lookup(task.User)
			if err != nil {
				return fmt.Errorf("user lookup %q: %v", task.User, err)
			}
			uid, _ = strconv.Atoi(u.Uid)
		}

		if task.Group != "" {
			g, err := user.LookupGroup(task.Group)
			if err != nil {
				return fmt.Errorf("group lookup %q: %v", task.Group, err)
			}
			gid, _ = strconv.Atoi(g.Gid)
		}

		if err := os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("chown failed: %w", err)
		}
	}

	return nil
}

func applyRecursivePermissions(root string, task MoveTask) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		isDir := info.IsDir()
		if err := applyOwnershipAndPermissions(path, isDir, task); err != nil {
			return fmt.Errorf("failed to apply perms on %s: %w", path, err)
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}
