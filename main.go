package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// Hop один узел на маршруте
type Hop struct {
	TTL     int           // номер прыжка
	IP      string        // IP-адрес маршрутизатора
	Host    string        // имя хоста (reverse DNS)
	RTT     time.Duration // время отклика
	Timeout bool          // флаг таймаута
}

// TraceResult результат всей трассировки
type TraceResult struct {
	Target string // целевой IP-адрес
	Hops   []Hop  // срез всех прыжков
}

const (
	maxHops    = 30
	timeout    = 3 * time.Second
	packetSize = 52
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <hostname>\nПример: geotracer google.com\n", os.Args[0])

		os.Exit(1)
	}

	target := args[1]

	fmt.Printf("Целевой хост: %s\n", target)

	targetIp, err := resolveHost(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %s\n", err)

		os.Exit(1)
	}

	fmt.Printf("IP-адрес: %s\n", targetIp)
	fmt.Println("Начинаем трассировку...")

	result, err := traceRoute(targetIp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка трассировки: %s\n", err)
		os.Exit(1)
	}

	for _, hop := range result.Hops {
		printHop(hop)
	}

	fmt.Printf("Трассировка завершена: найдено %d прыжков\n", len(result.Hops))
}

func resolveHost(host string) (string, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", err
	}

	for _, ip := range ips {
		if ip.To4() != nil {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no IPv4 address found")
}

func traceRoute(targetIP string) (*TraceResult, error) {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	packetConn := conn.IPv4PacketConn()

	traceResult := &TraceResult{
		Target: targetIP,
		Hops:   make([]Hop, 0, maxHops),
	}

	buf := make([]byte, 15000)

	for ttl := 1; ttl <= maxHops; ttl++ {
		if err := packetConn.SetTTL(ttl); err != nil {
			return nil, err
		}

		msg := icmp.Message{
			Type: ipv4.ICMPTypeEcho,
			Code: 0,
			Body: &icmp.Echo{
				ID:   1234,
				Seq:  ttl,
				Data: make([]byte, packetSize),
			},
		}

		msgBytes, err := msg.Marshal(nil)
		if err != nil {
			return nil, err
		}

		start := time.Now()

		_, err = packetConn.WriteTo(
			msgBytes,
			nil,
			&net.IPAddr{IP: net.ParseIP(targetIP)},
		)
		if err != nil {
			return nil, err
		}

		if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			return nil, err
		}

		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
				traceResult.Hops = append(traceResult.Hops, Hop{
					TTL:     ttl,
					Timeout: true,
				})
				continue
			}

			return nil, err
		}

		rtt := time.Since(start)

		ipAddr, ok := addr.(*net.IPAddr)
		if !ok {
			continue
		}

		routerIP := ipAddr.IP.String()

		reply, err := icmp.ParseMessage(1, buf[:n])
		if err != nil {
			return nil, err
		}

		switch reply.Type {
		case ipv4.ICMPTypeTimeExceeded, ipv4.ICMPTypeEchoReply:
			host := reverseLookup(routerIP)

			traceResult.Hops = append(traceResult.Hops, Hop{
				TTL:  ttl,
				IP:   routerIP,
				Host: host,
				RTT:  rtt,
			})

			if routerIP == targetIP {
				return traceResult, nil
			}
		}
	}

	return traceResult, nil
}

func reverseLookup(ip string) string {
	names, err := net.LookupAddr(ip)
	if err != nil || len(names) == 0 {
		return ""
	}

	return strings.TrimSuffix(names[0], ".")
}

func printHop(hop Hop) {
	if hop.Timeout {
		fmt.Printf("%3d  * * * (timeout)\n", hop.TTL)
		return
	}

	if hop.Host != "" {
		fmt.Printf("%3d  %-15s (%s) %v\n", hop.TTL, hop.IP, hop.Host, hop.RTT)
		return
	}

	fmt.Printf("%3d  %-15s %v\n", hop.TTL, hop.IP, hop.RTT)
}
