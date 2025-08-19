# HeheSwitch - IoT PlayStation Rental Management System

Sistem manajemen rental PlayStation berbasis IoT dengan kontrol relay real-time untuk mengelola sesi rental dan kontrol daya konsol PlayStation. Dibangun dengan Go, Fiber, SQLite, dan MQTT untuk integrasi IoT yang seamless.

## Overview

HeheSwitch adalah solusi lengkap untuk mengelola bisnis rental PlayStation yang terintegrasi dengan perangkat IoT. Sistem ini memungkinkan operator untuk mengontrol daya konsol secara remote, mengelola sesi rental, tracking transaksi, dan monitoring status konsol secara real-time melalui web interface yang responsif.

## Technology Stack

- **Backend**: Go 1.24+ dengan Fiber web framework
- **Database**: SQLite dengan optimasi performa WAL mode
- **Frontend**: HTML5, CSS3, JavaScript (Vanilla) dengan WebSocket
- **IoT Communication**: MQTT protocol untuk kontrol device
- **Real-time**: WebSocket untuk update status live
- **Authentication**: Session-based auth dengan bcrypt
- **Deployment**: Systemd service untuk SBC (Raspberry Pi, Orange Pi, dll)

## Features

### Core Features
- **Sesi Management**: Start, extend, dan stop sesi rental dengan timer otomatis
- **Auto Stop**: Background ticker yang otomatis menghentikan sesi saat waktu habis
- **Real-time Dashboard**: Web interface dengan update live status via WebSocket
- **User Authentication**: Multi-level user access (admin/user) dengan session management
- **Transaction Tracking**: Pencatatan lengkap riwayat transaksi per konsol
- **Dynamic Pricing**: Pengaturan harga per jam yang dapat diubah secara real-time

### IoT Integration
- **MQTT Control**: Kontrol relay device melalui MQTT protocol
- **Device Status Monitoring**: Monitoring status device dengan feedback real-time
- **Fallback System**: Mock in-memory controller saat MQTT tidak tersedia
- **Topic Conventions**: Standar topic MQTT untuk konsistensi integrasi

### Technical Features
- **SQLite Embedded**: Database lokal dengan optimasi performa (WAL mode)
- **Clean Architecture**: Struktur kode modular dengan separation of concerns
- **WebSocket Live Updates**: Update status konsol secara real-time tanpa refresh
- **Responsive UI**: Interface yang optimal untuk desktop dan mobile
- **Dark Mode**: Theme switcher untuk kenyamanan penggunaan

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | /login   | User authentication |
| POST   | /logout  | Logout current session |
| GET    | /me      | Get current user info |

### Console Management (User/Admin)
| Method | Endpoint | Body | Description |
|--------|----------|------|-------------|
| POST | /start | `{console_id, duration_minutes}` | Mulai sesi rental |
| POST | /extend | `{console_id, add_minutes}` | Tambah durasi sesi |
| POST | /stop | `{console_id}` | Stop sesi manual |
| GET | /status | - | Status semua konsol (real-time) |
| GET | /transactions/:console_id | - | Riwayat transaksi per konsol |

### Admin Only
| Method | Endpoint | Body | Description |
|--------|----------|------|-------------|
| GET | /users | - | List semua user |
| POST | /users | `{username, password, role}` | Buat user baru |
| DELETE | /users/:id | - | Hapus user |
| POST | /price | `{console_id, price_per_hour}` | Update harga per jam |

### MQTT Status
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /mqtt/status | Status koneksi MQTT |
| POST | /mqtt/config | Update konfigurasi MQTT |

### WebSocket
- **Endpoint**: `/ws`
- **Purpose**: Real-time status updates
- **Format**: JSON dengan status semua konsol

## System Requirements

- **Go**: Version 1.24 atau lebih baru
- **OS**: Linux, macOS, atau Windows
- **Memory**: Minimal 64MB RAM
- **Storage**: Minimal 50MB free space
- **Network**: Port 8080 (default, dapat dikonfigurasi)
- **MQTT Broker**: Optional, untuk integrasi IoT (Mosquitto, HiveMQ, dll)

## Quick Start

### 1. Installation & Build

```bash
# Clone repository
git clone https://github.com/izzamoe/switchaja.git
cd switchaja

# Install dependencies
go mod tidy

# Build application
make build
# atau untuk ARM64: make build-arm64
```

### 2. Configuration

Aplikasi menggunakan environment variables untuk konfigurasi:

```bash
# Server Configuration
export PORT=8080                    # Server port (default: 8080)
export DB_PATH=heheswitch.db        # Database file path (default: heheswitch.db)
export SQLITE_MODE=balanced         # SQLite mode: aggressive|balanced|safe

# MQTT Configuration (Optional)
export MQTT_BROKER=tcp://localhost:1883
export MQTT_PREFIX=ps               # Topic prefix (default: ps)
export MQTT_USERNAME=your_username  # Optional
export MQTT_PASSWORD=your_password  # Optional
export MQTT_CLIENT_ID=heheswitch    # Optional
```

### 3. Run Application

```bash
# Development mode
go run ./cmd/server

# Production mode (after build)
./dist/heheswitch
```

Buka browser dan akses: `http://localhost:8080`

### 4. Default Login

- **Username**: `admin`
- **Password**: `admin123`

⚠️ **Penting**: Ubah password default setelah login pertama!

## MQTT Integration (IoT)

HeheSwitch mendukung integrasi IoT melalui MQTT protocol untuk mengontrol relay device secara remote.

### Setup MQTT

#### Opsi 1: Environment Variables (Legacy)
```bash
export MQTT_BROKER=tcp://localhost:1883
export MQTT_PREFIX=ps        # optional (default: ps)
export MQTT_USERNAME=...     # optional
export MQTT_PASSWORD=...     # optional
export MQTT_CLIENT_ID=...    # optional
```

