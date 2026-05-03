# irdocker

Check Iranian Docker mirror registries for image availability — instantly.

[دایکومنت فارسی](./README.fa.md)


## Install

```bash
git clone https://github.com/mamahoos/irdocker
cd irdocker
sudo ./install.sh
```

Or manually:

```bash
go build -o irdocker .
sudo mv irdocker /usr/local/bin/
```

## Usage

```bash
# Check an image across all registries
irdocker nginx
irdocker nginx:1.25-alpine
irdocker gitea/gitea:latest

# Explicit check subcommand
irdocker check redis:7

# Mirror all images in a docker-compose file
irdocker docker-compose.yaml

# List all configured registries
irdocker list

# Add a new registry
irdocker add RunFlare mirror-docker.runflare.com

# Remove a registry
irdocker remove focker.ir

# Reset to built-in defaults
irdocker reset
```

## Example Output

**Image check:**

```
🔍 Checking image: library/nginx:latest
📦 Checking 5 registries...

✅ ArvanCloud         → FOUND
   docker pull docker.arvancloud.ir/nginx:latest

✅ Kernel.ir          → FOUND
   docker pull docker.kernel.ir/nginx:latest

❌ Focker.ir          → NOT FOUND

❌ MobinHost          → NOT FOUND

❌ Pardisco           → NOT FOUND

──────────────────────────────────────────────────
Result: 2 found, 3 not available
```

**Docker Compose:**

```
🐳 Docker Compose: docker-compose.yaml
📦 Found 3 unique image(s), checking 9 registries...

📋 Image Mirror Report:

    Image               Registry              Mirrored Image
  ───────────────────────────────────────────────────────────────────────
  ✅ nginx:latest        Focker.ir             focker.ir/nginx:latest
  ✅ postgres:15         Focker.ir             focker.ir/postgres:15
  ✅ redis:7-alpine      Focker.ir             focker.ir/redis:7-alpine

  3/3 images mirrored → wrote docker-compose-mirrored.yaml

🔧 Apply changes:

  mv docker-compose.yaml docker-compose.old.yaml
  mv docker-compose-mirrored.yaml docker-compose.yaml
  docker compose up -d
```

## Default Registries

| Name       | Host                      |
|------------|---------------------------|
| ArvanCloud       | docker.arvancloud.ir      |
| MobinHost        | docker.mobinhost.com      |
| Pardisco         | mirrors.pardisco.co       |
| Focker.ir        | focker.ir                 |
| Kernel.ir        | docker.kernel.ir          |
| Megan.ir         | hub.megan.ir              |
| Atlantiscloud.ir | hub.atlantiscloud.ir      |
| Iranserver.com   | docker.iranserver.com     |
| Liara.ir         | docker-mirror.liara.ir    |


Custom registries are saved to `~/.irdocker.json`.

Developed with love and Claude 4.6 by MatinSoleymani

---
Maintained by [mamahoos](https://github.com/mamahoos)
