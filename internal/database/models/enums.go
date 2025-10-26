package models

// ScheduleType defines the types of duty schedules
type ScheduleType string

const (
	ScheduleTypeOnCall      ScheduleType = "on_call"
	ScheduleTypeSupport     ScheduleType = "support"
	ScheduleTypeMaintenance ScheduleType = "maintenance"
	ScheduleTypeDeployment  ScheduleType = "deployment"
)

// ShiftType defines the types of shifts for duty schedules
type ShiftType string

const (
	ShiftTypeDay     ShiftType = "day"
	ShiftTypeNight   ShiftType = "night"
	ShiftTypeWeekend ShiftType = "weekend"
	ShiftTypeHoliday ShiftType = "holiday"
)

// IsValid checks if the ScheduleType is valid
func (s ScheduleType) IsValid() bool {
	switch s {
	case ScheduleTypeOnCall, ScheduleTypeSupport, ScheduleTypeMaintenance, ScheduleTypeDeployment:
		return true
	}
	return false
}

// IsValid checks if the ShiftType is valid
func (s ShiftType) IsValid() bool {
	switch s {
	case ShiftTypeDay, ShiftTypeNight, ShiftTypeWeekend, ShiftTypeHoliday:
		return true
	}
	return false
}
