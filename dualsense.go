package dualsense

import (
	"fmt"
	"sync"
	"time"

	hid "github.com/sstallion/go-hid"
)

const (
	DUALSENSE_VENDOR_ID  = 0x054C
	DUALSENSE_PRODUCT_ID = 0x0CE6
	DEFAULT_READ_TIMEOUT = 100 * time.Millisecond
	USB_PACKET_SIZE      = 64
	DEFAULT_POLLING_RATE = 50 * time.Millisecond
)

type callbacks struct {
	OnLeftStickXChange               []func(uint8)
	OnLeftStickYChange               []func(uint8)
	OnRightStickXChange              []func(uint8)
	OnRightStickYChange              []func(uint8)
	OnTriggerLeftChange              []func(uint8)
	OnTriggerRightChange             []func(uint8)
	OnDPadChange                     []func(Direction)
	OnButtonSquareChange             []func(bool)
	OnButtonCrossChange              []func(bool)
	OnButtonCircleChange             []func(bool)
	OnButtonTriangleChange           []func(bool)
	OnButtonL1Change                 []func(bool)
	OnButtonR1Change                 []func(bool)
	OnButtonL2Change                 []func(bool)
	OnButtonR2Change                 []func(bool)
	OnButtonCreateChange             []func(bool)
	OnButtonOptionsChange            []func(bool)
	OnButtonL3Change                 []func(bool)
	OnButtonR3Change                 []func(bool)
	OnButtonHomeChange               []func(bool)
	OnButtonPadChange                []func(bool)
	OnButtonMuteChange               []func(bool)
	OnButtonLeftFunctionChange       []func(bool)
	OnButtonRightFunctionChange      []func(bool)
	OnButtonLeftPaddleChange         []func(bool)
	OnButtonRightPaddleChange        []func(bool)
	OnAngularVelocityXChange         []func(int16)
	OnAngularVelocityZChange         []func(int16)
	OnAngularVelocityYChange         []func(int16)
	OnAccelerometerXChange           []func(int16)
	OnAccelerometerYChange           []func(int16)
	OnAccelerometerZChange           []func(int16)
	OnTemperatureChange              []func(int8)
	OnTouchFinger1Change             []func(TouchFinger)
	OnTouchFinger2Change             []func(TouchFinger)
	OnTriggerRightStopLocationChange []func(uint8)
	OnTriggerRightStatusChange       []func(uint8)
	OnTriggerLeftStopLocationChange  []func(uint8)
	OnTriggerLeftStatusChange        []func(uint8)
	OnTriggerRightEffectChange       []func(uint8)
	OnTriggerLeftEffectChange        []func(uint8)
	OnPowerPercentChange             []func(uint8)
	OnPowerStateChange               []func(PowerState)
	OnPluggedHeadphonesChange        []func(bool)
	OnPluggedMicChange               []func(bool)
	OnMicMutedChange                 []func(bool)
	OnPluggedUsbDataChange           []func(bool)
	OnPluggedExternalMicChange       []func(bool)
	OnHapticLowPassFilterChange      []func(bool)
}

type DualSense struct {
	device           *hid.Device
	GetStateData     USBGetStateData
	usbReportInClose chan bool
	setStateData     SetStateData
	setStateDataMu   sync.Mutex
	callbacks        callbacks
	pollingRate      time.Duration
}

func NewDualSense() (*DualSense, error) {
	device, err := hid.OpenFirst(DUALSENSE_VENDOR_ID, DUALSENSE_PRODUCT_ID)
	if err != nil {
		return nil, fmt.Errorf("error trying to open DualSense controller: %w", err)
	}
	err = device.SetNonblock(false)
	if err != nil {
		return nil, fmt.Errorf("error trying to set DualSense controller to blocking mode: %w", err)
	}
	usbReportInClose := make(chan bool)
	dualsense := &DualSense{
		device:           device,
		usbReportInClose: usbReportInClose,
		pollingRate:      DEFAULT_POLLING_RATE,
	}
	return dualsense, nil
}

