# port-scanner

<div align="center">

CLI berbasis Go untuk melakukan pemindaian port TCP pada host tertentu.

<img src="https://img.shields.io/badge/Language-Go-0f172a?style=for-the-badge&logo=go&logoColor=white" alt="Go badge" />
<img src="https://img.shields.io/badge/Type-CLI%20Tool-14b8a6?style=for-the-badge&logo=gnubash&logoColor=white" alt="CLI badge" />
<img src="https://img.shields.io/badge/Focus-Network%20Utility-f59e0b?style=for-the-badge&logo=datadog&logoColor=white" alt="Focus badge" />

</div>

---

## Ringkasan

`port-scanner` dibuat untuk kebutuhan pemindaian port TCP dengan cara yang sederhana, cepat, dan mudah dijalankan langsung dari terminal. Tool ini mendukung daftar port tunggal maupun range port, pemindaian paralel, serta output yang bisa dibaca langsung atau disimpan ke file.

## Fitur

- Mendukung port tunggal dan range port
- Pemindaian paralel dengan jumlah worker yang dapat diatur
- Output `table` dan `json`
- Opsi menyimpan hasil ke file
- Opsi untuk menampilkan hanya port yang terbuka
- Ringkasan hasil pemindaian
- Pengujian dasar untuk parser dan proses scan

## Menjalankan Project

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 22,80,443,8000-8100
```

## Opsi

- `--host` target host atau IP
- `--ports` daftar port, contoh `22,80,443,8000-8100`
- `--timeout` timeout koneksi per port
- `--concurrency` jumlah worker paralel
- `--format` `table` atau `json`
- `--output` simpan hasil ke file
- `--open-only` hanya menampilkan port terbuka

## Contoh Penggunaan

Scan dasar:

```bash
go run ./cmd/port-scanner --host scanme.nmap.org --ports 22,80,443
```

Output JSON:

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 22-30 --format json
```

Simpan hasil ke file:

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 22-30 --format json --output results.json
```

Tampilkan hanya port terbuka:

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 1-1024 --open-only
```
