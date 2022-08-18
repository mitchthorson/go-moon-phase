package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	"log"
)

// Struct to store an API response from https://aa.usno.navy.mil/data/api#phase
// https://aa.usno.navy.mil/api/moon/phases/date
type MoonApiResponse struct {
	Apiversion string     `json:"apiversion"`
	Day int	              `json:"day"`
	Month int             `json:"month"`
	Year int              `json:"year"`
	Numphases int         `json:"numphases"`
	Phasedata []MoonPhase `json:"phasedata"`
}

// Struct to store a Moon phase from API Data
type MoonPhase struct {
	Day int	      `json:"day"`
	Month int     `json:"month"`
	Year int      `json:"year"`
	Phase string  `json:"phase"`
	Time string   `json:"time"`
}

const dateFormat string = "2006-01-02"

// fetches data from the Astronomical Applications Department of the U.S. navy
// https://aa.usno.navy.mil/
// https://aa.usno.navy.mil/data/api#phase
// Note: the API docs and the API itself asks for dates like 01/02/2006, but really it wants 2006-01-02
func getMoonData(date string, numPhases int) []MoonPhase {
	apiUrl := fmt.Sprintf("https://aa.usno.navy.mil/api/moon/phases/date?date=%s&nump=%d", date, numPhases)
	resp, err := http.Get(apiUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var moonApiResponse = MoonApiResponse{}
	err = json.Unmarshal(body, &moonApiResponse)
	if err != nil {
		log.Fatal(err)
	}
	return moonApiResponse.Phasedata
}

// returns the location for local timezone
func getLocalTimeLocation() *time.Location {
	now := time.Now()
	locationName := now.Location().String()
	location, err := time.LoadLocation(locationName)
	if err != nil {
		log.Fatal(err)
	}
	return location
}

// returns a Time from a MoonPhase struct
func getPhaseDate(phase MoonPhase) time.Time {
	location := getLocalTimeLocation()
	phaseDate := time.Date(phase.Year, time.Month(phase.Month) , phase.Day, 0, 0, 0, 0, location)
	return phaseDate
}

// give me the moon phase for a given time
// fun to say "a slice of moon phase"
func getCurrentPhase(now time.Time, recentData []MoonPhase) string{
	for i, phase := range recentData {
		phaseDate := getPhaseDate(phase)
		// if phase is in future
		if (phaseDate.After(now)) {
			if (i < 1) {
				log.Fatal("date range of recent data doesn't have enough history")
			}
			//store reference to previous phase
			previousPhase := recentData[i - 1]
			previousPhaseDate := getPhaseDate(previousPhase)
			// if phase is within a day of previous, return previousPhase
			// days := t2.Sub(t1).Hours() / 24
			if ( now.Sub(previousPhaseDate).Hours() / 24 < 2 ) {
				return previousPhase.Phase
			}
			// if phase is within a day of next phase, return it
			if ( phaseDate.Sub(now).Hours() / 24 < 2 ) {
				return phase.Phase
			}
			// between New Moon and First Quarter ? Waxing Crescent
			if (previousPhase.Phase == "New Moon" && phase.Phase == "First Quarter") {
				return "Waxing Crescent"
			}
			// between First Quarter and Full Moon ? Waxing Gibbous
			if (previousPhase.Phase == "First Quarter" && phase.Phase == "Full Moon") {
				return "Waxing Crescent"
			}
			// between Full Moon and Last Quarter ? Waning Gibbous
			if (previousPhase.Phase == "Full Moon" && phase.Phase == "Last Quarter") {
				return "Waning Gibbous"
			}
			// between Last Quarter and New Moon ? Waning Crescent
			if (previousPhase.Phase == "Last Quarter" && phase.Phase == "New Moon") {
				return "Waning Crescent"
			}
		}
	}
	return "Error parsing phase"
}

// get the the given date minus the offset parameter number of days
func getOffsetDate(now time.Time, offset int) time.Time{
	offsetDate := now.AddDate(0, 0, -offset)
	return offsetDate
}

// Get the moon's phase for a given date
func getPhaseForDate(date time.Time) string {
	startTime := getOffsetDate(date, 7)
	dateStr := startTime.Format(dateFormat)
	recentData := getMoonData(dateStr, 4)
	return getCurrentPhase(date, recentData)
}

// Return output as string, either plaintext or convert to emoji
func getOutput(phase string, plaintext bool) string {
	if (plaintext) {
		return strings.Trim(phase, "\n")
	}

	emojiMap := map[string]string{
		"New Moon": "🌑",
		"Waxing Crescent": "🌒",
		"First Quarter": "🌓",
		"Waxing Gibbous": "🌔",
		"Full Moon": "🌕",
		"Waning Gibbous": "🌖",
		"Last Quarter": "🌗",
		"Waning Crescent": "🌘",
	}

	return emojiMap[strings.Trim(phase, "\n")]
}

// loads content of save file or returns nil?
func loadSaveFile(saveFilePath string) string {
	var output string
	content, err := ioutil.ReadFile(saveFilePath)
	if err == nil {
		output = string(content)
	}
	return output
}

// parses content of save file to time and phase string
func parseSaveFile(content string) (time.Time, string) {
	currentLocation := getLocalTimeLocation()
	splitContent := strings.Split(content, ",")
	saveTime, err := time.ParseInLocation(dateFormat, splitContent[0], currentLocation)
	if err != nil {
		log.Fatal(err)
	}
	savePhase := splitContent[1]
	return saveTime, savePhase
}

// saves current phase to local file
func savePhaseToFile(date time.Time, phase string, saveFilePath string) {
	saveText := []byte(fmt.Sprintf("%s,%s\n", date.Format(dateFormat), phase))
	err := os.WriteFile(saveFilePath, saveText, 0666)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// current date
	today := time.Now()
	// ger user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	// default save file location is ~/.moonphase
	defaultSaveFile := fmt.Sprintf("%s/%s", homeDir, ".moonphase") 
	// prefer plaintext or emoji output? defualts to emoji
	plaintextFlag := flag.Bool("plaintext", false, "Get result in plain english.")
	// output file to cache daily phase info, dafaults to $HOME/.moonphase
	saveFileFlag := flag.String("savefile", defaultSaveFile, "File to persist output to")
	// store passed date, default to current date in current time one
	var dateFlag string
	flag.StringVar(&dateFlag, "date", today.Format(dateFormat), "Date to get phase for, defaults to today")
	// need to parse the flags
	flag.Parse()
	// local timezone
	currentLocation := getLocalTimeLocation()
	// convert date string to real date
	dateFromFlag, err := time.ParseInLocation(dateFormat, dateFlag, currentLocation)
	if err != nil {
		log.Fatal(err)
	}
	// read from the save file location and check for cached moon phase
	saveFileContent := loadSaveFile(*saveFileFlag)
	if (saveFileContent != "") {
		saveDate, savePhase := parseSaveFile(saveFileContent)
		// if the save file contains the phase for the requested date, print the phase and exit
		if (saveDate == dateFromFlag) {
			fmt.Println(getOutput(savePhase, *plaintextFlag))
			os.Exit(0)
		}
	}
	// otherwise fetch a new phase from the API for the given date
	phase := getPhaseForDate(dateFromFlag)
	phaseOutput := getOutput(phase, *plaintextFlag)
	// cache result to local save file
	savePhaseToFile(dateFromFlag, phase, *saveFileFlag) 
	// print output
	fmt.Println(phaseOutput)
}