func (d *DualSense) Start(initialSetStateData *SetStateData) error {
	go d.listenReportIn()
	var err error
	if initialSetStateData == nil {
		err = d.SetStateData(defaultSetStateData)
	} else {
		err = d.SetStateData(*initialSetStateData)
	}
	if err != nil {
		return fmt.Errorf("setStateData: error trying to set initial state data for DualSense controller: %w", err)
	}
	return nil
}

func (d *DualSense) SetPollingRate(pollingRateHz int) {
	d.pollingRate = time.Duration(1000/pollingRateHz) * time.Millisecond
}

func (d *DualSense) Close() {
	d.usbReportInClose <- true
	d.device.Close()
}

func (d *DualSense) readReportIn() (USBReportIn, error) {
	buffer := make([]byte, USB_PACKET_SIZE)
	bytesRead, err := d.device.ReadWithTimeout(buffer, DEFAULT_READ_TIMEOUT)
	if err != nil {
		return USBReportIn{}, fmt.Errorf("device.ReadWithTimeout: error trying to read DualSense controller input report: %w", err)
	}
	if bytesRead != USB_PACKET_SIZE {
		return USBReportIn{}, fmt.Errorf("device.ReadWithTimeout: error trying to read DualSense controller input report: expected %d bytes, got %d bytes", USB_PACKET_SIZE, bytesRead)
	}
	reportIn, err := unpackUSBReportIn(buffer)
	if err != nil {
		return USBReportIn{}, fmt.Errorf("unpackUSBReportIn: error trying to unpack DualSense controller input report: %w", err)
	}
	return reportIn, err
}