#### Opsi 2: Database Configuration (Recommended)
Gunakan web interface untuk mengatur MQTT:
1. Login sebagai admin
2. Klik tombol "MQTT" di header
3. Isi konfigurasi MQTT broker
4. Konfigurasi akan tersimpan di database dan auto-reconnect saat restart

### MQTT Topic Convention

| Topic Pattern | Direction | Payload | Description |
|---------------|-----------|---------|-------------|
| `{prefix}/{console_id}/cmd` | Publish (Server → Device) | `ON` / `OFF` | Perintah kontrol relay |
| `{prefix}/{console_id}/status` | Subscribe (Device → Server) | Free format | Status feedback dari device |

**Contoh**:
- Command: `ps/1/cmd` dengan payload `ON` (nyalakan konsol 1)
- Status: `ps/1/status` dengan payload `relay_on` (feedback dari device)

### Fallback System

Jika MQTT tidak tersedia atau gagal connect:
- Sistem otomatis fallback ke **Mock Controller**
- Semua perintah tetap berfungsi (log ke console)
- UI tetap menampilkan status normal
- Tidak ada downtime pada sistem rental

### Device Implementation

Device IoT harus subscribe ke topic `{prefix}/+/cmd` dan respond dengan format:
```cpp
// Arduino/ESP32 example
void onMqttMessage(char* topic, byte* payload, unsigned int length) {
    String command = String((char*)payload).substring(0, length);
    if (command == "ON") {
        digitalWrite(RELAY_PIN, HIGH);
        mqttClient.publish("ps/1/status", "relay_on");
    } else if (command == "OFF") {
        digitalWrite(RELAY_PIN, LOW);
        mqttClient.publish("ps/1/status", "relay_off");
    }
}
```

## User Management

### User Roles
- **Admin**: Full access (user management, pricing, konsol management)
- **User**: Limited access (hanya konsol management)

### Default Admin
- Username: `admin`
- Password: `admin123`
- Role: `admin`

⚠️ **Security**: Segera ubah password default setelah instalasi!

### Membuat User Baru
```bash
# Via API (sebagai admin)
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"username":"operator","password":"password123","role":"user"}'
```

## Deployment

### Development
```bash
go run ./cmd/server
```

### Production - Manual
```bash
make build
./dist/heheswitch
```

### Production - Single Board Computer (SBC)
Untuk deployment di Raspberry Pi, Orange Pi, atau SBC lainnya, lihat panduan lengkap di [README_DEPLOY.md](README_DEPLOY.md) yang mencakup:
- Build multi-architecture
- Systemd service setup
- User dan permission management
- Port configuration
- Update procedure
- Security hardening
- Troubleshooting

## Database

### SQLite Optimization Modes
```bash
export SQLITE_MODE=aggressive  # Maximum speed, risk data loss
export SQLITE_MODE=balanced    # Good speed, reasonable safety (default)
export SQLITE_MODE=safe        # Maximum durability
```

### Database Schema
- **consoles**: Console information dan status
- **users**: User authentication dan roles
- **transactions**: Rental transaction history
- **mqtt_config**: MQTT configuration storage

### Backup
```bash
# Backup database
cp heheswitch.db heheswitch.db.backup

# Restore
cp heheswitch.db.backup heheswitch.db
```

## Configuration Options

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DB_PATH` | `heheswitch.db` | Database file path |
| `SQLITE_MODE` | `balanced` | SQLite performance mode |
| `MQTT_BROKER` | - | MQTT broker URL |
| `MQTT_PREFIX` | `ps` | MQTT topic prefix |
| `MQTT_USERNAME` | - | MQTT authentication |
| `MQTT_PASSWORD` | - | MQTT authentication |
| `MQTT_CLIENT_ID` | - | MQTT client identifier |

### Application Defaults
- **Console Count**: 5 konsol
- **Default Price**: Rp 40.000/jam
- **Session Timeout**: Auto-stop saat waktu habis
- **WebSocket**: Real-time updates setiap detik

## Troubleshooting

### Common Issues

| Problem | Solution |
|---------|----------|
| Port already in use | Ubah `PORT` environment variable |
| Database locked | Stop aplikasi, hapus file `*.db-wal` dan `*.db-shm` |
| MQTT connection failed | Cek broker URL, credentials, dan network connectivity |
| WebSocket disconnected | Refresh browser, cek network stability |
| Session tidak auto-stop | Restart aplikasi, cek background ticker di logs |

### Logs
```bash
# Development
go run ./cmd/server

# Production
journalctl -u heheswitch -f  # jika pakai systemd
./dist/heheswitch            # manual run
```

### Debug Mode
```bash
# Enable detailed logging
export LOG_LEVEL=debug
go run ./cmd/server
```

## Contributing

1. Fork repository
2. Create feature branch: `git checkout -b feature/nama-fitur`
3. Commit changes: `git commit -am 'Add fitur baru'`
4. Push branch: `git push origin feature/nama-fitur`
5. Submit Pull Request

### Development Setup
```bash
git clone https://github.com/izzamoe/switchaja.git
cd switchaja
go mod tidy
make build
```

## License

Project ini menggunakan lisensi yang belum ditentukan. Silakan hubungi maintainer untuk informasi lisensi.

## Support

- **Issues**: [GitHub Issues](https://github.com/izzamoe/switchaja/issues)
- **Discussions**: [GitHub Discussions](https://github.com/izzamoe/switchaja/discussions)
- **Wiki**: [Project Wiki](https://github.com/izzamoe/switchaja/wiki)

---

**HeheSwitch** - Making PlayStation rental management simple and connected! 🎮⚡
