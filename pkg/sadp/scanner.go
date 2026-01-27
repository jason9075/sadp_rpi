package sadp

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/ipv4"
)

const (
	MulticastAddr  = "239.255.255.250:37020"
	ReadBufferSize = 4096
)

type Scanner struct {
	InterfaceName string
	Timeout       time.Duration
}

func NewScanner(iface string, timeout time.Duration) *Scanner {
	return &Scanner{
		InterfaceName: iface,
		Timeout:       timeout,
	}
}

func (s *Scanner) Scan(ctx context.Context) ([]Device, error) {
	// 1. 獲取網卡介面
	ifi, err := net.InterfaceByName(s.InterfaceName)
	if err != nil {
		return nil, fmt.Errorf("無法獲取網卡介面 %s: %w", s.InterfaceName, err)
	}

	// 2. 建立 UDP 連線
	c, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("建立 UDP 監聽失敗: %w", err)
	}
	defer c.Close()

	p := ipv4.NewPacketConn(c)

	// 3. 設定組播發送介面 (關鍵：非 root 指定介面的方法)
	if err := p.SetMulticastInterface(ifi); err != nil {
		return nil, fmt.Errorf("設定組播介面失敗: %w", err)
	}

	// 4. 準備發送 Payload
	probe := Probe{
		Uuid:  uuid.New().String(),
		Types: "inquiry",
	}
	payload, err := xml.MarshalIndent(probe, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化 XML 失敗: %w", err)
	}
	header := []byte(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	fullPayload := append(header, payload...)

	// 5. 發送組播包
	addr, err := net.ResolveUDPAddr("udp4", MulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("解析組播地址失敗: %w", err)
	}

	if _, err := p.WriteTo(fullPayload, nil, addr); err != nil {
		return nil, fmt.Errorf("發送組播包失敗: %w", err)
	}

	// 6. 接收回應
	devices := make([]Device, 0)
	deviceMap := make(map[string]bool)
	resChan := make(chan Device)
	errChan := make(chan error)

	go func() {
		buf := make([]byte, ReadBufferSize)
		for {
			_ = c.SetReadDeadline(time.Now().Add(s.Timeout))
			n, _, _, err := p.ReadFrom(buf)
			if err != nil {
				errChan <- err
				return
			}

			var match ProbeMatch
			if err := xml.Unmarshal(buf[:n], &match); err != nil {
				continue // 忽略無法解析的包
			}
			resChan <- match.Device
		}
	}()

	timer := time.After(s.Timeout)
	for {
		select {
		case dev := <-resChan:
			if !deviceMap[dev.MAC] {
				deviceMap[dev.MAC] = true
				devices = append(devices, dev)
			}
		case <-errChan:
			return devices, nil
		case <-timer:
			return devices, nil
		case <-ctx.Done():
			return devices, ctx.Err()
		}
	}
}
