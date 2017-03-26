package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
)

func main() {
	var (
		yearPaths, yearsErr = filepath.Glob("/mnt/raw/photos/2*")
		maxYear             int
	)

	if yearsErr != nil {
		panic(yearsErr)
	}

	for _, path := range yearPaths {
		year64, intErr := strconv.ParseInt(filepath.Base(path), 10, 32)
		if intErr == nil {
			year := int(year64)

			if year > maxYear {
				maxYear = year
			}
		}
	}

	if maxYear == 0 {
		panic("Unable to get max year")
	}

	var (
		monthPaths, monthsErr = filepath.Glob(fmt.Sprintf("/mnt/raw/photos/%d/*", maxYear))
		maxMonth              int
		maxMonthDir           string
	)

	if monthsErr != nil {
		panic(monthsErr)
	}

	for _, path := range monthPaths {
		base := filepath.Base(path)
		month64, intErr := strconv.ParseInt(base, 10, 32)
		if intErr == nil {
			month := int(month64)

			if month > maxMonth {
				maxMonth = month
				maxMonthDir = base
			}
		}
	}

	// Most recent dir
	currentDir := filepath.Join("/mnt", "raw", "photos", fmt.Sprintf("%d", maxYear), maxMonthDir)

	photosPaths, photosErr := filepath.Glob(filepath.Join(currentDir, "IMG_*.JPG"))

	if photosErr != nil {
		panic(photosErr)
	}

	var (
		maxPhotoId int
		//maxPhotoFn string
		idMatcher = regexp.MustCompile("([0-9]+)")
	)

	for _, photoFn := range photosPaths {
		base := filepath.Base(photoFn)

		match := idMatcher.FindAllString(base, 1)

		id64, idIntErr := strconv.ParseInt(match[0], 10, 32)
		if idIntErr == nil {
			id := int(id64)

			if id > maxPhotoId {
				maxPhotoId = id
				//maxPhotoFn = photoFn
			}

		}
	}

	var result = struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Id    int `json:"id"`
	}{
		Year:  maxYear,
		Month: maxMonth,
		Id:    maxPhotoId,
	}

	j, _ := json.Marshal(result)

	fmt.Println(string(j))
}
