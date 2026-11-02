package main

import (
	"fmt"
	"net"
	"os"
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
