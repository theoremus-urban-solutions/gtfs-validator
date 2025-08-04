package schema

// FeedInfo represents feed metadata from feed_info.txt
type FeedInfo struct {
	FeedPublisherName string `csv:"feed_publisher_name"`
	FeedPublisherURL  string `csv:"feed_publisher_url"`
	FeedLang          string `csv:"feed_lang"`
	DefaultLang       string `csv:"default_lang"`
	FeedStartDate     string `csv:"feed_start_date"`
	FeedEndDate       string `csv:"feed_end_date"`
	FeedVersion       string `csv:"feed_version"`
	FeedContactEmail  string `csv:"feed_contact_email"`
	FeedContactURL    string `csv:"feed_contact_url"`
}
