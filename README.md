# Rental PS IoT (Go + Fiber + SQLite)

Sistem sederhana rental PlayStation dengan kontrol relay (mock) untuk ON/OFF listrik per konsol.

## Fitur
- Start / Extend / Stop sesi
- Auto stop saat waktu habis (background ticker)
- REST API + Admin Web (HTML/JS sederhana)
- SQLite embedded

## Endpoints
| Method | Path | Body | Deskripsi |
|--------|------|------|-----------|
| POST | /start | {console_id, duration_minutes} | Mulai sewa |
| POST | /extend | {console_id, add_minutes} | Tambah durasi |
| POST | /stop | {console_id} | Berhenti manual |
| GET | /status | - | Status semua konsol |

## Jalankan
```
go mod tidy
go run ./cmd/server
```
Buka http://localhost:8080

## Integrasi MQTT (IoT)
Set env var berikut agar server otomatis pakai MQTT:

```
export MQTT_BROKER=tcp://localhost:1883
export MQTT_PREFIX=ps        # optional (default ps)
export MQTT_USERNAME=...     # optional
export MQTT_PASSWORD=...     # optional
```
Kemudian jalankan server.

Topic Konvensi:
- Publish perintah: `<prefix>/<id>/cmd` (payload: `ON` / `OFF`)
- Device publish status (opsional): `<prefix>/<id>/status` (payload bebas, dicetak di log)

Kalau variabel `MQTT_BROKER` tidak ada, sistem fallback ke mock in-memory.
