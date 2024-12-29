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
		err = d.writeSetStateData(defaultSetStateData)
	} else {
		err = d.writeSetStateData(*initialSetStateData)
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
	if d.getStateData.TouchData.TouchFinger1 != previousGetStateData.TouchData.TouchFinger1 {
		for _, callback := range d.callbacks.OnTouchFinger1Change {
			callback(d.getStateData.TouchData.TouchFinger1)
		}
	}
	if d.getStateData.TouchData.TouchFinger2 != previousGetStateData.TouchData.TouchFinger2 {
		for _, callback := range d.callbacks.OnTouchFinger2Change {
			callback(d.getStateData.TouchData.TouchFinger2)
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

func (d *DualSense) writeSetStateData(setStateData SetStateData) error {
	packedUSBReportOut, err := packUSBReportOut(setStateData)
	if err != nil {
		return fmt.Errorf("packUSBReportOut: error trying to pack DualSense controller output report: %w", err)
	}
	_, err = d.device.Write(packedUSBReportOut)
	if err != nil {
		err = fmt.Errorf("device.Write: error trying to write DualSense controller output report: %w", err)
	} else {
		d.setStateData = setStateData
	}
	return err
}

func (d *DualSense) GetInStateData() USBGetStateData {
	return d.getStateData
}

func (d *DualSense) GetOutStateData() SetStateData {
	return d.setStateData
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

func (d *DualSense) SetStateData(setStateData SetStateData) error {
	if d.setStateData != setStateData {
		d.setStateDataMu.Lock()
		err := d.writeSetStateData(setStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error writing new setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetEnableRunbleEmulation(enable bool) error {
	if d.setStateData.EnableRumbleEmulation != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.EnableRumbleEmulation = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating EnableRunbleEmulation in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetUseRumbleNotHaptics(useRumbleNotHaptics bool) error {
	if d.setStateData.UseRumbleNotHaptics != useRumbleNotHaptics {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.UseRumbleNotHaptics = useRumbleNotHaptics
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating UseRumbleNotHaptics in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowRightTriggerFFB(allow bool) error {
	if d.setStateData.AllowRightTriggerFFB != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowRightTriggerFFB = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowRightTriggerFFB in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowLeftTriggerFFB(allow bool) error {
	if d.setStateData.AllowLeftTriggerFFB != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowLeftTriggerFFB = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowLeftTriggerFFB in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowHeadphoneVolume(allow bool) error {
	if d.setStateData.AllowHeadphoneVolume != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowHeadphoneVolume = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowHeadphoneVolume in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowSpeakerVolume(allow bool) error {
	if d.setStateData.AllowSpeakerVolume != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowSpeakerVolume = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowSpeakerVolume in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowMicVolume(allow bool) error {
	if d.setStateData.AllowMicVolume != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowMicVolume = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowMicVolume in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowAudioControl(allow bool) error {
	if d.setStateData.AllowAudioControl != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowAudioControl = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowAudioControl in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowMuteLight(allow bool) error {
	if d.setStateData.AllowMuteLight != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowMuteLight = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowMuteLight in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowAudioMute(allow bool) error {
	if d.setStateData.AllowAudioMute != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowAudioMute = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowAudioMute in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowLedColor(allow bool) error {
	if d.setStateData.AllowLedColor != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowLedColor = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowLedColor in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetResetLights(reset bool) error {
	if d.setStateData.ResetLights != reset {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.ResetLights = reset
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating ResetLights in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowPlayerIndicators(allow bool) error {
	if d.setStateData.AllowPlayerIndicators != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowPlayerIndicators = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowPlayerIndicators in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowHapticLowPassFilter(allow bool) error {
	if d.setStateData.AllowHapticLowPassFilter != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowHapticLowPassFilter = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowHapticLowPassFilter in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowMotorPowerLevel(allow bool) error {
	if d.setStateData.AllowMotorPowerLevel != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowMotorPowerLevel = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowMotorPowerLevel in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowAudioControl2(allow bool) error {
	if d.setStateData.AllowAudioControl2 != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowAudioControl2 = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowAudioControl2 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetRumbleEmulationRight(value uint8) error {
	if d.setStateData.RumbleEmulationRight != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.RumbleEmulationRight = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating RumbleEmulationRight in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetRumbleEmulationLeft(value uint8) error {
	if d.setStateData.RumbleEmulationLeft != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.RumbleEmulationLeft = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating RumbleEmulationLeft in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetVolumeHeadphones(value uint8) error {
	if d.setStateData.VolumeHeadphones != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.VolumeHeadphones = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating VolumeHeadphones in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetVolumeSpeaker(value uint8) error {
	if d.setStateData.VolumeSpeaker != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.VolumeSpeaker = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating VolumeSpeaker in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetVolumeMic(value uint8) error {
	if d.setStateData.VolumeMic != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.VolumeMic = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating VolumeMic in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetMicSelect(value MicSelectType) error {
	if d.setStateData.MicSelect != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.MicSelect = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating MicSelect in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetEchoCancelEnable(enable bool) error {
	if d.setStateData.EchoCancelEnable != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.EchoCancelEnable = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating EchoCancelEnable in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetNoiseCancelEnable(enable bool) error {
	if d.setStateData.NoiseCancelEnable != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.NoiseCancelEnable = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating NoiseCancelEnable in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetOutputPathSelect(value uint8) error {
	if d.setStateData.OutputPathSelect != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.OutputPathSelect = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating OutputPathSelect in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetInputPathSelect(value uint8) error {
	if d.setStateData.InputPathSelect != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.InputPathSelect = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating InputPathSelect in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetMuteLight(value MuteLightMode) error {
	if d.setStateData.MuteLight != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.MuteLight = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating MuteLight in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetTouchPowerSave(enable bool) error {
	if d.setStateData.TouchPowerSave != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.TouchPowerSave = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating TouchPowerSave in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetMotionPowerSave(enable bool) error {
	if d.setStateData.MotionPowerSave != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.MotionPowerSave = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating MotionPowerSave in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetHapticPowerSave(enable bool) error {
	if d.setStateData.HapticPowerSave != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.HapticPowerSave = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating HapticPowerSave in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAudioPowerSave(enable bool) error {
	if d.setStateData.AudioPowerSave != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AudioPowerSave = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AudioPowerSave in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetMicMute(enable bool) error {
	if d.setStateData.MicMute != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.MicMute = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating MicMute in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetSpeakerMute(enable bool) error {
	if d.setStateData.SpeakerMute != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.SpeakerMute = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating SpeakerMute in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetHeadphoneMute(enable bool) error {
	if d.setStateData.HeadphoneMute != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.HeadphoneMute = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating HeadphoneMute in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetHapticMute(enable bool) error {
	if d.setStateData.HapticMute != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.HapticMute = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating HapticMute in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetRightTriggerFFB(params [11]uint8) error {
	if d.setStateData.RightTriggerFFB != params {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.RightTriggerFFB = params
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating RightTriggerFFB in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLeftTriggerFFB(params [11]uint8) error {
	if d.setStateData.LeftTriggerFFB != params {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LeftTriggerFFB = params
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LeftTriggerFFB in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetTriggerMotorPowerReduction(level uint8) error {
	if d.setStateData.TriggerMotorPowerReduction != level {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.TriggerMotorPowerReduction = level
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating TriggerMotorPowerReduction in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetRumbleMotorPowerReduction(level uint8) error {
	if d.setStateData.RumbleMotorPowerReduction != level {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.RumbleMotorPowerReduction = level
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating RumbleMotorPowerReduction in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetSpeakerCompPreGain(gain uint8) error {
	if d.setStateData.SpeakerCompPreGain != gain {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.SpeakerCompPreGain = gain
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating SpeakerCompPreGain in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetBeamformingEnable(enable bool) error {
	if d.setStateData.BeamformingEnable != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.BeamformingEnable = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating BeamformingEnable in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowLightBrightnessChange(allow bool) error {
	if d.setStateData.AllowLightBrightnessChange != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowLightBrightnessChange = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowLightBrightnessChange in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetAllowColorLightFadeAnimation(allow bool) error {
	if d.setStateData.AllowColorLightFadeAnimation != allow {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.AllowColorLightFadeAnimation = allow
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating AllowColorLightFadeAnimation in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetEnableImprovedRumbleEmulation(enable bool) error {
	if d.setStateData.EnableImprovedRumbleEmulation != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.EnableImprovedRumbleEmulation = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating EnableImprovedRumbleEmulation in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLightFadeAnimation(animation LightFadeAnimation) error {
	if d.setStateData.LightFadeAnimation != animation {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LightFadeAnimation = animation
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LightFadeAnimation in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLightBrightness(brightness LightBrightness) error {
	if d.setStateData.LightBrightness != brightness {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LightBrightness = brightness
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LightBrightness in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLight1(enable bool) error {
	if d.setStateData.PlayerLight1 != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLight1 = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLight1 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLight2(enable bool) error {
	if d.setStateData.PlayerLight2 != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLight2 = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLight2 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLight3(enable bool) error {
	if d.setStateData.PlayerLight3 != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLight3 = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLight3 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLight4(enable bool) error {
	if d.setStateData.PlayerLight4 != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLight4 = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLight4 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLight5(enable bool) error {
	if d.setStateData.PlayerLight5 != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLight5 = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLight5 in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetPlayerLightFade(enable bool) error {
	if d.setStateData.PlayerLightFade != enable {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.PlayerLightFade = enable
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating PlayerLightFade in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLedRed(value uint8) error {
	if d.setStateData.LedRed != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LedRed = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LedRed in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLedGreen(value uint8) error {
	if d.setStateData.LedGreen != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LedGreen = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LedGreen in setStateData: %w", err)
		}
	}
	return nil
}

func (d *DualSense) SetLedBlue(value uint8) error {
	if d.setStateData.LedBlue != value {
		d.setStateDataMu.Lock()
		newSetStateData := d.setStateData
		newSetStateData.LedBlue = value
		err := d.writeSetStateData(newSetStateData)
		d.setStateDataMu.Unlock()
		if err != nil {
			return fmt.Errorf("error updating LedBlue in setStateData: %w", err)
		}
	}
	return nil
}
