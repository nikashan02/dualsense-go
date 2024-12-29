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
	OnTouchDataChange                []func(TouchData)
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
	getStateData     USBGetStateData
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
	if d.getStateData.LeftStickX != previousGetStateData.LeftStickX {
		for _, callback := range d.callbacks.OnLeftStickXChange {
			callback(d.getStateData.LeftStickX)
		}
	}
	if d.getStateData.LeftStickY != previousGetStateData.LeftStickY {
		for _, callback := range d.callbacks.OnLeftStickYChange {
			callback(d.getStateData.LeftStickY)
		}
	}
	if d.getStateData.RightStickX != previousGetStateData.RightStickX {
		for _, callback := range d.callbacks.OnRightStickXChange {
			callback(d.getStateData.RightStickX)
		}
	}
	if d.getStateData.RightStickY != previousGetStateData.RightStickY {
		for _, callback := range d.callbacks.OnRightStickYChange {
			callback(d.getStateData.RightStickY)
		}
	}
	if d.getStateData.TriggerLeft != previousGetStateData.TriggerLeft {
		for _, callback := range d.callbacks.OnTriggerLeftChange {
			callback(d.getStateData.TriggerLeft)
		}
	}
	if d.getStateData.TriggerRight != previousGetStateData.TriggerRight {
		for _, callback := range d.callbacks.OnTriggerRightChange {
			callback(d.getStateData.TriggerRight)
		}
	}
	if d.getStateData.DPad != previousGetStateData.DPad {
		for _, callback := range d.callbacks.OnDPadChange {
			callback(d.getStateData.DPad)
		}
	}
	if d.getStateData.ButtonSquare != previousGetStateData.ButtonSquare {
		for _, callback := range d.callbacks.OnButtonSquareChange {
			callback(d.getStateData.ButtonSquare)
		}
	}
	if d.getStateData.ButtonCross != previousGetStateData.ButtonCross {
		for _, callback := range d.callbacks.OnButtonCrossChange {
			callback(d.getStateData.ButtonCross)
		}
	}
	if d.getStateData.ButtonCircle != previousGetStateData.ButtonCircle {
		for _, callback := range d.callbacks.OnButtonCircleChange {
			callback(d.getStateData.ButtonCircle)
		}
	}
	if d.getStateData.ButtonTriangle != previousGetStateData.ButtonTriangle {
		for _, callback := range d.callbacks.OnButtonTriangleChange {
			callback(d.getStateData.ButtonTriangle)
		}
	}
	if d.getStateData.ButtonL1 != previousGetStateData.ButtonL1 {
		for _, callback := range d.callbacks.OnButtonL1Change {
			callback(d.getStateData.ButtonL1)
		}
	}
	if d.getStateData.ButtonR1 != previousGetStateData.ButtonR1 {
		for _, callback := range d.callbacks.OnButtonR1Change {
			callback(d.getStateData.ButtonR1)
		}
	}
	if d.getStateData.ButtonL2 != previousGetStateData.ButtonL2 {
		for _, callback := range d.callbacks.OnButtonL2Change {
			callback(d.getStateData.ButtonL2)
		}
	}
	if d.getStateData.ButtonR2 != previousGetStateData.ButtonR2 {
		for _, callback := range d.callbacks.OnButtonR2Change {
			callback(d.getStateData.ButtonR2)
		}
	}
	if d.getStateData.ButtonCreate != previousGetStateData.ButtonCreate {
		for _, callback := range d.callbacks.OnButtonCreateChange {
			callback(d.getStateData.ButtonCreate)
		}
	}
	if d.getStateData.ButtonOptions != previousGetStateData.ButtonOptions {
		for _, callback := range d.callbacks.OnButtonOptionsChange {
			callback(d.getStateData.ButtonOptions)
		}
	}
	if d.getStateData.ButtonL3 != previousGetStateData.ButtonL3 {
		for _, callback := range d.callbacks.OnButtonL3Change {
			callback(d.getStateData.ButtonL3)
		}
	}
	if d.getStateData.ButtonR3 != previousGetStateData.ButtonR3 {
		for _, callback := range d.callbacks.OnButtonR3Change {
			callback(d.getStateData.ButtonR3)
		}
	}
	if d.getStateData.ButtonHome != previousGetStateData.ButtonHome {
		for _, callback := range d.callbacks.OnButtonHomeChange {
			callback(d.getStateData.ButtonHome)
		}
	}
	if d.getStateData.ButtonPad != previousGetStateData.ButtonPad {
		for _, callback := range d.callbacks.OnButtonPadChange {
			callback(d.getStateData.ButtonPad)
		}
	}
	if d.getStateData.ButtonMute != previousGetStateData.ButtonMute {
		for _, callback := range d.callbacks.OnButtonMuteChange {
			callback(d.getStateData.ButtonMute)
		}
	}
	if d.getStateData.ButtonLeftFunction != previousGetStateData.ButtonLeftFunction {
		for _, callback := range d.callbacks.OnButtonLeftFunctionChange {
			callback(d.getStateData.ButtonLeftFunction)
		}
	}
	if d.getStateData.ButtonRightFunction != previousGetStateData.ButtonRightFunction {
		for _, callback := range d.callbacks.OnButtonRightFunctionChange {
			callback(d.getStateData.ButtonRightFunction)
		}
	}
	if d.getStateData.ButtonLeftPaddle != previousGetStateData.ButtonLeftPaddle {
		for _, callback := range d.callbacks.OnButtonLeftPaddleChange {
			callback(d.getStateData.ButtonLeftPaddle)
		}
	}
	if d.getStateData.ButtonRightPaddle != previousGetStateData.ButtonRightPaddle {
		for _, callback := range d.callbacks.OnButtonRightPaddleChange {
			callback(d.getStateData.ButtonRightPaddle)
		}
	}
	if d.getStateData.AngularVelocityX != previousGetStateData.AngularVelocityX {
		for _, callback := range d.callbacks.OnAngularVelocityXChange {
			callback(d.getStateData.AngularVelocityX)
		}
	}
	if d.getStateData.AngularVelocityZ != previousGetStateData.AngularVelocityZ {
		for _, callback := range d.callbacks.OnAngularVelocityZChange {
			callback(d.getStateData.AngularVelocityZ)
		}
	}
	if d.getStateData.AngularVelocityY != previousGetStateData.AngularVelocityY {
		for _, callback := range d.callbacks.OnAngularVelocityYChange {
			callback(d.getStateData.AngularVelocityY)
		}
	}
	if d.getStateData.AccelerometerX != previousGetStateData.AccelerometerX {
		for _, callback := range d.callbacks.OnAccelerometerXChange {
			callback(d.getStateData.AccelerometerX)
		}
	}
	if d.getStateData.AccelerometerY != previousGetStateData.AccelerometerY {
		for _, callback := range d.callbacks.OnAccelerometerYChange {
			callback(d.getStateData.AccelerometerY)
		}
	}
	if d.getStateData.AccelerometerZ != previousGetStateData.AccelerometerZ {
		for _, callback := range d.callbacks.OnAccelerometerZChange {
			callback(d.getStateData.AccelerometerZ)
		}
	}
	if d.getStateData.Temperature != previousGetStateData.Temperature {
		for _, callback := range d.callbacks.OnTemperatureChange {
			callback(d.getStateData.Temperature)
		}
	}
	if d.getStateData.TouchData != previousGetStateData.TouchData {
		for _, callback := range d.callbacks.OnTouchDataChange {
			callback(d.getStateData.TouchData)
		}
	}
	if d.getStateData.TriggerRightStopLocation != previousGetStateData.TriggerRightStopLocation {
		for _, callback := range d.callbacks.OnTriggerRightStopLocationChange {
			callback(d.getStateData.TriggerRightStopLocation)
		}
	}
	if d.getStateData.TriggerRightStatus != previousGetStateData.TriggerRightStatus {
		for _, callback := range d.callbacks.OnTriggerRightStatusChange {
			callback(d.getStateData.TriggerRightStatus)
		}
	}
	if d.getStateData.TriggerLeftStopLocation != previousGetStateData.TriggerLeftStopLocation {
		for _, callback := range d.callbacks.OnTriggerLeftStopLocationChange {
			callback(d.getStateData.TriggerLeftStopLocation)
		}
	}
	if d.getStateData.TriggerLeftStatus != previousGetStateData.TriggerLeftStatus {
		for _, callback := range d.callbacks.OnTriggerLeftStatusChange {
			callback(d.getStateData.TriggerLeftStatus)
		}
	}
	if d.getStateData.TriggerRightEffect != previousGetStateData.TriggerRightEffect {
		for _, callback := range d.callbacks.OnTriggerRightEffectChange {
			callback(d.getStateData.TriggerRightEffect)
		}
	}
	if d.getStateData.TriggerLeftEffect != previousGetStateData.TriggerLeftEffect {
		for _, callback := range d.callbacks.OnTriggerLeftEffectChange {
			callback(d.getStateData.TriggerLeftEffect)
		}
	}
	if d.getStateData.PowerPercent != previousGetStateData.PowerPercent {
		for _, callback := range d.callbacks.OnPowerPercentChange {
			callback(d.getStateData.PowerPercent)
		}
	}
	if d.getStateData.PowerState != previousGetStateData.PowerState {
		for _, callback := range d.callbacks.OnPowerStateChange {
			callback(d.getStateData.PowerState)
		}
	}
	if d.getStateData.PluggedHeadphones != previousGetStateData.PluggedHeadphones {
		for _, callback := range d.callbacks.OnPluggedHeadphonesChange {
			callback(d.getStateData.PluggedHeadphones)
		}
	}
	if d.getStateData.PluggedMic != previousGetStateData.PluggedMic {
		for _, callback := range d.callbacks.OnPluggedMicChange {
			callback(d.getStateData.PluggedMic)
		}
	}
	if d.getStateData.MicMuted != previousGetStateData.MicMuted {
		for _, callback := range d.callbacks.OnMicMutedChange {
			callback(d.getStateData.MicMuted)
		}
	}
	if d.getStateData.PluggedUsbData != previousGetStateData.PluggedUsbData {
		for _, callback := range d.callbacks.OnPluggedUsbDataChange {
			callback(d.getStateData.PluggedUsbData)
		}
	}
	if d.getStateData.PluggedExternalMic != previousGetStateData.PluggedExternalMic {
		for _, callback := range d.callbacks.OnPluggedExternalMicChange {
			callback(d.getStateData.PluggedExternalMic)
		}
	}
	if d.getStateData.HapticLowPassFilter != previousGetStateData.HapticLowPassFilter {
		for _, callback := range d.callbacks.OnHapticLowPassFilterChange {
			callback(d.getStateData.HapticLowPassFilter)
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
				previousGetStateData := d.getStateData
				d.getStateData = reportIn.USBGetStateData
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

func (d *DualSense) OnTouchDataChange(callback func(TouchData)) {
	d.callbacks.OnTouchDataChange = append(d.callbacks.OnTouchDataChange, callback)
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
