// organizepics.go is a tool that assists in organizing pictures/videos in a set
// of appropriately named directories, corresponding to the date the pictures
// were taken. The directories are named in the form YYYY-MM-DD, and are created
// as needed. This tool makes the assumption that the appropriate date is
// encoded in the file name.
//
// Usage:
//  $ organizepics [path_to_directory_with_pictures]

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s [path to picture directory]\n", os.Args[0])
}

// TODO: Consider pulling a lot of this out into a Go library (which can be
// separately tested).

// MediaFileMatcher represents an element capable of parsing date information
// for a given file name. Each MediaFileMatcher is specifically intended to
// handle certain file types and is capable of parsing date information from
// those applicable file names. For example, a MediaFileMatcher intended to
// match image files of format "IMG_YYYYMMDD_*.jpg" is capable of parsing out
// the intended date in format YYYY-MM-DD but is unable to reliably do so for
// other file formats it is not designed for.
type MediaFileMatcher struct {
	supportedRegexps []*regexp.Regexp
	parseDate        func(s string) (year, month, day string)
}

// MatchFileName determines whether or not the MediaFileMatcher supports the
// file with name givey by the parameter s.
func (m *MediaFileMatcher) MatchFileName(s string) bool {
	for _, re := range m.supportedRegexps {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}

// ParseFormattedDate parses the provided string `s` into a string of format
// YYYY-MM-DD. Note that calling ParseFormattedDate on file name for which
// MatchFileName returns false is not deterministic and would most likely not
// provide meaningful results.
//
// Suggested usage pattern:
//
//	if (matcher.MatchFileName(s)) {
//	  formattedDate := matcher.ParseFormattedDate(s)
//	  // Do something with `formattedDate`.
//	}
func (m *MediaFileMatcher) ParseFormattedDate(s string) string {
	year, month, day := m.parseDate(s)
	return fmt.Sprintf("%s-%s-%s", year, month, day)
}

var mediaMatchers = []*MediaFileMatcher{
	{
		// Intended to match files of format
		//  - IMG_YYYYMMDD_NUMBER.jpg
		//  - VID_YYYYMMDD_NUMBER.mp4
		//  - PXL_YYYYMMDD_NUMBER.{jpg,mp4}
		supportedRegexps: []*regexp.Regexp{
			regexp.MustCompile(`IMG_\d{8}_.+jpg$`),
			regexp.MustCompile(`VID_\d{8}_.+mp4$`),
			regexp.MustCompile(`PXL_\d{8}_.+jpg$`),
			regexp.MustCompile(`PXL_\d{8}_.+mp4$`),
		},
		parseDate: func(s string) (year, month, day string) {
			date := strings.Split(s, "_")[1]
			year = date[:4]
			month = date[4:6]
			day = date[6:]
			return
		},
	},
	{
		// Intended to match C360_YYYY-MM-DD-hh-mm-ss-mmm.jpg.
		supportedRegexps: []*regexp.Regexp{
			regexp.MustCompile(`^C360_\d{4}-\d\d-\d\d-\d\d-\d\d-\d\d-\d{3}\.jpg`),
		},
		parseDate: func(s string) (year, month, day string) {
			date := strings.Split(s, "_")[1]
			dateVals := strings.Split(date, "-")
			year = dateVals[0]
			month = dateVals[1]
			day = dateVals[2]
			return
		},
	},
	{
		// Intended to match files of format
		//	- YYYYMMDD_NUMBER.jpg
		//	- YYYYMMDD_NUMBER.mp4
		supportedRegexps: []*regexp.Regexp{
			regexp.MustCompile(`^\d{8}_.+jpg$`),
			regexp.MustCompile(`^\d{8}_.+mp4$`),
		},
		parseDate: func(s string) (year, month, day string) {
			date := strings.Split(s, "_")[0]
			year = date[:4]
			month = date[4:6]
			day = date[6:]
			return
		},
	},
	{
		// Intended to match files of format
		//	- Screenshot_YYYYMMDD_*.jpg
		supportedRegexps: []*regexp.Regexp{
			regexp.MustCompile(`^Screenshot_\d{8}_.+jpg$`),
		},
		parseDate: func(s string) (year, month, day string) {
			date := strings.Split(s, "_")[1]
			year = date[:4]
			month = date[4:6]
			day = date[6:]
			return
		},
	},
}

// organizePics accepts a directory name and organizes all recognized files
// (images, videos) into appropriate directories.
// TODO: Consider accepting a slice of os.FileInfo to reduce dependency on file
// system and make it easier to test (although that might not be entirely
// easy).
func organizePics(dirName string) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			fileName := file.Name()

			destDirName, err := getFolderName(fileName)
			if err != nil {
				log.Print(err)
				continue
			}
			destPath := filepath.Join(dirName, destDirName)

			// Check if dir exists, making it if it doesn't.
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				// Now create it.
				err := os.Mkdir(destPath, 0700)
				if err != nil {
					log.Fatalf("unable to mkdir %q: %v", destPath, err)
				}
			}

			// Ensure intended path doesn't already exist.
			destFilePath := filepath.Join(destPath, fileName)
			if _, err := os.Stat(destFilePath); err == nil {
				// File exists, and that's not okay. Probably safer not to
				// overwrite the existing file. Log a warning and continue to
				// the next file; the user can decide what to do.
				log.Printf("Destination file %q already exists in %q\n", fileName, destPath)
				continue
			}
			// Move file to new location.
			os.Rename(filepath.Join(dirName, fileName), destFilePath)
		}
	}
}

// getFolderName accepts a file name and returns name that would be appropriate
// to store that given file. If no such folder name can be determined then this
// function returns a non-nil error.
func getFolderName(fileName string) (string, error) {
	for _, matcher := range mediaMatchers {
		if matcher.MatchFileName(fileName) {
			return matcher.ParseFormattedDate(fileName), nil
		}
	}
	return "", fmt.Errorf("no matcher found for %q", fileName)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Incorrect number of args to program. Expected 1, received %d\n", flag.NArg())
		usage()
		os.Exit(1)
	}

	dirName := flag.Arg(0)
	dir, err := os.Stat(dirName)
	if err != nil {
		log.Fatalf("Error stating path: %s\n", err)
	}
	if !dir.IsDir() {
		log.Fatalf("Provider path is not a directory: %s", dirName)
	}

	organizePics(dirName)
}