func (d *DualSense) triggerCallbacks(previousGetStateData USBGetStateData) {
	if d.GetStateData.LeftStickX != previousGetStateData.LeftStickX {
		for _, callback := range d.callbacks.OnLeftStickXChange {
			callback(d.GetStateData.LeftStickX)
		}
	}
	if d.GetStateData.LeftStickY != previousGetStateData.LeftStickY {
		for _, callback := range d.callbacks.OnLeftStickYChange {
			callback(d.GetStateData.LeftStickY)
		}
	}
	if d.GetStateData.RightStickX != previousGetStateData.RightStickX {
		for _, callback := range d.callbacks.OnRightStickXChange {
			callback(d.GetStateData.RightStickX)
		}
	}
	if d.GetStateData.RightStickY != previousGetStateData.RightStickY {
		for _, callback := range d.callbacks.OnRightStickYChange {
			callback(d.GetStateData.RightStickY)
		}
	}
	if d.GetStateData.TriggerLeft != previousGetStateData.TriggerLeft {
		for _, callback := range d.callbacks.OnTriggerLeftChange {
			callback(d.GetStateData.TriggerLeft)
		}
	}
	if d.GetStateData.TriggerRight != previousGetStateData.TriggerRight {
		for _, callback := range d.callbacks.OnTriggerRightChange {
			callback(d.GetStateData.TriggerRight)
		}
	}
	if d.GetStateData.DPad != previousGetStateData.DPad {
		for _, callback := range d.callbacks.OnDPadChange {
			callback(d.GetStateData.DPad)
		}
	}
	if d.GetStateData.ButtonSquare != previousGetStateData.ButtonSquare {
		for _, callback := range d.callbacks.OnButtonSquareChange {
			callback(d.GetStateData.ButtonSquare)
		}
	}
	if d.GetStateData.ButtonCross != previousGetStateData.ButtonCross {
		for _, callback := range d.callbacks.OnButtonCrossChange {
			callback(d.GetStateData.ButtonCross)
		}
	}
	if d.GetStateData.ButtonCircle != previousGetStateData.ButtonCircle {
		for _, callback := range d.callbacks.OnButtonCircleChange {
			callback(d.GetStateData.ButtonCircle)
		}
	}
	if d.GetStateData.ButtonTriangle != previousGetStateData.ButtonTriangle {
		for _, callback := range d.callbacks.OnButtonTriangleChange {
			callback(d.GetStateData.ButtonTriangle)
		}
	}
	if d.GetStateData.ButtonL1 != previousGetStateData.ButtonL1 {
		for _, callback := range d.callbacks.OnButtonL1Change {
			callback(d.GetStateData.ButtonL1)
		}
	}
	if d.GetStateData.ButtonR1 != previousGetStateData.ButtonR1 {
		for _, callback := range d.callbacks.OnButtonR1Change {
			callback(d.GetStateData.ButtonR1)
		}
	}
	if d.GetStateData.ButtonL2 != previousGetStateData.ButtonL2 {
		for _, callback := range d.callbacks.OnButtonL2Change {
			callback(d.GetStateData.ButtonL2)
		}
	}
	if d.GetStateData.ButtonR2 != previousGetStateData.ButtonR2 {
		for _, callback := range d.callbacks.OnButtonR2Change {
			callback(d.GetStateData.ButtonR2)
		}
	}
	if d.GetStateData.ButtonCreate != previousGetStateData.ButtonCreate {
		for _, callback := range d.callbacks.OnButtonCreateChange {
			callback(d.GetStateData.ButtonCreate)
		}
	}
	if d.GetStateData.ButtonOptions != previousGetStateData.ButtonOptions {
		for _, callback := range d.callbacks.OnButtonOptionsChange {
			callback(d.GetStateData.ButtonOptions)
		}
	}
	if d.GetStateData.ButtonL3 != previousGetStateData.ButtonL3 {
		for _, callback := range d.callbacks.OnButtonL3Change {
			callback(d.GetStateData.ButtonL3)
		}
	}
	if d.GetStateData.ButtonR3 != previousGetStateData.ButtonR3 {
		for _, callback := range d.callbacks.OnButtonR3Change {
			callback(d.GetStateData.ButtonR3)
		}
	}
	if d.GetStateData.ButtonHome != previousGetStateData.ButtonHome {
		for _, callback := range d.callbacks.OnButtonHomeChange {
			callback(d.GetStateData.ButtonHome)
		}
	}
	if d.GetStateData.ButtonPad != previousGetStateData.ButtonPad {
		for _, callback := range d.callbacks.OnButtonPadChange {
			callback(d.GetStateData.ButtonPad)
		}
	}
	if d.GetStateData.ButtonMute != previousGetStateData.ButtonMute {
		for _, callback := range d.callbacks.OnButtonMuteChange {
			callback(d.GetStateData.ButtonMute)
		}
	}
	if d.GetStateData.ButtonLeftFunction != previousGetStateData.ButtonLeftFunction {
		for _, callback := range d.callbacks.OnButtonLeftFunctionChange {
			callback(d.GetStateData.ButtonLeftFunction)
		}
	}
	if d.GetStateData.ButtonRightFunction != previousGetStateData.ButtonRightFunction {
		for _, callback := range d.callbacks.OnButtonRightFunctionChange {
			callback(d.GetStateData.ButtonRightFunction)
		}
	}
	if d.GetStateData.ButtonLeftPaddle != previousGetStateData.ButtonLeftPaddle {
		for _, callback := range d.callbacks.OnButtonLeftPaddleChange {
			callback(d.GetStateData.ButtonLeftPaddle)
		}
	}
	if d.GetStateData.ButtonRightPaddle != previousGetStateData.ButtonRightPaddle {
		for _, callback := range d.callbacks.OnButtonRightPaddleChange {
			callback(d.GetStateData.ButtonRightPaddle)
		}
	}
	if d.GetStateData.AngularVelocityX != previousGetStateData.AngularVelocityX {
		for _, callback := range d.callbacks.OnAngularVelocityXChange {
			callback(d.GetStateData.AngularVelocityX)
		}
	}
	if d.GetStateData.AngularVelocityZ != previousGetStateData.AngularVelocityZ {
		for _, callback := range d.callbacks.OnAngularVelocityZChange {
			callback(d.GetStateData.AngularVelocityZ)
		}
	}
	if d.GetStateData.AngularVelocityY != previousGetStateData.AngularVelocityY {
		for _, callback := range d.callbacks.OnAngularVelocityYChange {
			callback(d.GetStateData.AngularVelocityY)
		}
	}
	if d.GetStateData.AccelerometerX != previousGetStateData.AccelerometerX {
		for _, callback := range d.callbacks.OnAccelerometerXChange {
			callback(d.GetStateData.AccelerometerX)
		}
	}
	if d.GetStateData.AccelerometerY != previousGetStateData.AccelerometerY {
		for _, callback := range d.callbacks.OnAccelerometerYChange {
			callback(d.GetStateData.AccelerometerY)
		}
	}
	if d.GetStateData.AccelerometerZ != previousGetStateData.AccelerometerZ {
		for _, callback := range d.callbacks.OnAccelerometerZChange {
			callback(d.GetStateData.AccelerometerZ)
		}
	}
	if d.GetStateData.Temperature != previousGetStateData.Temperature {
		for _, callback := range d.callbacks.OnTemperatureChange {
			callback(d.GetStateData.Temperature)
		}
	}
	if d.GetStateData.TouchData.TouchFinger1 != previousGetStateData.TouchData.TouchFinger1 {
		for _, callback := range d.callbacks.OnTouchFinger1Change {
			callback(d.GetStateData.TouchData.TouchFinger1)
		}
	}
	if d.GetStateData.TouchData.TouchFinger2 != previousGetStateData.TouchData.TouchFinger2 {
		for _, callback := range d.callbacks.OnTouchFinger2Change {
			callback(d.GetStateData.TouchData.TouchFinger2)
		}
	}
	if d.GetStateData.TriggerRightStopLocation != previousGetStateData.TriggerRightStopLocation {
		for _, callback := range d.callbacks.OnTriggerRightStopLocationChange {
			callback(d.GetStateData.TriggerRightStopLocation)
		}
	}
	if d.GetStateData.TriggerRightStatus != previousGetStateData.TriggerRightStatus {
		for _, callback := range d.callbacks.OnTriggerRightStatusChange {
			callback(d.GetStateData.TriggerRightStatus)
		}
	}
	if d.GetStateData.TriggerLeftStopLocation != previousGetStateData.TriggerLeftStopLocation {
		for _, callback := range d.callbacks.OnTriggerLeftStopLocationChange {
			callback(d.GetStateData.TriggerLeftStopLocation)
		}
	}
	if d.GetStateData.TriggerLeftStatus != previousGetStateData.TriggerLeftStatus {
		for _, callback := range d.callbacks.OnTriggerLeftStatusChange {
			callback(d.GetStateData.TriggerLeftStatus)
		}
	}
	if d.GetStateData.TriggerRightEffect != previousGetStateData.TriggerRightEffect {
		for _, callback := range d.callbacks.OnTriggerRightEffectChange {
			callback(d.GetStateData.TriggerRightEffect)
		}
	}
	if d.GetStateData.TriggerLeftEffect != previousGetStateData.TriggerLeftEffect {
		for _, callback := range d.callbacks.OnTriggerLeftEffectChange {
			callback(d.GetStateData.TriggerLeftEffect)
		}
	}
	if d.GetStateData.PowerPercent != previousGetStateData.PowerPercent {
		for _, callback := range d.callbacks.OnPowerPercentChange {
			callback(d.GetStateData.PowerPercent)
		}
	}
	if d.GetStateData.PowerState != previousGetStateData.PowerState {
		for _, callback := range d.callbacks.OnPowerStateChange {
			callback(d.GetStateData.PowerState)
		}
	}
	if d.GetStateData.PluggedHeadphones != previousGetStateData.PluggedHeadphones {
		for _, callback := range d.callbacks.OnPluggedHeadphonesChange {
			callback(d.GetStateData.PluggedHeadphones)
		}
	}
	if d.GetStateData.PluggedMic != previousGetStateData.PluggedMic {
		for _, callback := range d.callbacks.OnPluggedMicChange {
			callback(d.GetStateData.PluggedMic)
		}
	}
	if d.GetStateData.MicMuted != previousGetStateData.MicMuted {
		for _, callback := range d.callbacks.OnMicMutedChange {
			callback(d.GetStateData.MicMuted)
		}
	}
	if d.GetStateData.PluggedUsbData != previousGetStateData.PluggedUsbData {
		for _, callback := range d.callbacks.OnPluggedUsbDataChange {
			callback(d.GetStateData.PluggedUsbData)
		}
	}
	if d.GetStateData.PluggedExternalMic != previousGetStateData.PluggedExternalMic {
		for _, callback := range d.callbacks.OnPluggedExternalMicChange {
			callback(d.GetStateData.PluggedExternalMic)
		}
	}
	if d.GetStateData.HapticLowPassFilter != previousGetStateData.HapticLowPassFilter {
		for _, callback := range d.callbacks.OnHapticLowPassFilterChange {
			callback(d.GetStateData.HapticLowPassFilter)
		}
	}
}

