# Deploying Foundry to EC2

The whole stack (Postgres + migrations, Go API, Next.js, Caddy for HTTPS) runs
on one EC2 instance with `docker compose`. Caddy is the only thing exposed to
the internet; it gets Let's Encrypt certificates automatically.

```
internet ──443──► caddy ─┬─► frontend:3000   (app.<ip>.sslip.io)
                         └─► api:8080 ──► db:5432   (api.<ip>.sslip.io)
```

Cost: t4g.small ≈ $12.30/mo + public IPv4 $3.65 + 20GB gp3 $1.60 ≈ **$18/mo**.

---

## 1. Launch the instance (AWS console)

EC2 → **Launch instance**:

| Setting | Value | Why |
|---|---|---|
| AMI | **Ubuntu Server 24.04 LTS, 64-bit (Arm)** | t4g is ARM (Graviton). x86 AMIs won't appear. |
| Instance type | **t4g.small** (2 vCPU, 2GB) | `next build` needs ~1.5GB; micro (1GB) will OOM. |
| Key pair | Create new, ED25519, `.pem` | Your SSH key. Downloaded once; keep it. |
| Storage | **20 GiB gp3** | Docker images + DB + backups. Default 8GB fills fast. |

**Network settings → Edit → Create security group** (a firewall in front of the instance):

| Type | Port | Source | Why |
|---|---|---|---|
| SSH | 22 | **My IP** | Only you can SSH in. Never 0.0.0.0/0. |
| HTTP | 80 | Anywhere | Let's Encrypt challenge + redirect to https |
| HTTPS | 443 | Anywhere | The app |
| Custom UDP | 443 | Anywhere | HTTP/3 (optional) |

Postgres (5432), the API (8080) and Next (3000) are **not** opened: compose binds
them to `127.0.0.1` on the box, and the security group blocks them anyway.

Then:
- **Elastic IP**: EC2 → Elastic IPs → Allocate → Associate with the instance.
  Without it the public IP changes on every stop/start, and your URLs
  (which contain the IP) would break.
- Your key: `chmod 400 ~/Downloads/foundry.pem` (SSH refuses keys others can read).

```bash
ssh -i ~/Downloads/foundry.pem ubuntu@<ELASTIC_IP>
```

## 2. Prepare the box (on EC2)

```bash
# Updates
sudo apt update && sudo apt upgrade -y

# 2GB swap: a safety net so `next build` doesn't get OOM-killed on 2GB RAM
sudo fallocate -l 2G /swapfile && sudo chmod 600 /swapfile
sudo mkswap /swapfile && sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab

# Docker Engine + compose plugin (Docker's official install script)
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker ubuntu     # run docker without sudo
exit                               # log out and back in for the group to apply
```

Reconnect, then check: `docker run --rm hello-world` and `docker compose version`.

## 3. Get the code and secrets onto the box

```bash
# on EC2
git clone https://github.com/Adam-Omarali/foundry.git ~/foundry
```
(Private repo: create a GitHub fine-grained token with read access to this repo,
or add a deploy key, and clone with that.)

Secrets never go through git. Copy `.env` from your Mac, then edit it on the box:
```bash
# on your Mac, from the repo root
scp -i ~/Downloads/foundry.pem .env ubuntu@<ELASTIC_IP>:~/foundry/.env
```
```bash
# on EC2
cd ~/foundry && chmod 600 .env && nano .env
```
Production changes in `.env` (IP with dashes, e.g. `3-14-15-92`):
```bash
COMPOSE_PROFILES=prod
PUBLIC_APP_URL=https://app.<ip-with-dashes>.sslip.io
PUBLIC_API_URL=https://api.<ip-with-dashes>.sslip.io
LOG_LEVEL=info
```
Use a **new** `DB_PASSWORD`, `JWT_SECRET` and `NEXTAUTH_SECRET` for production
(`openssl rand -hex 32`); don't reuse your local ones.

**sslip.io** is free wildcard DNS: `app.3-14-15-92.sslip.io` resolves to
`3.14.15.92`. No domain to buy; swap in a real domain later by changing these
two URLs (and the Google redirect URI).

## 4. Google OAuth (Cloud console)

On the web client, add to **Authorized redirect URIs**:
```
https://app.<ip-with-dashes>.sslip.io/api/auth/callback/google
```

## 5. Start it

```bash
cd ~/foundry
docker compose up -d --build      # first build takes a few minutes on a t4g.small
docker compose ps
docker compose logs -f caddy      # watch for "certificate obtained successfully"
```

Check:
```bash
curl -i https://api.<ip-with-dashes>.sslip.io/api/verifytoken    # → 401 = API reachable over HTTPS
```
Then open `https://app.<ip-with-dashes>.sslip.io` and sign in.

## 6. Point the Chrome extension at production

In `chrome-extension/auth.js`: `DEV_MODE = false` and set the production
`API_URL` to `PUBLIC_API_URL`. In `manifest.json`, replace the railway URL in
`content_security_policy.extension_pages` `connect-src` with the same URL.
Reload the extension in `chrome://extensions`.

## 7. Nightly backups

```bash
chmod +x ~/foundry/deploy/backup.sh
~/foundry/deploy/backup.sh          # run once by hand to check it works
crontab -e
```
Add:
```
0 3 * * * /home/ubuntu/foundry/deploy/backup.sh >> /home/ubuntu/backups/backup.log 2>&1
```
Backups live on the same disk as the database, so they protect against bad
migrations and mistakes, not a lost instance. For that, also copy them off the
box (e.g. `aws s3 cp` to a bucket, or scp to your Mac) or enable EBS snapshots
(EC2 → Lifecycle Manager).

Restore:
```bash
docker compose exec -T db sh -c 'pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists' < ~/backups/foundry-YYYY-MM-DD.dump
```

## Redeploying

```bash
ssh -i ~/Downloads/foundry.pem ubuntu@<ELASTIC_IP>
cd ~/foundry && git pull && docker compose up -d --build
```
Only images whose code changed rebuild; only changed containers restart.
`migrate` runs on every `up` and applies just the new migrations.

## Troubleshooting

| Symptom | Look at |
|---|---|
| Site doesn't load at all | Security group 80/443; `docker compose ps`; `docker compose logs caddy` |
| Caddy cert errors | Port 80 must be open; URL must contain the *current* Elastic IP |
| `redirect_uri_mismatch` | Google redirect URI must match `PUBLIC_APP_URL` exactly |
| Build killed / exits 137 | Out of memory: check swap is on (`free -h`) |
| API restarting | `docker compose logs api --tail 50` |
| Disk full | `df -h`; `docker system prune` (removes unused images, keeps volumes) |

## Later: build in CI instead of on the box

Building on a t4g.small works but is slow and uses its RAM. The next step is a
GitHub Actions workflow that builds ARM64 images (`docker buildx --platform
linux/arm64`), pushes them to GHCR, then SSHes in and runs
`docker compose pull && docker compose up -d`. Compose would use `image:`
instead of `build:` for api/frontend.
