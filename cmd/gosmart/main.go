package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/rnmz/gosmart"
)

type jsonResult struct {
	Path  string        `json:"path"`
	Info  *gosmart.Info `json:"info,omitempty"`
	Error string        `json:"error,omitempty"`
}

func main() {
	asJSON := flag.Bool("json", false, "output as JSON")
	noAttrs := flag.Bool("no-attrs", false, "do not print the ATA attribute table")
	flag.Parse()

	r := gosmart.New()
	var results []gosmart.Result
	if flag.NArg() == 0 {
		var err error
		if results, err = r.ReadAll(context.Background()); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	} else {
		for _, p := range flag.Args() {
			info, err := r.Read(p)
			results = append(results, gosmart.Result{Path: p, Info: info, Err: err})
		}
	}

	code := 0
	for _, res := range results {
		if res.Err != nil {
			code |= 1
		}
		if res.Info != nil && res.Info.Health == gosmart.HealthFailing {
			code |= 2
		}
		if res.Info != nil && res.Info.Health == gosmart.HealthWarning {
			code |= 4
		}
	}

	if *asJSON {
		out := make([]jsonResult, len(results))
		for i, res := range results {
			out[i] = jsonResult{Path: res.Path, Info: res.Info}
			if res.Err != nil {
				out[i].Error = res.Err.Error()
			}
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	} else {
		for _, res := range results {
			printText(os.Stdout, res, !*noAttrs)
		}
	}
	os.Exit(code)
}

func printText(w io.Writer, res gosmart.Result, attrs bool) {
	fmt.Fprintf(w, "== %s\n", res.Path)
	if res.Err != nil {
		fmt.Fprintf(w, "error: %v\n", res.Err)
	}
	i := res.Info
	if i == nil {
		return
	}
	fmt.Fprintf(w, "%s | %s | S/N %s | FW %s | %.1f GB\n",
		i.Protocol, i.Model, i.Serial, i.Firmware, float64(i.CapacityBytes)/1e9)
	temp := "n/a"
	if i.TemperatureC != nil {
		temp = fmt.Sprintf("%d°C", *i.TemperatureC)
	}
	fmt.Fprintf(w, "Health: %s | Temp: %s | Power-on: %d h | Cycles: %d\n",
		i.Health, temp, i.PowerOnHours, i.PowerCycles)
	if attrs {
		for _, a := range i.Attributes {
			fmt.Fprintf(w, "%3d %-26s cur=%3d worst=%3d thr=%3d raw=%d\n",
				a.ID, a.Name, a.Current, a.Worst, a.Threshold, a.Raw)
		}
	}
	if n := i.NVMe; n != nil {
		fmt.Fprintf(w, "Used: %d%% | Spare: %d%% (threshold %d%%) | Media errors: %d | Unsafe shutdowns: %d | Warning: %#x\n",
			n.PercentageUsed, n.AvailableSpare, n.SpareThreshold, n.MediaErrors, n.UnsafeShutdowns, n.CriticalWarning)
		fmt.Fprintf(w, "Written: %.1f GB | Read: %.1f GB\n",
			float64(n.DataUnitsWritten)*512000/1e9, float64(n.DataUnitsRead)*512000/1e9)
	}
}
