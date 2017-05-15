package admin

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"encoding/csv"

	"io/ioutil"
	"os"

	"github.com/ponzu-cms/ponzu/management/format"
	"github.com/ponzu-cms/ponzu/system/db"
	"github.com/ponzu-cms/ponzu/system/item"
	"github.com/tidwall/gjson"
)

func exportHandler(res http.ResponseWriter, req *http.Request) {
	// /admin/contents/export?type=Blogpost&format=csv
	q := req.URL.Query()
	t := q.Get("type")
	f := strings.ToLower(q.Get("format"))

	if t == "" || f == "" {
		v, err := Error400()
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		res.WriteHeader(http.StatusBadRequest)
		_, err = res.Write(v)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

	}

	pt, ok := item.Types[t]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	switch f {
	case "csv":
		csv, ok := pt().(format.CSVFormattable)
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		fields := csv.FormatCSV()
		exportCSV(res, req, pt, fields)

	default:
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func exportCSV(res http.ResponseWriter, req *http.Request, pt func() interface{}, fields []string) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "exportcsv-")
	if err != nil {
		log.Println("Failed to create tmp file for CSV export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer os.Remove(filepath.Join(os.TempDir(), tmpFile.Name()))
	defer tmpFile.Close()

	csvBuf := csv.NewWriter(tmpFile)

	t := req.URL.Query().Get("type")

	// get content data and loop through creating a csv row per result
	bb := db.ContentAll(t)

	err = csvBuf.Write(fields)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Println("Failed to write column headers:", fields)
		return
	}

	for row := range bb {
		// unmarshal data and loop over fields
		rowBuf := []string{}

		for _, col := range fields {
			// pull out each field as the column value
			result := gjson.GetBytes(bb[row], col)

			// append it to the buffer
			rowBuf = append(rowBuf, result.String())
		}

		// write row to csv
		err := csvBuf.Write(rowBuf)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Println("Failed to write column headers:", fields)
			return
		}
	}

	// write the buffer to a content-disposition response
	csvB, err := ioutil.ReadAll(tmpFile)
	if err != nil {
		log.Println("Failed to read tmp file for CSV export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	ts := time.Now().Unix()
	disposition := `attachment; filename="export-%s-%d.csv"`

	res.Header().Set("Content-Type", "text/csv")
	res.Header().Set("Content-Disposition", fmt.Sprintf(disposition, t, ts))
	res.Header().Set("Content-Length", fmt.Sprintf("%d", len(csvB)))

	_, err = res.Write(csvB)
	if err != nil {
		log.Println("Failed to write tmp file to response for CSV export:", err)
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
