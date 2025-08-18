# HeheSwitch Deployment (Single Board Computer)

Panduan cepat menjalankan service ini otomatis saat boot di SBC (Raspberry Pi / Orange Pi / Jetson / dsb) menggunakan systemd.

## 1. Build Binary Multi-Arch (opsional dari dev machine)

Host x86 build untuk ARM64:

```
make build-arm64
```

Hasil: `dist/heheswitch-linux-arm64`

Atau build langsung di SBC:
```
make build
```

## 2. Buat User Service

Login ke SBC lalu:
```
sudo useradd -r -s /usr/sbin/nologin heheswitch || true
sudo mkdir -p /opt/heheswitch
sudo chown heheswitch:heheswitch /opt/heheswitch
```

## 3. Copy Binary & Database (jika diperlukan)

Dari dev machine (ganti HOST):
```
scp dist/heheswitch-linux-arm64 user@HOST:/tmp/heheswitch
ssh user@HOST 'sudo mv /tmp/heheswitch /opt/heheswitch/heheswitch && sudo chown heheswitch:heheswitch /opt/heheswitch/heheswitch && sudo chmod 755 /opt/heheswitch/heheswitch'
```

Jika ingin mulai fresh, database akan otomatis dibuat (`heheswitch.db`). Jika mau migrasi file DB lama:
```
scp heheswitch.db user@HOST:/tmp/heheswitch.db
ssh user@HOST 'sudo mv /tmp/heheswitch.db /opt/heheswitch/heheswitch.db && sudo chown heheswitch:heheswitch /opt/heheswitch/heheswitch.db'
```

## 4. Systemd Service

Copy contoh service:
```
scp deploy/heheswitch.service user@HOST:/tmp/
ssh user@HOST 'sudo mv /tmp/heheswitch.service /etc/systemd/system/heheswitch.service'
```

Reload & enable:
```
ssh user@HOST 'sudo systemctl daemon-reload && sudo systemctl enable --now heheswitch'
```

Cek status:
```
ssh user@HOST 'systemctl status heheswitch --no-pager'
```

Logs live:
```
ssh user@HOST 'journalctl -u heheswitch -f -n 100'
```

## 5. Ubah Port

Aplikasi membaca port dari env `PORT` (default 8080).

Cara override tanpa edit file asli:
```
sudo systemctl edit heheswitch
```
Tambah:
```
[Service]
Environment=PORT=9090
```
Simpan, lalu:
```
sudo systemctl daemon-reload
sudo systemctl restart heheswitch
```

Verifikasi:
```
journalctl -u heheswitch -n 20 | grep listening
```

## 6. MQTT (opsi)

Dua opsi konfigurasi:

1. Via UI (disarankan): Set broker, prefix, user, password di modal. Tersimpan di DB dan akan auto-reconnect saat restart.
2. Via environment variables jika mau statis:
   Tambahkan di override systemd:
   ```
   [Service]
   Environment=MQTT_BROKER=tcp://192.168.1.10:1883
   Environment=MQTT_USERNAME=youruser
   Environment=MQTT_PASSWORD=yourpass
   ```
   Restart service.

## 7. Update Versi

Saat rilis baru:
```
make build-arm64
scp dist/heheswitch-linux-arm64 user@HOST:/tmp/heheswitch
ssh user@HOST 'sudo systemctl stop heheswitch && sudo mv /tmp/heheswitch /opt/heheswitch/heheswitch && sudo chown heheswitch:heheswitch /opt/heheswitch/heheswitch && sudo systemctl start heheswitch'
```

## 8. Keamanan Tambahan (opsional)

- Reverse proxy (nginx / caddy) untuk TLS.
- Batasi firewall hanya port HTTP yang diperlukan.
- Simpan credential MQTT di override file (permission root) bukan di DB kalau sensitif.

## 9. Troubleshooting

| Masalah | Solusi Singkat |
|---------|----------------|
| Port tidak berubah | Pastikan override systemd benar dan lakukan daemon-reload + restart |
| MQTT tidak connect | Cek logs journal untuk "MQTT" string; pastikan broker reachable dari SBC |
| DB corrupt | Hentikan service, backup file, hapus lock file *.db-wal / *.db-shm bila ada |
| UI tidak update realtime | Cek koneksi WebSocket (DevTools) dan pastikan tidak ada proxy memutus connection |

## 10. Uninstall
```
sudo systemctl disable --now heheswitch
sudo rm /etc/systemd/system/heheswitch.service
sudo systemctl daemon-reload
sudo rm -rf /opt/heheswitch
sudo userdel heheswitch
```

Selesai.
