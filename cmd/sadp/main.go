package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/jason9075/sadp-rpi/pkg/sadp"
)

func main() {
	var ifaceName string
	flag.StringVar(&ifaceName, "iface", "eth0", "指定要掃描的網路介面 (例如 eth0, wlan0)")
	flag.StringVar(&ifaceName, "i", "eth0", "指定要掃描的網路介面 (簡寫)")
	timeout := flag.Duration("timeout", 3*time.Second, "掃描超時時間")
	asJSON := flag.Bool("json", false, "輸出 JSON 格式")
	flag.Parse()

	if ifaceName == "" {
		fmt.Println("錯誤: 必須指定網路介面 (-i 或 -iface)")
		flag.Usage()
		os.Exit(1)
	}

	scanner := sadp.NewScanner(ifaceName, *timeout)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout+time.Second)
	defer cancel()

	fmt.Fprintf(os.Stderr, "正在介面 %s 上啟動 SADP 掃描 (超時: %v)...\n", ifaceName, *timeout)
	devices, err := scanner.Scan(ctx)
	if err != nil {
		log.Fatalf("掃描失敗: %v", err)
	}

	if *asJSON {
		output, _ := json.MarshalIndent(devices, "", "  ")
		fmt.Println(string(output))
		return
	}

	printTable(devices)
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
