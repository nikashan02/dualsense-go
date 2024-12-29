// References C++ structures defined at https://controllers.fandom.com/wiki/Sony_DualSense#HID_Report_0x01_Input_USB

package dualsense

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type packedTouchData struct {
	TouchFinger1 uint32
	TouchFinger2 uint32
	Timestamp    uint8
}

type packedUSBGetStateData struct {
	LeftStickX             uint8
	LeftStickY             uint8
	RightStickX            uint8
	RightStickY            uint8
	TriggerLeft            uint8
	TriggerRight           uint8
	SeqNo                  uint8
	DPadActionButtons      uint8 // Contains DPad, Square, Cross, Circle, Triangle
	LeftRightCreateOptions uint8 // Contains L1, R1, L2, R2, Create, Options, L3, R3
	OtherButtons           uint8 // Contains Home, Touchpad, Mute, UNK1, ButtonLeftFunction, ButtonRightFunction, ButtonLeftPaddle, ButtonRightPaddle
	UNK2                   uint8 // unused
	UNK_COUNTER            uint32
	AngularVelocityX       int16
	AngularVelocityZ       int16
	AngularVelocityY       int16
	AccelerometerX         int16
	AccelerometerY         int16
	AccelerometerZ         int16
	SensorTimestamp        uint32
	Temperature            int8
	TouchData              packedTouchData
	TriggerRightDetails    uint8 // contains TriggerRightStopLocation and TriggerRight Status
	TriggerLeftDetails     uint8 // contains TriggerLeftStopLocation and TriggerLeft Status
	HostTimestamp          uint32
	TriggerEffects         uint8 // contains TriggerRightEffect and TriggerLeftEffect
	DeviceTimestamp        uint32
	PowerDetails           uint8 // contains PowerPercent and PowerState
	PlugInfoA              uint8 // contains PluggedHeadphones, PluggedMic, MicMuted, PluggedUsbData, PluggedUsbPower, PluggedUnk1
	PlugInfoB              uint8 // contains PluggedExternalMic, HapticLowPassFilter, PluggedUnk3
	AesCmac                uint64
}

type packedUSBReportIn struct {
	ReportID        uint8
	USBGetStateData packedUSBGetStateData
}

type TouchFinger struct {
	Index       uint8
	NotTouching bool
	FingerX     uint16
	FingerY     uint16
}

type TouchData struct {
	TouchFinger1 TouchFinger
	TouchFinger2 TouchFinger
	Timestamp    uint8
}

type Direction uint8

const (
	DirectionNorth Direction = iota
	DirectionNorthEast
	DirectionEast
	DirectionSouthEast
	DirectionSouth
	DirectionSouthWest
	DirectionWest
	DirectionNorthWest
	DirectionNone
)

type PowerState uint8

const (
	PowerStateDischarging         PowerState = 0x00
	PowerStateCharging            PowerState = 0x01
	PowerStateComplete            PowerState = 0x02
	PowerStateAbnormalVoltage     PowerState = 0x0A
	PowerStateAbnormalTemperature PowerState = 0x0B
	PowerStateChargingError       PowerState = 0x0F
)

type USBGetStateData struct {
	LeftStickX               uint8
	LeftStickY               uint8
	RightStickX              uint8
	RightStickY              uint8
	TriggerLeft              uint8
	TriggerRight             uint8
	SeqNo                    uint8
	DPad                     Direction
	ButtonSquare             bool
	ButtonCross              bool
	ButtonCircle             bool
	ButtonTriangle           bool
	ButtonL1                 bool
	ButtonR1                 bool
	ButtonL2                 bool
	ButtonR2                 bool
	ButtonCreate             bool
	ButtonOptions            bool
	ButtonL3                 bool
	ButtonR3                 bool
	ButtonHome               bool
	ButtonPad                bool
	ButtonMute               bool
	ButtonLeftFunction       bool // DualSense Edge
	ButtonRightFunction      bool // DualSense Edge
	ButtonLeftPaddle         bool // DualSense Edge
	ButtonRightPaddle        bool // DualSense Edge
	AngularVelocityX         int16
	AngularVelocityZ         int16
	AngularVelocityY         int16
	AccelerometerX           int16
	AccelerometerY           int16
	AccelerometerZ           int16
	SensorTimestamp          uint32
	Temperature              int8
	TouchData                TouchData
	TriggerRightStopLocation uint8
	TriggerRightStatus       uint8
	TriggerLeftStopLocation  uint8
	TriggerLeftStatus        uint8
	HostTimestamp            uint32
	TriggerRightEffect       uint8
	TriggerLeftEffect        uint8
	DeviceTimestamp          uint32
	PowerPercent             uint8
	PowerState               PowerState
	PluggedHeadphones        bool
	PluggedMic               bool
	MicMuted                 bool
	PluggedUsbData           bool
	PluggedUsbPower          bool
	PluggedExternalMic       bool
	HapticLowPassFilter      bool
	AesCmac                  uint64
}

type USBReportIn struct {
	ReportID        uint8
	USBGetStateData USBGetStateData
}

func getNthLittleEndianBitUint8(b uint8, n uint) uint8 {
	return (b >> n) & 1
}

