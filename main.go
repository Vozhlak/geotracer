package main

import (
	"fmt"
	"net"
	"os"
	"time"
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

	traceRoutes, _ := traceRoute(targetIp)
	fmt.Printf("Трассировка завершена: найдено %d прыжков\n", len(traceRoutes.Hops))
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
	return &TraceResult{
		Target: targetIP,
		Hops:   make([]Hop, 0),
	}, nil
}
