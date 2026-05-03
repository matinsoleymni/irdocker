# irdocker 🐳

یه ابزار خط فرمان (CLI) ساده برای بررسی اینکه یه ایمیج داکر روی میرورهای ایرانی موجوده یا نه — سریع و همزمان.

---

## نصب

```bash
git clone https://github.com/mamahoos/irdocker
cd irdocker
sudo ./install.sh
```

یا دستی:

```bash
go build -o irdocker .
sudo mv irdocker /usr/local/bin/
```

> نیاز به Go 1.21 یا بالاتر داری.

---

## استفاده

```bash
# بررسی یه ایمیج روی تمام میرورها
irdocker nginx
irdocker nginx:1.25-alpine
irdocker gitea/gitea:latest

# میرور کردن تمام ایمیج‌های یه فایل docker-compose
irdocker docker-compose.yaml

# نمایش لیست میرورها
irdocker list

# اضافه کردن میرور جدید
irdocker add RunFlare mirror-docker.runflare.com

# حذف یه میرور
irdocker remove focker.ir

# برگشت به میرورهای پیش‌فرض
irdocker reset
```

---

## نمونه خروجی

**بررسی ایمیج:**

```
🔍 Checking image: library/nginx:latest
📦 Checking 5 registries...

✅ ArvanCloud         → FOUND
   docker pull docker.arvancloud.ir/nginx:latest

✅ Focker.ir          → FOUND
   docker pull focker.ir/nginx:latest

❌ MobinHost          → NOT FOUND

⏱️  Kernel.ir          → TIMEOUT     (connection timed out)

🔌 Pardisco           → NET ERROR   (DNS lookup failed)

────────────────────────────────────────────────────
Result: 2 found, 1 not found, 2 error(s)
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

### معنی آیکون‌ها

| آیکون | معنی |
|-------|------|
| ✅ | ایمیج پیدا شد — دستور pull نمایش داده می‌شه |
| ❌ | میرور در دسترسه ولی ایمیج وجود نداره |
| ⏱️ | اتصال timeout شد |
| 🔌 | خطای شبکه (DNS، TLS، connection refused و...) |
| ⚠️ | نامشخص (نیاز به احراز هویت یا پاسخ غیرمنتظره) |

---

## میرورهای پیش‌فرض

| Name             | Host                      |
|------------------|---------------------------|
| ArvanCloud       | docker.arvancloud.ir      |
| MobinHost        | docker.mobinhost.com      |
| Pardisco         | mirrors.pardisco.co       |
| Focker.ir        | focker.ir                 |
| Kernel.ir        | docker.kernel.ir          |
| Megan.ir         | hub.megan.ir              |
| Atlantiscloud.ir | hub.atlantiscloud.ir      |
| Iranserver.com   | docker.iranserver.com     |
| Liara.ir         | docker-mirror.liara.ir    |

میرورهای اضافه‌شده توسط کاربر در فایل `~/.irdocker.json` ذخیره می‌شن.

---

## ویژگی‌ها

- بررسی **همزمان** تمام میرورها (سریع)
- پشتیبانی از **docker-compose** — میرور کردن خودکار همه ایمیج‌ها و تولید فایل جدید
- پشتیبانی از **Docker Registry v2 Auth** (نتایج دقیق)
- نمایش **دقیق خطاهای شبکه** (timeout، DNS، TLS و...)
- قابلیت **افزودن میرور دلخواه**
- تنظیمات ذخیره‌شده در `~/.irdocker.json`

---

Developed with love and Claude 4.6 by MatinSoleymani

Maintained by [mamahoos](https://github.com/mamahoos)