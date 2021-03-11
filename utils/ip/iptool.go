package iptool

import (
	"net"
	"net/http"
	"strconv"
	"strings"
)

func RealIP(r *http.Request) string {
	remote := strings.Split(r.RemoteAddr, ":")[0]
	if !IsInnerIp(remote) {
		return remote
	}
	forwarded := r.Header.Get("X-Forwarded-For")
	if len(forwarded) > 0 {
		ip := GetGlobalIPFromXforwardedFor(forwarded)
		if ip != "" {
			return ip
		}
	}
	forwarded = r.Header.Get("x-forwarded-for")
	if len(forwarded) > 0 {
		ip := GetGlobalIPFromXforwardedFor(forwarded)
		if ip != "" {
			return ip
		}
	}
	return remote
}

func CheckIp(ipStr string) bool {
	address := net.ParseIP(ipStr)
	if address == nil {
		//fmt.Println("ip地址格式不正确")
		return false
	} else {
		//fmt.Println("正确的ip地址", address.String())
		return true
	}
}

// ip to int64
func InetAton(ipStr string) int64 {
	bits := strings.Split(ipStr, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

//int64 to IP
func InetNtoa(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func IsInnerIp(ipStr string) bool {
	if !CheckIp(ipStr) {
		return false
	}
	inputIpNum := InetAton(ipStr)
	innerIpA := InetAton("10.255.255.255")
	innerIpB := InetAton("172.16.255.255")
	innerIpC := InetAton("192.168.255.255")
	innerIpD := InetAton("100.64.255.255")
	innerIpF := InetAton("127.255.255.255")

	return inputIpNum>>24 == innerIpA>>24 || inputIpNum>>20 == innerIpB>>20 ||
		inputIpNum>>16 == innerIpC>>16 || inputIpNum>>22 == innerIpD>>22 ||
		inputIpNum>>24 == innerIpF>>24
}

func GetGlobalIPFromXforwardedFor(xforwardedFor string) string {
	ips := strings.Split(xforwardedFor, ",")
	for _, ip := range ips {
		ip = strings.TrimSpace(ip)
		if !IsInnerIp(ip) {
			return ip
		}
	}
	return ""
}
