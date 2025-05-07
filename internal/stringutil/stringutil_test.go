package stringutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var caseConversionTests = []struct {
	pascalCase string
	snakeCase  string
}{
	{"ID", "id"},
	{"ExternalID", "external_id"},
	{"Title", "title"},
	{"MetadataUpdatedAt", "metadata_updated_at"},
	{"ChannelID", "channel_id"},
	{"ExternalChannelID", "external_channel_id"},
	{"VideoCount", "video_count"},
	{"PublishedTime", "published_time"},
	{"ShouldDownload", "should_download"},
	{"CreatedAt", "created_at"},
	{"QueueName", "queue_name"},
	{"Payload", "payload"},
	{"RunAfter", "run_after"},
	{"FailureDelaySeconds", "failure_delay_seconds"},
	{"AttemptsRemaining", "attempts_remaining"},
	{"ReservedAt", "reserved_at"},
	{"ReservedUntil", "reserved_until"},
	{"FinishedAt", "finished_at"},
	{"ErrorMessage", "error_message"},
}

var pluralTests = []struct {
	singular string
	plural   string
}{
	{"ID", "IDs"},
	{"id", "ids"},
	{"ExternalID", "ExternalIDs"},
	{"external_id", "external_ids"},
	{"Title", "Titles"},
	{"title", "titles"},
	// special cases
	{"Fish", "Fish"},
	{"fish", "fish"},
	{"Sheep", "Sheep"},
	{"sheep", "sheep"},
}

func TestPascalToSnake(t *testing.T) {
	for _, tc := range caseConversionTests {
		t.Run(tc.pascalCase, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.snakeCase, PascalToSnake(tc.pascalCase))
		})
	}
}

func BenchmarkPascalToSnake(b *testing.B) {
	for _, tc := range caseConversionTests {
		b.Run(tc.pascalCase, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				PascalToSnake(tc.pascalCase)
			}
		})
	}
}

func TestPlural(t *testing.T) {
	for _, tc := range pluralTests {
		t.Run(tc.singular, func(t *testing.T) {
			a := assert.New(t)
			a.Equal(tc.plural, Plural(tc.singular))
		})
	}
}

func BenchmarkPlural(b *testing.B) {
	for _, tc := range pluralTests {
		b.Run(tc.singular, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Plural(tc.singular)
			}
		})
	}
}
