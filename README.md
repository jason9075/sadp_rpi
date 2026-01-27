# SADP Tool for Raspberry Pi

這是一個用 Go 語言編寫的輕量化工具，專門用於在區域網路中發現海康威視 (Hikvision) 設備。

## 功能
- 支援 UDP Multicast (239.255.255.250:37020)
- 自動解析設備回傳的 XML 資訊
- 支援指定網路介面 (使用 `golang.org/x/net/ipv4` 精確控制)
- 提供人性化表格與 JSON 輸出格式
- **不需要 Root 權限**即可運行（需注意防火牆與權限設定）

## 安裝與建置

本專案支援 Nix Flake。

### 使用 Nix 進入開發環境
```bash
nix develop
```

### 編譯
```bash
make build
```

### 跨平台編譯 (ARM64 for Raspberry Pi)
```bash
make build-arm64
```

## CI/CD
本專案包含 GitHub Actions 工作流 (`.github/workflows/build.yml`)：
- **自動發佈**：當您推送以 `v` 開頭的 Tag（如 `v0.1.0`）時，會自動觸發建置並建立 GitHub Release。
- **支援平台**：自動編譯包含 Linux (amd64, arm64) 與 macOS (Intel, Apple Silicon) 的二進位檔。

## 使用說明

### 基本掃描
指定您的網路介面（例如 `eth0`）：
```bash
./sadp-rpi -i eth0
```

### 輸出 JSON 格式
```bash
./sadp-rpi -i eth0 -json
```

### 參數說明
- `-i`, `-iface`: (必填，預設為 eth0) 指定要使用的網卡介面，例如 `eth0`, `wlan0` 或 `enp3s0`。
- `-timeout`: 掃描持續時間，預設為 `3s`。
- `-json`: 以 JSON 格式輸出結果。

## 權限與防火牆 (NixOS)

### 1. 權限設定
雖然本工具使用隨機高位埠發送組播，但在某些系統下發送組播包可能仍受限。如果無法發現設備，可以嘗試：
- 使用 `sudo` 運行。
- 或者為二進位檔賦予網路權限（非 Root 方案）：
  ```bash
  sudo setcap cap_net_raw,cap_net_admin=eip ./sadp-rpi
  ```

### 2. 防火牆 (NixOS)
如果無法接收到設備回傳的 XML，請確保防火牆允許 `37020` 埠的 UDP 流量。
在 `configuration.nix` 中加入：
```nix
networking.firewall.allowedUDPPorts = [ 37020 ];
```

## License
MIT
