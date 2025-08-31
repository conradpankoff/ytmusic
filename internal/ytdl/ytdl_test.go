package ytdl

// TODO: Implement tests for progress callback functionality
// These tests are disabled as they reference functions not yet implemented

/*
import (
	"testing"
)

func TestGetChannel(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		expected *Channel
		wantErr  bool
	}{
		{
			name: "Valid channel ID",
			id:   "UC-lHJZR3Gqxm24_Vd_AJ5Yw",
			expected: &Channel{
				ID:          "UC-lHJZR3Gqxm24_Vd_AJ5Yw",
				Title:       "The Coding Train",
				Description: "Coding tutorials and challenges.",
				Thumbnails: []struct{ URL string }{
					{URL: "https://yt3.ggpht.com/ytc/AKedOLQsbNKh1h-V-BW0bDX6ee5ZcEoUckL3q6fHByG9eQ=s88-c-k-c0x00ffffff-no-rj"},
				},
				Subscribers:    "2.39M",
				ViewCount:      "312,217,083",
				VideoCount:     "1,184",
				Country:        "US",
				UploadPlaylist: "UULNgu_OupwoeESgtab33CCw",
			},
			wantErr: false,
		},
		{
			name:     "Invalid channel ID",
			id:       "invalid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetChannel(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.expected != nil && *got != *tt.expected {
				t.Errorf("GetChannel() = %v, want %v", got, tt.expected)
			}
		})
	}
}
*/
