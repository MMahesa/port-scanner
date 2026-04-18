# port-scanner

Utilitas CLI berbasis Go untuk melakukan pemindaian port TCP pada host tertentu.

## Fitur

- Mendukung daftar port tunggal dan range port
- Pemindaian paralel dengan jumlah worker yang bisa diatur
- Output tabel atau JSON
- Opsi menyimpan hasil scan ke file
- Opsi untuk menampilkan hanya port yang terbuka
- Ringkasan hasil pemindaian
- Pengujian dasar untuk parser dan proses scan

## Menjalankan Project

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 22,80,443,8000-8100
```

## Opsi Penting

- `--host` target host atau IP
- `--ports` daftar port, contoh `22,80,443,8000-8100`
- `--timeout` timeout koneksi per port
- `--concurrency` jumlah worker paralel
- `--format` `table` atau `json`
- `--output` simpan hasil ke file
- `--open-only` hanya menampilkan port terbuka

## Contoh Penggunaan

Output tabel:

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

Hanya port terbuka:

```bash
go run ./cmd/port-scanner --host 127.0.0.1 --ports 1-1024 --open-only
```
