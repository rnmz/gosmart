# gosmart

Read disk SMART data on **Linux** and **Windows** in **pure Go** — standard library only, no cgo, no `smartctl`, no external tools.

- **SATA / ATA** drives: IDENTIFY, SMART attributes, thresholds, health verdict
- **NVMe** drives: Identify Controller, SMART / Health Information log
- Usable both as a **Go library** and as a **CLI** (text or JSON output)
- Protocol logic is separated from OS code and covered by unit tests that need no hardware

> **Status:** early. The protocol parsers are covered by unit tests, and both backends have been run on real machines (the Windows backend correctly flagged a failing drive). Coverage across drive models, bridges and Windows versions is still limited, so bug reports with real-world output are very welcome.

## Install

```bash
git clone https://github.com/rnmz/gosmart.git
cd gosmart
go build -o gosmart ./cmd/gosmart

# cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o gosmart.exe ./cmd/gosmart
```

Requires Go 1.21+. The module path is `vrx/gosmart`.

## CLI usage

Raw device access needs elevated privileges: **root** on Linux, an **Administrator** shell on Windows.

```bash
sudo ./gosmart                        # all detected disks
sudo ./gosmart /dev/sda /dev/nvme0n1  # specific devices
sudo ./gosmart -json                  # machine-readable output
sudo ./gosmart -no-attrs /dev/sda     # hide the ATA attribute table
```

```powershell
.\gosmart.exe                          # all disks
.\gosmart.exe \\.\PhysicalDrive0       # specific disk
```

Example output (illustrative):

```
== /dev/sda
ata | Samsung SSD 860 EVO 500GB | S/N S3Z2NB0K123456A | FW RVT04B6Q | 500.1 GB
Health: ok | Temp: 34°C | Power-on: 12873 h | Cycles: 412
  5 Reallocated_Sector_Ct      cur=100 worst=100 thr= 10 raw=0
  9 Power_On_Hours             cur= 94 worst= 94 thr=  0 raw=12873
...
== /dev/nvme0n1
nvme | WDC WDS100T3X0C | S/N 21234A800123 | FW 102000WD | 1000.2 GB
Health: ok | Temp: 41°C | Power-on: 3120 h | Cycles: 188
Used: 3% | Spare: 100% (threshold 10%) | Media errors: 0 | Unsafe shutdowns: 12 | Warning: 0x0
Written: 21450.3 GB | Read: 30211.7 GB
```

**Exit code** is a bitmask, handy for monitoring scripts:

| Bit | Meaning                                  |
|-----|------------------------------------------|
| 1   | at least one device could not be read    |
| 2   | at least one device reports `failing`    |
| 4   | at least one device reports `warning`    |

JSON output works well with `jq`:

```bash
sudo ./gosmart -json | jq '.[] | {path, health: .info.health, temp: .info.temperature_c}'
```

## Library usage

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"vrx/gosmart"
)

func main() {
	r := gosmart.New()

	// One disk. info may be non-nil even if err != nil (partial data).
	info, err := r.Read("/dev/sda")
	if errors.Is(err, gosmart.ErrPermission) {
		log.Fatal("run as root / Administrator")
	}
	if info != nil {
		fmt.Println(info.Model, info.Health)
	}

	// All disks; a failure on one disk does not stop the others.
	results, err := r.ReadAll(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, res := range results {
		if res.Err != nil {
			fmt.Printf("%s: %v\n", res.Path, res.Err)
			continue
		}
		fmt.Printf("%s: %s\n", res.Path, res.Info.Health)
	}
}
```

Key types (all exposed from the root package):

| Type         | Description                                                                  |
|--------------|------------------------------------------------------------------------------|
| `Info`       | model, serial, firmware, capacity, health, temperature, power-on hours, etc. |
| `Attribute`  | one ATA SMART attribute (id, name, current/worst/threshold, raw, prefail)    |
| `NVMeHealth` | NVMe SMART / Health log counters                                             |
| `Health`     | `ok`, `warning`, `failing` or `unknown`                                      |
| `Result`     | `{Path, Info, Err}` returned by `ReadAll`                                    |

Errors can be matched with `errors.Is`: `ErrPermission`, `ErrInvalidData`, `ErrUnsupportedPlatform`. Device-level failures are `*DeviceError{Path, Op, Err}`.

## How it works

| OS      | ATA / SATA                                       | NVMe                                               |
|---------|--------------------------------------------------|----------------------------------------------------|
| Linux   | `SG_IO` ioctl + SCSI ATA PASS-THROUGH(16)        | `NVME_IOCTL_ADMIN_CMD` (Identify, Get Log Page)    |
| Windows | `IOCTL_ATA_PASS_THROUGH` (`ATA_PASS_THROUGH_EX`) | `IOCTL_STORAGE_QUERY_PROPERTY` (protocol-specific) |

Health verdict:
- **ATA:** `failing` if any pre-fail attribute has `current <= threshold`; `warning` if the drive is not failing but the raw value of attribute 5 (reallocated sectors), 187 (reported uncorrectable errors), 197 (pending sectors) or 198 (offline uncorrectable) is non-zero
- **NVMe:** `failing` if the Critical Warning byte is non-zero; `warning` if there are media errors or the drive reports 90% or more of its rated life used

Vendors such as Samsung use very low failure thresholds, so a drive can accumulate bad blocks and still never cross them. That is why `warning` looks at raw error counters instead of only comparing against thresholds.

## Architecture

```
cmd/gosmart                 thin CLI
gosmart (root package)      facade: Reader.List / Read / ReadAll, backend selection per OS
model/                      domain types and errors, no dependencies
internal/transport/         ports: Backend, ATADevice, NVMeDevice
internal/ata/               pure ATA logic: commands, IDENTIFY/SMART parsing, summary
internal/nvme/              pure NVMe logic: Identify / health log parsing
internal/platform/linux/    SG_IO and NVMe admin ioctl
internal/platform/windows/  DeviceIoControl-based implementation
```

Dependencies point inward: `platform → transport ← ata / nvme ← gosmart`, and `model` depends on nothing. Protocol code never touches the OS, so it is tested against fake devices. Supporting a new OS or transport (USB-SAT bridges, RAID controllers) means implementing `transport.Backend`; nothing else changes.

## Testing

```bash
go test ./...                    # no hardware or root required
go test -race -cover ./...
GOOS=windows go vet ./...        # check the Windows backend compiles
```

## Limitations

- Needs root / Administrator.
- USB bridges and many hardware RAID controllers do not pass ATA commands through; you will get an error or only identification data.
- Virtual machine disks (`/dev/vda`, etc.) usually expose no SMART.
- NVMe 128-bit counters are truncated to the lower 64 bits.
- Reading SMART never enables it: if SMART is supported but disabled on the drive, the read fails with an error instead of changing drive state.

## Contributing

Issues and pull requests are welcome. The most useful contribution right now is real-world output: run `gosmart -json` next to `smartctl -a` on your drives and open an issue if the values differ.

## License

See `LICENSE`.
