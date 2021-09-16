package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	http.HandleFunc("/api/resources", handler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

type Parameters struct {
	sheetId   string
	sheetName string
}

func NewParameter(query url.Values) (*Parameters, error) {
	sheetId := query.Get("sheetId")
	sheetName := query.Get("sheetName")
	if sheetId == "" || sheetName == "" {
		return nil, fmt.Errorf("sheetId and sheetName is required!!! ")
	}
	return &Parameters{
		sheetId:   sheetId,
		sheetName: sheetName,
	}, nil
}

func response(w http.ResponseWriter, res []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(res)
	if err != nil {
		log.Println(err)
	}
}

func errorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	m, _ := json.Marshal(map[string]interface{}{
		"error": err.Error(),
	})
	_, err = w.Write(m)
	if err != nil {
		log.Println(err)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	params, err := NewParameter(req.URL.Query())
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err)
		return
	}

	res, err := func(ctx context.Context, params *Parameters) ([]byte, error) {
		client, err := NewClient(ctx, params.sheetId, []string{"https://www.googleapis.com/auth/spreadsheets.readonly"}...)
		if err != nil {
			return nil, fmt.Errorf("spreadsheet client initialized failed... : %w", err)
		}

		headers, err := client.Header(ctx, params.sheetName)
		if err != nil {
			return nil, fmt.Errorf("data headers fetch failed... : %w", err)
		}

		values, err := client.AllRows(ctx, params.sheetName)
		if err != nil {
			return nil, fmt.Errorf("data fetch failed... : %w", err)
		}

		mapped := jsonMap(headers, values)
		jb, err := json.Marshal(map[string]interface{}{
			"data": mapped,
		})
		if err != nil {
			return nil, fmt.Errorf("data json marshal failed... : %w", err)
		}
		return jb, nil
	}(ctx, params)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, err)
		return
	}
	response(w, res)
}

var am map[int]string

func LastColumnIndexToRangeChar(lastColumnIndex int) string {
	alpha := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if am == nil {
		am = make(map[int]string)
		for i, a := range alpha {
			am[i] = string(a)
		}
	}
	if lastColumnIndex < 26 {
		return am[lastColumnIndex]
	}
	// TODO スプシのカラムに合わせた文字列に変換する
	// 26 -> AA
	// 27 -> AB
	// ...
	return ""
}

func jsonMap(header []interface{}, values [][]interface{}) []map[string]interface{} {
	res := make([]map[string]interface{}, len(values)-1)
	for i, row := range values {
		if i == 0 {
			continue
		}
		for j, column := range row {
			if j == 0 {
				res[i-1] = make(map[string]interface{})
			}
			res[i-1][header[j].(string)] = column
		}
	}
	return res
}

const (
	ValueInput = "USER_ENTERED"
)

type Client struct {
	ser           *sheets.Service
	spreadsheetID string
}

func NewClient(ctx context.Context, spreadsheetID string, scopes ...string) (*Client, error) {
	var (
		ts  oauth2.TokenSource
		err error
	)
	if os.Getenv("GCP_PROJECT") == "" {
		ts, err = tokenSourceFromJSON(ctx, os.Getenv("SHEET_USER_CREDENTIALS"), scopes...)
	} else {
		ts, err = tokenSource(ctx, scopes...)
	}

	ser, err := sheets.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve Sheets client: %w ", err)
	}
	return &Client{
		ser:           ser,
		spreadsheetID: spreadsheetID,
	}, nil
}

func (c *Client) Header(ctx context.Context, sheetName string) ([]interface{}, error) {
	colNum, err := c.CountColumns(ctx, sheetName, true)
	if err != nil {
		return nil, err
	}
	vr, err := c.ser.Spreadsheets.Values.Get(c.spreadsheetID, fmt.Sprintf("%s!A1:%s", sheetName, LastColumnIndexToRangeChar(colNum))).Do()
	if err != nil {
		return nil, err
	}
	return vr.Values[0], nil
}

func (c *Client) AllRows(ctx context.Context, sheetName string) ([][]interface{}, error) {
	colNum, err := c.CountColumns(ctx, sheetName, true)
	if err != nil {
		return nil, err
	}
	vr, err := c.ser.Spreadsheets.Values.Get(c.spreadsheetID, fmt.Sprintf("%s!A1:%s", sheetName, LastColumnIndexToRangeChar(colNum))).Do()
	if err != nil {
		return nil, err
	}
	return vr.Values, nil
}

func (c *Client) CountColumns(ctx context.Context, sheetName string, headerExists bool) (int, error) {
	sheet, err := c.ser.Spreadsheets.Get(c.spreadsheetID).Do()
	if err != nil {
		return 0, fmt.Errorf("get failed. %w", err)
	}
	res := 0
	for _, s := range sheet.Sheets {
		if s.Properties.Title == sheetName {
			res = int(s.Properties.GridProperties.ColumnCount)
		}
	}
	if headerExists {
		res--
	}
	return res, nil
}

func tokenSourceFromJSON(ctx context.Context, filePath string, scopes ...string) (oauth2.TokenSource, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v ", err)
	}

	cred, err := google.CredentialsFromJSON(ctx, b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v ", err)
	}
	return cred.TokenSource, nil
}

func tokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
	ts, err := google.DefaultTokenSource(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v ", err)
	}
	return ts, nil
}
