# Backups

Rsync-based backup jobs managed by Ansible. Each job gets a script at `/usr/local/bin/backup-<name>` and a cron entry at `/etc/cron.d/backup-<name>`.

## Current jobs

All run on dev-ct (which has bind mounts to the storage drives).

| Job | Source | Destination | Schedule |
|-----|--------|-------------|----------|
| footage | `/mnt/hdd-6tb/footage/` | `/mnt/hdd-3tb/footage/` | 3am daily |
| jellyfin | `/mnt/hdd-3tb/jellyfin/` | `/mnt/hdd-6tb/jellyfin/` | 4am daily |

## Running manually

```bash
sudo backup-footage
sudo backup-jellyfin
```

To send output to the journal (and therefore Loki/Grafana) instead of the terminal:

```bash
sudo backup-footage 2>&1 | logger -t backup-footage
```

## Adding a new job

Add an entry to `backup_jobs` in `ansible/playbooks/backups.yml`:

```yaml
- name: my-stuff
  src: /mnt/hdd-6tb/my-stuff/
  dest: /mnt/hdd-3tb/my-stuff/
  schedule: "0 5 * * *"
```

Then deploy:

```bash
ansible-playbook playbooks/backups.yml --ask-become-pass --ask-vault-pass
```

## How it works

The `rsync_backup` role templates a shell script and a cron file for each job. Scripts use `rsync -a --delete` so the destination mirrors the source exactly — deleted files on the source get deleted on the destination too.
