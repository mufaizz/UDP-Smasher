# UDP-Smasher  
**Elite High-Performance UDP Network Stress Testing Tool**
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/mufaizz/UDP-Smasher/blob/main/LICENSE)
---

## 📌 Overview
**UDP-Smasher** is a high-performance UDP packet generator designed for **authorized network stress testing**, **protocol validation**, and **research environments**.  
Written in **Go** using **raw sockets**, it achieves extremely high packet throughput by minimizing kernel overhead and leveraging efficient memory handling techniques.

License: MIT — see the [LICENSE](./LICENSE) file for full terms.

> ⚠️ This tool is strictly for **legal and authorized use only**.

---

## ⚡ Performance Highlights
- **Throughput:** ~350,000 – 400,000 packets per second (sustained)
- **Packet Size:** 28 bytes (minimum UDP + IPv4 headers)
- **Concurrency Model:** CPU cores × 8 workers
- **Source IP Diversity:** /16 subnet spoofing (65,536 unique IPs)
- **Latency:** Sub-microsecond batch dispatch

---

## 🏗️ Architecture Overview
```
┌─────────────────────────────────────────────┐
│ Control Plane                               │
│ • Interactive CLI                           │
│ • Automatic interface detection             │
│ • Real-time PPS monitoring                  │
└───────────────────┬─────────────────────────┘
                    │
┌───────────────────▼─────────────────────────┐
│ Worker Pool (CPU × 8)                        │
│ ┌─────────┐ ┌─────────┐ ┌─────────┐         │
│ │ Worker  │ │ Worker  │ │ Worker  │  ...    │
│ └────┬────┘ └────┬────┘ └────┬────┘         │
└───────┼───────────┼──────────────┼─────────┘
        │           │              │
┌───────▼───────────▼──────────────▼─────────┐
│ Packet Factory                              │
│ • Zero-allocation packet crafting           │
│ • Incremental checksum computation          │
│ • Randomized source IP & port generation    │
└───────┬────────────────────────────────────┘
        │
┌───────▼────────────────────────────────────┐
│ Raw Socket Layer                            │
│ • IP_HDRINCL enabled                        │
│ • Interface-bound sockets                  │
│ • Large socket buffers (128MB)              │
└───────────────────────────────────────────┘
```

---

## 🚀 Installation

### Requirements
- Linux (raw socket support required)
- Go 1.20+
- Root or `cap_net_raw` capability

### Build
```bash
```bash
git clone https://github.com/mufaizz/UDP-Smasher.git
cd UDP-Smasher
go build -ldflags "-s -w" -o attack main.go
```

### Set Capabilities (Recommended)
```bash
sudo setcap cap_net_raw=ep attack
```

---

## 📖 Usage
```bash
sudo ./attack
```
Interactive Prompt
```
Target IP:   192.168.1.100
Target Port: 8080
```

The tool will immediately begin packet transmission and display live PPS statistics.

---
## 🔧 Technical Implementation
Packet Structure
```
[ IPv4 Header : 20 bytes ]
[ UDP Header  :  8 bytes ]
```


Key Fields

- TTL: 64

- Protocol: UDP (17)

- IP Identification: Randomized

- Source IP: Randomized from /16 range

- Source Port: Random (1024–65535)

- Checksums: Incrementally computed

## ⚙️ Optimization Techniques

- Batch Processing: 1024 packets per syscall loop

- Memory Pooling: Pre-allocated reusable buffers

- Lock-Free Counters: Atomic operations for stats

- CPU Affinity: Workers pinned to CPU cores

- Kernel Bypass: Raw sockets with full header control

## 📊 Performance Tuning
Recommended Kernel Parameters
```
sudo sysctl -w net.core.rmem_max=134217728
sudo sysctl -w net.core.wmem_max=134217728
sudo sysctl -w net.ipv4.udp_mem="134217728 134217728 134217728"
```
Optional NIC Offload Disable
```
sudo ethtool -K eth0 tx off rx off gso off gro off
```

## 🧪 Testing & Validation

- Baseline: iperf -s -u

- Packet Capture: tcpdump -i eth0 udp

- Traffic Monitor: iftop -i eth0

- Socket Stats: ss -unp

## 📈 Runtime Profile
```
CPU Usage:        90–95% kernel, 5–10% user
Memory Usage:     < 50 MB RSS
Threads:          CPU cores + 2
File Descriptors: One raw socket per worker
Context Switches: Minimal (LockOSThread)
```

## 🐛 Known Limitations

- Reduced throughput in virtualized environments

- Linux-only (raw socket dependency)

- Requires elevated privileges

- Bypasses netfilter / conntrack
## 🛡️ Legal & Ethical Use
 ✔ Allowed

- Authorized penetration testing

- Network equipment benchmarking

- Academic & protocol research

- CTFs and lab environments (with permission)

 ❌ Prohibited

- Unauthorized testing

- Malicious activity

- Violation of local or international laws

- Users are solely responsible for legal compliance.

## 🤝 Contributing

1.Fork the repository

2.Create a feature branch

3.Commit your changes

4.Push to your fork

5.Open a Pull Request

## License

This repository is licensed under the MIT License — see the [LICENSE](./LICENSE) file for details.

##Contact

E-mail: mufaizmalik9622@gmail.com

Instagram: https://www.instagram.com/muf4iz

## 🙏 Acknowledgments

- **Linux kernel networking stack

- **Go syscall and net packages

- **Open-source security research community

## Disclaimer:
This project demonstrates low-level networking capabilities for research and testing purposes.
Use responsibly and only on networks you own or are explicitly authorized to test.