func (d *DualSense) listenReportIn() {
	for {
		select {
		case <-d.usbReportInClose:
			return
		default:
			reportIn, err := d.readReportIn()
			if err == nil {
				previousGetStateData := d.GetStateData
				d.GetStateData = reportIn.USBGetStateData
				d.triggerCallbacks(previousGetStateData)
			}
			time.Sleep(d.pollingRate)
		}
	}
}

func (d *DualSense) SetStateData(setStateData SetStateData) error {
	packedUSBReportOut, err := packUSBReportOut(setStateData)
	if err != nil {
		return fmt.Errorf("packUSBReportOut: error trying to pack DualSense controller output report: %w", err)
	}
	d.setStateDataMu.Lock()
	_, err = d.device.Write(packedUSBReportOut)
	if err != nil {
		err = fmt.Errorf("device.Write: error trying to write DualSense controller output report: %w", err)
	} else {
		d.setStateData = setStateData
	}
	d.setStateDataMu.Unlock()
	return err
}

func (d *DualSense) OnLeftStickXChange(callback func(uint8)) {
	d.callbacks.OnLeftStickXChange = append(d.callbacks.OnLeftStickXChange, callback)
}

func (d *DualSense) OnLeftStickYChange(callback func(uint8)) {
	d.callbacks.OnLeftStickYChange = append(d.callbacks.OnLeftStickYChange, callback)
}

