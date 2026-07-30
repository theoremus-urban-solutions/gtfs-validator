package notice

// Notices for the canonical rules about feed_info.txt, agency.txt and the
// presence of recommended files. Codes and severities match
// <https://gtfs-validator.mobilitydata.org/rules.html>; the checks are our own
// implementations.

// MissingFeedInfoDateNotice reports a feed_info.txt row giving one of
// feed_start_date and feed_end_date but not the other. Both are optional, but
// half a range does not say when the feed stops being valid.
type MissingFeedInfoDateNotice struct {
	*BaseNotice
}

func NewMissingFeedInfoDateNotice(rowNumber int, fieldName string) *MissingFeedInfoDateNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"fieldName":    fieldName,
	}
	return &MissingFeedInfoDateNotice{
		BaseNotice: NewBaseNotice("missing_feed_info_date", WARNING, context),
	}
}

// MissingFeedContactEmailAndUrlNotice reports a feed_info.txt row with neither
// feed_contact_email nor feed_contact_url. Without one of them a consumer who
// finds a defect in the feed has no way to report it.
type MissingFeedContactEmailAndUrlNotice struct {
	*BaseNotice
}

func NewMissingFeedContactEmailAndUrlNotice(rowNumber int) *MissingFeedContactEmailAndUrlNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
	}
	return &MissingFeedContactEmailAndUrlNotice{
		BaseNotice: NewBaseNotice("missing_feed_contact_email_and_url", WARNING, context),
	}
}

// MissingRecommendedFileNotice reports a file that the spec recommends but does
// not require, and that the archive does not contain.
type MissingRecommendedFileNotice struct {
	*BaseNotice
}

func NewMissingRecommendedFileNotice(filename string) *MissingRecommendedFileNotice {
	context := map[string]interface{}{
		"filename": filename,
	}
	return &MissingRecommendedFileNotice{
		BaseNotice: NewBaseNotice("missing_recommended_file", WARNING, context),
	}
}

// InconsistentAgencyLangNotice reports an agency declaring a different
// agency_lang from the other agencies in the same feed. One feed describes one
// body of text, so the agencies should agree on the language it is written in.
type InconsistentAgencyLangNotice struct {
	*BaseNotice
}

func NewInconsistentAgencyLangNotice(rowNumber int, expected string, actual string) *InconsistentAgencyLangNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"expected":     expected,
		"actual":       actual,
	}
	return &InconsistentAgencyLangNotice{
		BaseNotice: NewBaseNotice("inconsistent_agency_lang", WARNING, context),
	}
}

// FeedInfoLangAndAgencyLangMismatchNotice reports feed_info.feed_lang naming a
// different language from an agency's agency_lang. The two describe the same
// text, so a disagreement leaves consumers guessing which to translate from.
//
// A feed_lang of "mul" is not a mismatch: it declares the feed multilingual,
// in which case no single agency_lang can match it.
type FeedInfoLangAndAgencyLangMismatchNotice struct {
	*BaseNotice
}

func NewFeedInfoLangAndAgencyLangMismatchNotice(rowNumber int, agencyID string, agencyName string, agencyLang string, feedLang string) *FeedInfoLangAndAgencyLangMismatchNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"agencyId":     agencyID,
		"agencyName":   agencyName,
		"agencyLang":   agencyLang,
		"feedLang":     feedLang,
	}
	return &FeedInfoLangAndAgencyLangMismatchNotice{
		BaseNotice: NewBaseNotice("feed_info_lang_and_agency_lang_mismatch", WARNING, context),
	}
}

// MissingBikeAllowanceNotice reports a ferry trip that does not say whether
// bikes may be carried. Ferries are the mode where the answer most often
// decides whether a cycling passenger can make the journey at all, so leaving
// bikes_allowed unspecified there withholds information riders act on.
type MissingBikeAllowanceNotice struct {
	*BaseNotice
}

func NewMissingBikeAllowanceNotice(rowNumber int, routeID string, tripID string) *MissingBikeAllowanceNotice {
	context := map[string]interface{}{
		"csvRowNumber": rowNumber,
		"routeId":      routeID,
		"tripId":       tripID,
	}
	return &MissingBikeAllowanceNotice{
		BaseNotice: NewBaseNotice("missing_bike_allowance", WARNING, context),
	}
}
