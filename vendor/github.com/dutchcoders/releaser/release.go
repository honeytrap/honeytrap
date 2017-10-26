package releaser

type Release struct {
	Assets []struct {
		BrowserDownloadURL string `json:"browser_download_url"`
		ContentType        string `json:"content_type"`
		CreatedAt          string `json:"created_at"`
		DownloadCount      int64  `json:"download_count"`
		ID                 int64  `json:"id"`
		Label              string `json:"label"`
		Name               string `json:"name"`
		Size               int64  `json:"size"`
		State              string `json:"state"`
		UpdatedAt          string `json:"updated_at"`
		Uploader           struct {
			AvatarURL         string `json:"avatar_url"`
			EventsURL         string `json:"events_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			GravatarID        string `json:"gravatar_id"`
			HtmlURL           string `json:"html_url"`
			ID                int64  `json:"id"`
			Login             string `json:"login"`
			OrganizationsURL  string `json:"organizations_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			ReposURL          string `json:"repos_url"`
			SiteAdmin         bool   `json:"site_admin"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			Type              string `json:"type"`
			URL               string `json:"url"`
		} `json:"uploader"`
		URL string `json:"url"`
	} `json:"assets"`
	AssetsURL string `json:"assets_url"`
	Author    struct {
		AvatarURL         string `json:"avatar_url"`
		EventsURL         string `json:"events_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		GravatarID        string `json:"gravatar_id"`
		HtmlURL           string `json:"html_url"`
		ID                int64  `json:"id"`
		Login             string `json:"login"`
		OrganizationsURL  string `json:"organizations_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		ReposURL          string `json:"repos_url"`
		SiteAdmin         bool   `json:"site_admin"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		Type              string `json:"type"`
		URL               string `json:"url"`
	} `json:"author"`
	Body            string `json:"body"`
	CreatedAt       string `json:"created_at"`
	Draft           bool   `json:"draft"`
	HtmlURL         string `json:"html_url"`
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Prerelease      bool   `json:"prerelease"`
	PublishedAt     string `json:"published_at"`
	TagName         string `json:"tag_name"`
	TarballURL      string `json:"tarball_url"`
	TargetCommitish string `json:"target_commitish"`
	UploadURL       string `json:"upload_url"`
	URL             string `json:"url"`
	ZipballURL      string `json:"zipball_url"`
}
