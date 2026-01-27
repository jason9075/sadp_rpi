# SADP Tool for Raspberry Pi - Agent Development Guide

本文件為專門供 AI Agent（如 opencode）使用的開發規範，旨在確保在 SADP Tool 專案中產出的程式碼符合專案架構、NixOS 環境要求以及 Senior Engineer 的編碼標準。

## 1. 專案概述 (Project Overview)

本專案是一個使用 Go 語言開發的輕量化工具，用於在區域網路（LAN）中發現海康威視（Hikvision）設備。
- **通訊協議:** UDP Multicast
- **組播位址:** `239.255.255.250:37020`
- **目標環境:** Raspberry Pi 4/5 (ARM64) 執行 NixOS。
- **核心功能:** 發送 XML Probe、接收並解析設備回傳之 XML、輸出 JSON/Table 格式。

## 2. 開發環境與自動化 (Development & Automation)

專案嚴格遵循 NixOS 生態與 Makefile 自動化規範。

### 2.1 建置與測試指令
- **進入開發環境:** `nix develop` (由 `flake.nix` 定義)
- **編譯專案:** `make build`
- **執行全域測試:** `make test`
- **執行單一測試 (Single Test):**
  ```bash
  go test -v -run ^TestName$ ./path/to/package
  ```
- **代碼格式化與檢查:** `make lint` (整合 `golangci-lint`)
- **跨平台編譯 (ARM64):** `GOARCH=arm64 GOOS=linux go build -o sadp-rpi`

### 2.2 專案目錄結構
- `/cmd/sadp/`: 進入點 (Main entry point)
- `/pkg/sadp/`: 核心邏輯 (Scanner, Parser, Models)
- `/internal/`: 內部工具與輔助函式
- `/guidelines`: 原始設計需求文檔

## 3. 程式碼風格與規範 (Code Style & Standards)

### 3.1 Golang 慣例
- **Idiomatic Go:** 遵循 [Effective Go](https://golang.org/doc/effective_go) 與 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。
- **強型別 (Strong Typing):** 嚴格定義所有通訊協定相關的 `struct`。
- **並發控制 (Concurrency):** 
  - 必須使用 `context.Context` 傳遞超時與取消信號。
  - 網路掃描需設定 `ReadDeadline` (建議 3-5 秒)。
  - 接收回應需在獨立 Goroutine 執行，並透過 Channel 彙整結果。

### 3.2 命名規則 (Naming)
- **技術術語:** 保留 English (如 `Multicast`, `Payload`, `Interface`)。
- **變數/函式:** 遵循 `camelCase`，匯出者為 `PascalCase`。
- **縮寫:** 縮寫應一致大寫 (如 `JSON`, `UDP`, `IP`, `XML`)。

### 3.3 錯誤處理 (Error Handling)
- **不忽略錯誤:** 嚴禁使用 `_ = function()`，必須顯式處理 `error`。
- **錯誤包裝:** 使用 `fmt.Errorf("context: %w", err)` 提供詳細的錯誤路徑。
- **不輕易 Panic:** 除非是不可恢復的系統初始化錯誤。

### 3.4 匯入規範 (Imports)
匯入區塊應分為三組，並以空行分隔：
1. 標準庫 (Standard Library)
2. 第三方庫 (External Libraries)
3. 本地模組 (Internal Modules)

### 3.5 Git 規範 (Git Standards)
- **提交訊息 (Commit Messages):** 必須使用 Traditional Chinese (Taiwan)。
- **格式:** 遵循 `type: description` 格式 (例如 `feat: 實作 UDP 組播發送邏輯`)。
- **類型:** `feat`, `fix`, `refactor`, `docs`, `test`, `chore` 等。

## 4. 網路與系統集成規範 (Networking & System)

### 4.1 Multicast 處理
- **網卡綁定 (Interface Binding):** 必須支援使用者指定 `Network Interface` (例如 `eth0`)。在 Linux 下，未指定 Interface 發送組播包常會失敗。
- **通訊埠:** 預設綁定 `:0` (OS 分配) 以接收回傳，或提供固定埠號選項。
- **封包結構:** 嚴格遵守海康威視 XML 格式，包含正確的 `<?xml ...?>` 宣告。

### 4.2 NixOS 適配
- **不可變性:** 程式碼不應嘗試修改 `/etc` 等唯讀目錄。
- **防火牆:** 程式啟動時若發現接收失敗，應提示使用者檢查 `networking.firewall` 設定。
- **跨平台:** 確保編譯出的二進位檔能在 ARM64 架構的 Raspberry Pi 上正確運作。

## 5. 測試規範 (Testing Standards)

- **單元測試:** 核心邏輯 (如 XML 解析) 必須具備單元測試。
- **模擬 (Mocking):** 網路通訊部分建議使用 Interface 進行抽象，以便在測試中模擬 UDP Server/Client。
- **效能測試:** 若涉及大量設備解析，應考慮 Benchmark 測試。

## 6. 註解與說明文件 (Documentation)

- **語言:** 所有的邏輯說明、註解、Git Commit Message 應使用 **Traditional Chinese (Taiwan)**。
- **技術名詞:** 保持 **English** 不翻譯（如：Interface, Buffer, Struct, Interface）。
- **說明重點:** 註解應著重於 *Why* (為什麼這樣實作) 而非 *What* (程式碼做了什麼)。

## 7. 範例代碼片段 (Reference Snippets)

### 7.1 發送 Probe 的 XML 結構
```go
type Probe struct {
	XMLName xml.Name `xml:"Probe"`
	Uuid    string   `xml:"Uuid"`
	Types   string   `xml:"Types"`
}
```

### 7.2 指定 Interface 的 UDP 連線
```go
// 範例邏輯：如何透過指定 interface 開啟 UDP 連線
ifi, err := net.InterfaceByName(interfaceName)
if err != nil {
    return fmt.Errorf("獲取 interface %s 失敗: %w", interfaceName, err)
}

// 加入 Multicast Group 或指定發送介面
// ... 具體實作
```

### 7.3 錯誤處理範例
```go
if err := scanner.Scan(ctx); err != nil {
    log.Printf("掃描過程中發生錯誤: %v", err)
    return err
}
```

## 7. 常見任務與工作流 (Common Tasks & Workflows)

### 7.1 新增設備屬性
1. 在 `pkg/sadp/models.go` 中更新 `ProbeMatch` 結構體。
2. 確保 XML tag 與海康協議一致（區分大小寫）。
3. 在單元測試中新增對應的 XML 樣本進行驗證。

### 7.2 實作新的輸出格式
1. 在 `pkg/sadp/printer.go` (建議名稱) 實作對應格式。
2. 確保 `--json` 輸出的欄位名稱維持 English 且符合 camelCase。

### 7.3 調試網路問題
1. 使用 `tcpdump -i <interface> udp port 37020 -A` 觀察封包。
2. 檢查 `net.Interface.Flags` 是否包含 `net.FlagMulticast`。
3. 增加程式內的 Debug Log，詳細記錄發送與接收的 Buffer 長度。

## 8. 參考文件與連結 (References)

- [Hikvision SADP Protocol Analysis (External)](https://github.com/mscandurra/sadp-protocol)
- [NixOS Wiki: Networking](https://nixos.wiki/wiki/Networking)
- [Go Language Specification](https://golang.org/ref/spec)

---
*本文件將隨專案進度持續更新。Agent 在進行大規模改動前，應先閱讀此規範。*