func unpackUSBReportIn(data []byte) (USBReportIn, error) {
	if len(data) != USB_PACKET_SIZE {
		return USBReportIn{}, fmt.Errorf("invalid length of data: %d", len(data))
	}

	var report packedUSBReportIn
	err := binary.Read(bytes.NewReader(data), binary.LittleEndian, &report)
	if err != nil {
		return USBReportIn{}, fmt.Errorf("error trying to unpack USBReportIn: %w", err)
	}

	return USBReportIn{
		ReportID: report.ReportID,
		USBGetStateData: USBGetStateData{
			LeftStickX:          report.USBGetStateData.LeftStickX,
			LeftStickY:          report.USBGetStateData.LeftStickY,
			RightStickX:         report.USBGetStateData.RightStickX,
			RightStickY:         report.USBGetStateData.RightStickY,
			TriggerLeft:         report.USBGetStateData.TriggerLeft,
			TriggerRight:        report.USBGetStateData.TriggerRight,
			SeqNo:               report.USBGetStateData.SeqNo,
			DPad:                Direction(report.USBGetStateData.DPadActionButtons & 0x0F),
			ButtonSquare:        getNthLittleEndianBitUint8(report.USBGetStateData.DPadActionButtons, 4) == 1,
			ButtonCross:         getNthLittleEndianBitUint8(report.USBGetStateData.DPadActionButtons, 5) == 1,
			ButtonCircle:        getNthLittleEndianBitUint8(report.USBGetStateData.DPadActionButtons, 6) == 1,
			ButtonTriangle:      getNthLittleEndianBitUint8(report.USBGetStateData.DPadActionButtons, 7) == 1,
			ButtonL1:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 0) == 1,
			ButtonR1:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 1) == 1,
			ButtonL2:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 2) == 1,
			ButtonR2:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 3) == 1,
			ButtonCreate:        getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 4) == 1,
			ButtonOptions:       getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 5) == 1,
			ButtonL3:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 6) == 1,
			ButtonR3:            getNthLittleEndianBitUint8(report.USBGetStateData.LeftRightCreateOptions, 7) == 1,
			ButtonHome:          getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 0) == 1,
			ButtonPad:           getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 1) == 1,
			ButtonMute:          getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 2) == 1,
			ButtonLeftFunction:  getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 4) == 1,
			ButtonRightFunction: getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 5) == 1,
			ButtonLeftPaddle:    getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 6) == 1,
			ButtonRightPaddle:   getNthLittleEndianBitUint8(report.USBGetStateData.OtherButtons, 7) == 1,
			AngularVelocityX:    report.USBGetStateData.AngularVelocityX,
			AngularVelocityZ:    report.USBGetStateData.AngularVelocityZ,
			AngularVelocityY:    report.USBGetStateData.AngularVelocityY,
			AccelerometerX:      report.USBGetStateData.AccelerometerX,
			AccelerometerY:      report.USBGetStateData.AccelerometerY,
			AccelerometerZ:      report.USBGetStateData.AccelerometerZ,
			SensorTimestamp:     report.USBGetStateData.SensorTimestamp,
			Temperature:         report.USBGetStateData.Temperature,
			TouchData: TouchData{
				TouchFinger1: TouchFinger{
					Index:       uint8(report.USBGetStateData.TouchData.TouchFinger1 & 0x7F),
					NotTouching: ((report.USBGetStateData.TouchData.TouchFinger1 >> 7) & 1) == 1,
					FingerX:     uint16((report.USBGetStateData.TouchData.TouchFinger1 >> 8) & 0xFFF),
					FingerY:     uint16((report.USBGetStateData.TouchData.TouchFinger1 >> 20) & 0xFFF),
				},
				TouchFinger2: TouchFinger{
					Index:       uint8(report.USBGetStateData.TouchData.TouchFinger2 & 0x7F),
					NotTouching: ((report.USBGetStateData.TouchData.TouchFinger2 >> 7) & 1) == 1,
					FingerX:     uint16((report.USBGetStateData.TouchData.TouchFinger2 >> 8) & 0xFFF),
					FingerY:     uint16((report.USBGetStateData.TouchData.TouchFinger2 >> 20) & 0xFFF),
				},
				Timestamp: report.USBGetStateData.TouchData.Timestamp,
			},
			TriggerRightStopLocation: report.USBGetStateData.TriggerRightDetails & 0x0F,
			TriggerRightStatus:       report.USBGetStateData.TriggerRightDetails >> 4,
			TriggerLeftStopLocation:  report.USBGetStateData.TriggerLeftDetails & 0x0F,
			TriggerLeftStatus:        report.USBGetStateData.TriggerLeftDetails >> 4,
			HostTimestamp:            report.USBGetStateData.HostTimestamp,
			TriggerRightEffect:       report.USBGetStateData.TriggerEffects & 0x0F,
			TriggerLeftEffect:        report.USBGetStateData.TriggerEffects >> 4,
			DeviceTimestamp:          report.USBGetStateData.DeviceTimestamp,
			PowerPercent:             report.USBGetStateData.PowerDetails & 0x0F,
			PowerState:               PowerState(report.USBGetStateData.PowerDetails >> 4),
			PluggedHeadphones:        getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoA, 0) == 1,
			PluggedMic:               getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoA, 1) == 1,
			MicMuted:                 getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoA, 2) == 1,
			PluggedUsbData:           getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoA, 3) == 1,
			PluggedUsbPower:          getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoA, 4) == 1,
			PluggedExternalMic:       getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoB, 0) == 1,
			HapticLowPassFilter:      getNthLittleEndianBitUint8(report.USBGetStateData.PlugInfoB, 1) == 1,
			AesCmac:                  report.USBGetStateData.AesCmac,
		},
	}, nil
}