func (d *DualSense) OnRightStickXChange(callback func(uint8)) {
	d.callbacks.OnRightStickXChange = append(d.callbacks.OnRightStickXChange, callback)
}

func (d *DualSense) OnRightStickYChange(callback func(uint8)) {
	d.callbacks.OnRightStickYChange = append(d.callbacks.OnRightStickYChange, callback)
}

func (d *DualSense) OnTriggerLeftChange(callback func(uint8)) {
	d.callbacks.OnTriggerLeftChange = append(d.callbacks.OnTriggerLeftChange, callback)
}

func (d *DualSense) OnTriggerRightChange(callback func(uint8)) {
	d.callbacks.OnTriggerRightChange = append(d.callbacks.OnTriggerRightChange, callback)
}

func (d *DualSense) OnDPadChange(callback func(Direction)) {
	d.callbacks.OnDPadChange = append(d.callbacks.OnDPadChange, callback)
}

func (d *DualSense) OnButtonSquareChange(callback func(bool)) {
	d.callbacks.OnButtonSquareChange = append(d.callbacks.OnButtonSquareChange, callback)
}

func (d *DualSense) OnButtonCrossChange(callback func(bool)) {
	d.callbacks.OnButtonCrossChange = append(d.callbacks.OnButtonCrossChange, callback)
}

