package dualsense

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/rivo/tview"
	hid "github.com/sstallion/go-hid"
)

func init() {
	hid.Init()
}

func displayStructAsTable(data USBGetStateData, table *tview.Table) {
	table.Clear()

	val := reflect.ValueOf(data)
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldName := typeOfS.Field(i).Name
		fieldValue := val.Field(i).Interface()

		row := table.GetRowCount()
		table.SetCell(row, 0, tview.NewTableCell(fieldName).SetAlign(tview.AlignRight))

		//handle bools
		if b, ok := fieldValue.(bool); ok {
			fieldValue = strconv.FormatBool(b)
		}

		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%v", fieldValue)).SetAlign(tview.AlignLeft))
	}

	table.SetBorder(true).SetTitle("USB Get State Data").SetTitleAlign(tview.AlignLeft)

}

func displayStatus(dualsense *DualSense) {
	app := tview.NewApplication()
	table := tview.NewTable()
	go func() {
		for {
			reportIn := dualsense.GetReportIn()
			app.QueueUpdateDraw(func() {
				displayStructAsTable(reportIn.USBGetStateData, table)
			})
			// time.Sleep(100 * time.Millisecond) // Update every 100ms
		}
	}()
	if err := app.SetRoot(table, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func TestMain(t *testing.T) {
	dualsense, err := NewDualSense()
	if err != nil {
		panic(err)
	}
	defer dualsense.Close()
	displayStatus(dualsense)
}
