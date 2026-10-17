package ata

import "fmt"

func AttributeName(id uint8) string {
	if n, ok := names[id]; ok {
		return n
	}
	return fmt.Sprintf("Unknown_Attribute_%d", id)
}

var names = map[uint8]string{
	1: "Raw_Read_Error_Rate", 3: "Spin_Up_Time", 4: "Start_Stop_Count",
	5: "Reallocated_Sector_Ct", 7: "Seek_Error_Rate", 9: "Power_On_Hours",
	10: "Spin_Retry_Count", 12: "Power_Cycle_Count", 177: "Wear_Leveling_Count",
	179: "Used_Rsvd_Blk_Cnt_Tot", 181: "Program_Fail_Cnt_Total", 182: "Erase_Fail_Count_Total", 183: "Runtime_Bad_Block",
	187: "Reported_Uncorrect", 188: "Command_Timeout", 190: "Airflow_Temperature_Cel",
	191: "G-Sense_Error_Rate", 192: "Power-Off_Retract_Count", 193: "Load_Cycle_Count",
	194: "Temperature_Celsius", 195: "Hardware_ECC_Recovered", 196: "Reallocated_Event_Count",
	197: "Current_Pending_Sector", 198: "Offline_Uncorrectable", 199: "UDMA_CRC_Error_Count",
	202: "Percent_Lifetime_Remain", 231: "SSD_Life_Left", 233: "Media_Wearout_Indicator", 235: "POR_Recovery_Count",
	241: "Total_LBAs_Written", 242: "Total_LBAs_Read",
}
