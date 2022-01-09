package main

import "testing"

func TestGetFolderName(t *testing.T) {
	tests := []struct {
		fileName           string
		expectedFolderName string
		errExpected        bool
	}{
		{"1234", "", true}, // No matcher should exist for this, so expect an error.
		{"IMG_20210222_213525.jpg", "2021-02-22", false},
		{"VID_20201012_124124.mp4", "2020-10-12", false},
		{"VID_20201012_124124_325_someextrastuff.mp4", "2020-10-12", false},
		{"PXL_20210123_124124.mp4", "2021-01-23", false},
		{"PXL_19891211_124124.jpg", "1989-12-11", false},
		{"C360_2019-07-17-04-02-45-169.jpg", "2019-07-17", false},
		{"C360_2019-07-17-04-02-45-169-12.jpg", "", true},
		{"C360_2019-07-17-169.jpg", "", true},
		{"20170402_1979.jpg", "2017-04-02", false},
		{"20181030_1985.mp4", "2018-10-30", false},
	}

	for _, tt := range tests {
		name, err := getFolderName(tt.fileName)
		if err != nil && !tt.errExpected {
			t.Errorf("Expected no error but received: %s", err)
		}
		if tt.errExpected && err == nil {
			t.Error("Expected error but received none")
		}
		if name != tt.expectedFolderName {
			t.Errorf("got %s, want %s (file name: %s)", name, tt.expectedFolderName, tt.fileName)
		}
	}
}
