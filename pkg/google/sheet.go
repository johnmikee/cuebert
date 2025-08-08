package google

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	sheets "google.golang.org/api/sheets/v4"
)

// SheetInfo is the struct that holds the sheet info.
type SheetInfo struct {
	EmployeeID int
	Status     string
	Email      string
	Name       string
	Manager    string
}

var expectedHeaders = []string{"Employee Id", "Status", "Email", "Name", "Manager"}

// GetSheetVals returns a ValueRange of the values in the sheet.
func (c *Client) GetSheetVals(s *sheets.Spreadsheet) (*sheets.ValueRange, error) {
	return c.sheets.Spreadsheets.Values.Get(c.sheetID, getRangeSet(s)).Do()
}

func (c *Client) getSheet() (*sheets.Spreadsheet, error) {
	return c.sheets.Spreadsheets.Get(c.sheetID).Do()
}

// GetSheet returns the spreadsheet sorted.
func (c *Client) GetSheet() []SheetInfo {
	sheet, err := c.getSheet()
	if err != nil {
		c.log.Error().AnErr("error", err).Msg("error getting sheet")
		return nil
	}
	vr, err := c.GetSheetVals(sheet)
	if err != nil {
		c.log.Error().AnErr("error", err).Msg("error getting sheet values")
		return nil
	}

	return c.sortSheetVals(vr)
}

func (c *Client) sortSheetVals(vr *sheets.ValueRange) []SheetInfo {
	var si []SheetInfo
	if vr.Values == nil {
		c.log.Error().Msg("No data found.")
		return nil
	}

	headers := rowStringer(vr.Values[0])
	// compare the headers to the expected headers and exit if they don't match
	equal := true

	for i := 0; i < len(headers); i++ {
		if headers[i] != expectedHeaders[i] {
			equal = false
			break
		}
	}

	if !equal {
		c.log.Error().
			Interface("expected_headers", expectedHeaders).
			Interface("actual_headers", headers).
			Msg("headers do not match")
		return nil
	}

	// range over the rest of the slice after the headers index
	for _, v := range vr.Values[1:] {
		row := rowStringer(v)
		var eid int
		eid, err := strconv.Atoi(row[0])
		if err == nil {
			serviceIDS = append(serviceIDS, eid)
		} else {
			eid = generateID(2000)
		}
		si = append(si, SheetInfo{
			EmployeeID: eid,
			Status:     row[1],
			Email:      row[2],
			Name:       row[3],
			Manager:    row[4],
		})
	}

	return si
}

// getRangeSet makes an assumption there is only one sheet in the spreadsheet
// in our case, there is only one sheet but should that change this should be
// updated to handle multiple sheets ranges.
func getRangeSet(s *sheets.Spreadsheet) string {
	var maxRow int64
	for _, j := range s.Sheets {
		maxRow = j.Properties.GridProperties.RowCount
	}

	return fmt.Sprintf("%d:%d", 1, maxRow)
}

func rowStringer(row []interface{}) []string {
	var r []string
	for _, j := range row {
		r = append(r, strings.TrimSpace(j.(string)))
	}
	return r
}

var serviceIDS = []int{}

func generateID(max int) int {
	// Specify the range for random numbers (e.g., 1 to 100).
	min := 1

	// Generate a random number.
	randomNumber := rand.Intn(max-min+1) + min //nolint:gosec

	// Check if the generated number is not already in the array.
	if !contains(serviceIDS, randomNumber) {
		serviceIDS = append(serviceIDS, randomNumber)
	} else {
		// If it is, generate a new number.
		generateID(max)
	}

	return randomNumber
}

// Helper function to check if an integer is in a slice.
func contains(slice []int, item int) bool {
	for _, val := range slice {
		if val == item {
			return true
		}
	}
	return false
}