func (d *DualSense) OnButtonCircleChange(callback func(bool)) {
	d.callbacks.OnButtonCircleChange = append(d.callbacks.OnButtonCircleChange, callback)
}

func (d *DualSense) OnButtonTriangleChange(callback func(bool)) {
	d.callbacks.OnButtonTriangleChange = append(d.callbacks.OnButtonTriangleChange, callback)
}

func (d *DualSense) OnButtonL1Change(callback func(bool)) {
	d.callbacks.OnButtonL1Change = append(d.callbacks.OnButtonL1Change, callback)
}

func (d *DualSense) OnButtonR1Change(callback func(bool)) {
	d.callbacks.OnButtonR1Change = append(d.callbacks.OnButtonR1Change, callback)
}

func (d *DualSense) OnButtonL2Change(callback func(bool)) {
	d.callbacks.OnButtonL2Change = append(d.callbacks.OnButtonL2Change, callback)
}

func (d *DualSense) OnButtonR2Change(callback func(bool)) {
	d.callbacks.OnButtonR2Change = append(d.callbacks.OnButtonR2Change, callback)
}

func (d *DualSense) OnButtonCreateChange(callback func(bool)) {
	d.callbacks.OnButtonCreateChange = append(d.callbacks.OnButtonCreateChange, callback)
}

func (d *DualSense) OnButtonOptionsChange(callback func(bool)) {
	d.callbacks.OnButtonOptionsChange = append(d.callbacks.OnButtonOptionsChange, callback)
}

func (d *DualSense) OnButtonL3Change(callback func(bool)) {
	d.callbacks.OnButtonL3Change = append(d.callbacks.OnButtonL3Change, callback)
}

func (d *DualSense) OnButtonR3Change(callback func(bool)) {
	d.callbacks.OnButtonR3Change = append(d.callbacks.OnButtonR3Change, callback)
}

func (d *DualSense) OnButtonHomeChange(callback func(bool)) {
	d.callbacks.OnButtonHomeChange = append(d.callbacks.OnButtonHomeChange, callback)
}

func (d *DualSense) OnButtonPadChange(callback func(bool)) {
	d.callbacks.OnButtonPadChange = append(d.callbacks.OnButtonPadChange, callback)
}

func (d *DualSense) OnButtonMuteChange(callback func(bool)) {
	d.callbacks.OnButtonMuteChange = append(d.callbacks.OnButtonMuteChange, callback)
}

func (d *DualSense) OnButtonLeftFunctionChange(callback func(bool)) {
	d.callbacks.OnButtonLeftFunctionChange = append(d.callbacks.OnButtonLeftFunctionChange, callback)
}

func (d *DualSense) OnButtonRightFunctionChange(callback func(bool)) {
	d.callbacks.OnButtonRightFunctionChange = append(d.callbacks.OnButtonRightFunctionChange, callback)
}

func (d *DualSense) OnButtonLeftPaddleChange(callback func(bool)) {
	d.callbacks.OnButtonLeftPaddleChange = append(d.callbacks.OnButtonLeftPaddleChange, callback)
}

func (d *DualSense) OnButtonRightPaddleChange(callback func(bool)) {
	d.callbacks.OnButtonRightPaddleChange = append(d.callbacks.OnButtonRightPaddleChange, callback)
}

func (d *DualSense) OnAngularVelocityXChange(callback func(int16)) {
	d.callbacks.OnAngularVelocityXChange = append(d.callbacks.OnAngularVelocityXChange, callback)
}

func (d *DualSense) OnAngularVelocityZChange(callback func(int16)) {
	d.callbacks.OnAngularVelocityZChange = append(d.callbacks.OnAngularVelocityZChange, callback)
}

func (d *DualSense) OnAngularVelocityYChange(callback func(int16)) {
	d.callbacks.OnAngularVelocityYChange = append(d.callbacks.OnAngularVelocityYChange, callback)
}

func (d *DualSense) OnAccelerometerXChange(callback func(int16)) {
	d.callbacks.OnAccelerometerXChange = append(d.callbacks.OnAccelerometerXChange, callback)
}

