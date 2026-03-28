# irdocker

Check Iranian Docker mirror registries for image availability — instantly.

[دایکومنت فارسی](./Readme.fa.md)

## Install

```bash
git clone https://github.com/matinsoleymni/irdocker
cd irdocker
chmod +x install.sh && ./install.sh
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
