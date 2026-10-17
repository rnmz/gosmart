package model

type Protocol string

const (
	ProtocolATA  Protocol = "ata"
	ProtocolNVMe Protocol = "nvme"
)

type Health string

const (
	HealthUnknown Health = "unknown"
	HealthOK      Health = "ok"
	HealthWarning Health = "warning"
	HealthFailing Health = "failing"
)

type Attribute struct {
	ID        uint8  `json:"id"`
	Name      string `json:"name"`
	Flags     uint16 `json:"flags"`
	PreFail   bool   `json:"prefail"`
	Current   uint8  `json:"current"`
	Worst     uint8  `json:"worst"`
	Threshold uint8  `json:"threshold"`
	Raw       uint64 `json:"raw"`
}

func (a Attribute) Failing() bool {
	return a.PreFail && a.Threshold != 0 && a.Current <= a.Threshold
}

type NVMeHealth struct {
	CriticalWarning   uint8  `json:"critical_warning"`
	AvailableSpare    uint8  `json:"available_spare_pct"`
	SpareThreshold    uint8  `json:"spare_threshold_pct"`
	PercentageUsed    uint8  `json:"percentage_used"`
	DataUnitsRead     uint64 `json:"data_units_read"`
	DataUnitsWritten  uint64 `json:"data_units_written"`
	HostReads         uint64 `json:"host_reads"`
	HostWrites        uint64 `json:"host_writes"`
	ControllerBusyMin uint64 `json:"controller_busy_min"`
	UnsafeShutdowns   uint64 `json:"unsafe_shutdowns"`
	MediaErrors       uint64 `json:"media_errors"`
	ErrorLogEntries   uint64 `json:"error_log_entries"`
}

type Info struct {
	Path           string      `json:"path"`
	Protocol       Protocol    `json:"protocol"`
	Model          string      `json:"model"`
	Serial         string      `json:"serial"`
	Firmware       string      `json:"firmware"`
	CapacityBytes  uint64      `json:"capacity_bytes,omitempty"`
	RotationRate   uint16      `json:"rotation_rate,omitempty"`
	SMARTSupported bool        `json:"smart_supported"`
	Health         Health      `json:"health"`
	TemperatureC   *int        `json:"temperature_c,omitempty"`
	PowerOnHours   uint64      `json:"power_on_hours,omitempty"`
	PowerCycles    uint64      `json:"power_cycles,omitempty"`
	Attributes     []Attribute `json:"attributes,omitempty"`
	NVMe           *NVMeHealth `json:"nvme,omitempty"`
}
