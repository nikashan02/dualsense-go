package dualsense

import (
	"fmt"
	"time"

	hid "github.com/sstallion/go-hid"
)

const (
	DUALSENSE_VENDOR_ID  = 0x054C
	DUALSENSE_PRODUCT_ID = 0x0CE6
	DEFAULT_READ_TIMEOUT = 100 * time.Millisecond
	USB_PACKET_SIZE      = 64
)

type DualSense struct {
	device           *hid.Device
	usbReportIn      USBReportIn
	usbReportInClose chan bool
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
	}
	go dualsense.listenReportIn()
	return dualsense, nil
}

func (d *DualSense) Close() {
	d.usbReportInClose <- true
	d.device.Close()
}

func (d *DualSense) readReportIn() (USBReportIn, error) {
	buffer := make([]byte, USB_PACKET_SIZE)
	_, err := d.device.ReadWithTimeout(buffer, DEFAULT_READ_TIMEOUT)
	if err != nil {
		return USBReportIn{}, fmt.Errorf("device.ReadWithTimeout: error trying to read DualSense controller input report: %w", err)
	}
	reportIn, err := unpackUSBReportIn(buffer)
	if err != nil {
		return USBReportIn{}, fmt.Errorf("unpackUSBReportIn: error trying to unpack DualSense controller input report: %w", err)
	}
	return reportIn, err
}

func (d *DualSense) listenReportIn() {
	for {
		select {
		case <-d.usbReportInClose:
			fmt.Println("NIKASHAN", d.usbReportIn.USBGetStateData.LeftStickX)
			return
		default:
			reportIn, err := d.readReportIn()
			if err == nil {
				d.usbReportIn = reportIn
			}
		}
	}
}

func (d *DualSense) GetReportIn() USBReportIn {
	return d.usbReportIn
}
