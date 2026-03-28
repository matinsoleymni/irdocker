# irdocker 🐳

یه ابزار خط فرمان (CLI) ساده برای بررسی اینکه یه ایمیج داکر روی میرورهای ایرانی موجوده یا نه — سریع و همزمان.

---

## نصب

```bash
git clone https://github.com/matinsoleymni/irdocker
cd irdocker
chmod +x install.sh && ./install.sh
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
- پشتیبانی از **Docker Registry v2 Auth** (نتایج دقیق)
- نمایش **دقیق خطاهای شبکه** (timeout، DNS، TLS و...)
- قابلیت **افزودن میرور دلخواه**
- تنظیمات ذخیره‌شده در `~/.irdocker.json`

---

Developed with love and Claude 4.6 by MatinSoleymani