func (d *DualSense) OnAccelerometerYChange(callback func(int16)) {
	d.callbacks.OnAccelerometerYChange = append(d.callbacks.OnAccelerometerYChange, callback)
}

func (d *DualSense) OnAccelerometerZChange(callback func(int16)) {
	d.callbacks.OnAccelerometerZChange = append(d.callbacks.OnAccelerometerZChange, callback)
}

func (d *DualSense) OnTemperatureChange(callback func(int8)) {
	d.callbacks.OnTemperatureChange = append(d.callbacks.OnTemperatureChange, callback)
}

func (d *DualSense) OnTouchFinger1Change(callback func(TouchFinger)) {
	d.callbacks.OnTouchFinger1Change = append(d.callbacks.OnTouchFinger1Change, callback)
}

func (d *DualSense) OnTouchFinger2Change(callback func(TouchFinger)) {
	d.callbacks.OnTouchFinger2Change = append(d.callbacks.OnTouchFinger2Change, callback)
}

func (d *DualSense) OnTriggerRightStopLocationChange(callback func(uint8)) {
	d.callbacks.OnTriggerRightStopLocationChange = append(d.callbacks.OnTriggerRightStopLocationChange, callback)
}

func (d *DualSense) OnTriggerRightStatusChange(callback func(uint8)) {
	d.callbacks.OnTriggerRightStatusChange = append(d.callbacks.OnTriggerRightStatusChange, callback)
}

func (d *DualSense) OnTriggerLeftStopLocationChange(callback func(uint8)) {
	d.callbacks.OnTriggerLeftStopLocationChange = append(d.callbacks.OnTriggerLeftStopLocationChange, callback)
}

func (d *DualSense) OnTriggerLeftStatusChange(callback func(uint8)) {
	d.callbacks.OnTriggerLeftStatusChange = append(d.callbacks.OnTriggerLeftStatusChange, callback)
}

func (d *DualSense) OnTriggerRightEffectChange(callback func(uint8)) {
	d.callbacks.OnTriggerRightEffectChange = append(d.callbacks.OnTriggerRightEffectChange, callback)
}

func (d *DualSense) OnTriggerLeftEffectChange(callback func(uint8)) {
	d.callbacks.OnTriggerLeftEffectChange = append(d.callbacks.OnTriggerLeftEffectChange, callback)
}

func (d *DualSense) OnPowerPercentChange(callback func(uint8)) {
	d.callbacks.OnPowerPercentChange = append(d.callbacks.OnPowerPercentChange, callback)
}

func (d *DualSense) OnPowerStateChange(callback func(PowerState)) {
	d.callbacks.OnPowerStateChange = append(d.callbacks.OnPowerStateChange, callback)
}

func (d *DualSense) OnPluggedHeadphonesChange(callback func(bool)) {
	d.callbacks.OnPluggedHeadphonesChange = append(d.callbacks.OnPluggedHeadphonesChange, callback)
}

func (d *DualSense) OnPluggedMicChange(callback func(bool)) {
	d.callbacks.OnPluggedMicChange = append(d.callbacks.OnPluggedMicChange, callback)
}

func (d *DualSense) OnMicMutedChange(callback func(bool)) {
	d.callbacks.OnMicMutedChange = append(d.callbacks.OnMicMutedChange, callback)
}

func (d *DualSense) OnPluggedUsbDataChange(callback func(bool)) {
	d.callbacks.OnPluggedUsbDataChange = append(d.callbacks.OnPluggedUsbDataChange, callback)
}

func (d *DualSense) OnPluggedExternalMicChange(callback func(bool)) {
	d.callbacks.OnPluggedExternalMicChange = append(d.callbacks.OnPluggedExternalMicChange, callback)
}

func (d *DualSense) OnHapticLowPassFilterChange(callback func(bool)) {
	d.callbacks.OnHapticLowPassFilterChange = append(d.callbacks.OnHapticLowPassFilterChange, callback)
}
