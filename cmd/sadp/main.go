package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/jason9075/sadp-rpi/pkg/sadp"
)

func main() {
	var ifaceName string
	flag.StringVar(&ifaceName, "iface", "eth0", "指定要掃描的網路介面，逗號分隔或 all (例如 eth0, wlan0, eth0,wlan0, all)")
	flag.StringVar(&ifaceName, "i", "eth0", "指定要掃描的網路介面 (簡寫)")
	timeout := flag.Duration("timeout", 3*time.Second, "掃描超時時間")
	asJSON := flag.Bool("json", false, "輸出 JSON 格式")
	flag.Parse()

	if ifaceName == "" {
		fmt.Println("錯誤: 必須指定網路介面 (-i 或 -iface)")
		flag.Usage()
		os.Exit(1)
	}

	// 解析介面清單：支援 "all" 或逗號分隔
	var ifaceNames []string
	if ifaceName == "all" {
		ifaces, err := net.Interfaces()
		if err != nil {
			log.Fatalf("無法列舉網路介面: %v", err)
		}
		for _, iface := range ifaces {
			// 只選擇已啟用且支援 Multicast 的介面
			if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagMulticast != 0 {
				ifaceNames = append(ifaceNames, iface.Name)
			}
		}
		if len(ifaceNames) == 0 {
			log.Fatal("找不到支援 Multicast 的網路介面")
		}
	} else {
		for _, name := range strings.Split(ifaceName, ",") {
			if trimmed := strings.TrimSpace(name); trimmed != "" {
				ifaceNames = append(ifaceNames, trimmed)
			}
		}
	}

	// 並發掃描所有指定介面，以 MAC 去重
	var (
		mu         sync.Mutex
		allDevices []sadp.Device
		seen       = make(map[string]bool)
		wg         sync.WaitGroup
	)

	for _, name := range ifaceNames {
		wg.Add(1)
		go func(ifName string) {
			defer wg.Done()
			scanner := sadp.NewScanner(ifName, *timeout)
			ctx, cancel := context.WithTimeout(context.Background(), *timeout+time.Second)
			defer cancel()

			fmt.Fprintf(os.Stderr, "正在介面 %s 上啟動 SADP 掃描 (超時: %v)...\n", ifName, *timeout)
			devices, err := scanner.Scan(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "介面 %s 掃描失敗: %v\n", ifName, err)
				return
			}

			mu.Lock()
			for _, d := range devices {
				if !seen[d.MAC] {
					seen[d.MAC] = true
					allDevices = append(allDevices, d)
				}
			}
			mu.Unlock()
		}(name)
	}
	wg.Wait()

	if *asJSON {
		output, _ := json.MarshalIndent(allDevices, "", "  ")
		fmt.Println(string(output))
		return
	}

	printTable(allDevices)
}

func printTable(devices []sadp.Device) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "IP ADDRESS\tPORT\tMAC ADDRESS\tVERSION\tSTATUS\tSERIAL NUMBER")
	for _, d := range devices {
		status := "Inactive"
		if d.Activated {
			status = "Active"
		}
		fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n",
			d.IPv4Address, d.HttpPort, d.MAC, d.SoftwareVersion, status, d.DeviceSN)
	}
	w.Flush()
}
