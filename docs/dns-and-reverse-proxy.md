# Internal DNS and Reverse Proxy

All homelab names live under `lab.taylor-meador.com`. The zone exists only in Unbound on OPNsense and is not published in Cloudflare. DHCP hands OPNsense out as the resolver, so LAN clients resolve these names automatically.

## How it works

- **Web services** (Jellyfin, etc.): every name resolves to the nginx CT (`10.0.10.120`). nginx reads the `Host` header and proxies to the backend `IP:port`. Backend IPs and ports are recorded only in nginx.
- **Non-HTTP services** (NFS, Postgres, PBS, SSH): the name resolves directly to the host that serves it. The client knows the port, so nginx isn't involved.
- **Machines**: each host gets an A record (`srv1`, `srv2`, ...). Mount and job configs should use names, never IPs.

## Add a new web service

1. On the nginx CT, create `/etc/nginx/sites-available/<service>.conf`:

   ```nginx
   server {
       listen 80;
       server_name <service>.lab.taylor-meador.com;

       location / {
           proxy_pass http://BACKEND_IP:PORT;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
           proxy_http_version 1.1;
           proxy_set_header Upgrade $http_upgrade;
           proxy_set_header Connection "upgrade";
       }
   }
   ```

2. Enable, test, and reload:

   ```bash
   ln -s /etc/nginx/sites-available/<service>.conf /etc/nginx/sites-enabled/<service>.conf
   nginx -t
   systemctl reload nginx
   ```

3. Test before touching DNS:

   ```bash
   curl -I -H "Host: <service>.lab.taylor-meador.com" http://10.0.10.120
   ```

4. In OPNsense: Services > Unbound DNS > Overrides. Add a host override (host `<service>`, domain `lab.taylor-meador.com`, IP `10.0.10.120`) and apply.

5. Verify: `dig <service>.lab.taylor-meador.com @<opnsense-ip>` should return `10.0.10.120`.

## Add a non-HTTP name (NFS, DB, backup target)

Add an Unbound host override pointing at the serving host's own IP. Use a role name like `nfs` or `db` so the data can move between machines by changing one record.

NFS mount syntax is `host:/exported/path`:

```
nfs.lab.taylor-meador.com:/mnt/storage  /mnt/storage  nfs  defaults  0  0
```

## If an IP changes

- **Machine IP changed:** edit its Unbound record. For a backend behind nginx, also update the `proxy_pass` line, since it uses the raw IP.
- **Service moved to a different machine:** update `proxy_pass` (web) or repoint the role record (non-HTTP).

## Gotchas

- Statically configured machines (Proxmox hosts, CTs with fixed DNS) need OPNsense as their nameserver. Short names like `jellyfin` also need the search domain `lab.taylor-meador.com` set.
- Devices with hardcoded DNS or browser DNS-over-HTTPS bypass Unbound and won't resolve these names.
- Anything required before DNS is up (the DNS host itself, Proxmox storage at boot) should have an `/etc/hosts` fallback or use an IP.
- `proxy_pass` uses IPs on purpose: nginx resolves names once at startup and won't start if DNS is unreachable.
- Back up OPNsense config (System > Configuration > Backups); it includes the Unbound overrides.
- No TLS yet. Cloudflare DNS-01 with a wildcard cert for `*.lab.taylor-meador.com` is the path if that's wanted later.
