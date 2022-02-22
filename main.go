package main

import (
	"fmt"
	"time"
	"encoding/json"
	"net/http"
	"io/ioutil"
)

type MoonApiResponse struct {
	Apiversion string     `json:"apiversion"`
	Day int	              `json:"day"`
	Month int             `json:"month"`
	Year int              `json:"year"`
	Numphases int         `json:"numphases"`
	Phasedata []MoonPhase `json:"phasedata"`
}

type MoonPhase struct {
	Day int `json:"day"`
	Month int `json:"month"`
	Year int `json:"year"`
	Phase string `json:"phase"`
	Time string `json:"time"`
}

func getMoonData(date string, numPhases int) {
	apiUrl := fmt.Sprintf("https://aa.usno.navy.mil/api/moon/phases/date?date=%s&nump=%d", date, numPhases)
	resp, err := http.Get(apiUrl)
	if err != nil {
		panic("error getting data")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic("error reding response")
	}
	var moonApiResponse = new(MoonApiResponse)
	err = json.Unmarshal(body, &moonApiResponse)
	if err != nil {
		panic("error converting data")
	}
	for _,phase := range moonApiResponse.Phasedata {
		fmt.Printf("\nDate: %d/%d/%d\nPhase: %s\n", phase.Month, phase.Day, phase.Year, phase.Phase)
	}
}

// get the current date minus the offset parameter number of days
func getDate(offset int) time.Time{
	now := time.Now()
	offsetDate := now.AddDate(0, 0, -offset)
	return offsetDate
}

func main() {
	now := getDate(7)
	dateStr := now.Format("1/2/2006")
	getMoonData(dateStr, 4)
}
