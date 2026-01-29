# UDP-Smasher  
**High-Performance UDP Stress Testing & DDoS Simulation Tool**

---

## 📌 Overview
**UDP-Smasher** is a high-performance UDP packet generator designed for **network stress testing**, **DDoS resilience testing**, and **protocol validation** in **authorized environments**.

It is written in **Go** and uses **raw sockets** to achieve extremely high packet throughput, enabling realistic simulation of **UDP-based Distributed Denial-of-Service (DDoS) traffic patterns** for defensive testing, benchmarking, and research.

> ⚠️ This tool is intended **only** for networks you own or have **explicit permission** to test.

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
- Root privileges or `cap_net_raw`

### Build
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

---

## 🛡️ Legal & Ethical Use
This tool is intended strictly for **authorized testing**, **DDoS simulation**, and **research purposes**.
Unauthorized usage may be illegal.

---

## 📄 License
MIT License
