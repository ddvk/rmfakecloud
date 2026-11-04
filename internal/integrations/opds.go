package integrations

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/sirupsen/logrus"
)

const (
	loggerOPDS = "[opds] "

	// OPDS link relations
	relSubsection    = "subsection"
	relAcquisition   = "http://opds-spec.org/acquisition"
	relAcquisitionOW = "http://opds-spec.org/acquisition/open-access"

	// Content types
	opdsNavigationFeed  = "application/atom+xml;profile=opds-catalog;kind=navigation"
	opdsAcquisitionFeed = "application/atom+xml;profile=opds-catalog;kind=acquisition"
	atomFeed            = "application/atom+xml"
)

// OPDS 1.x Atom feed structures
type opdsFeed struct {
	XMLName xml.Name    `xml:"feed"`
	Title   string      `xml:"title"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Links   []opdsLink  `xml:"link"`
	Entries []opdsEntry `xml:"entry"`
}

type opdsEntry struct {
	Title   string     `xml:"title"`
	ID      string     `xml:"id"`
	Updated string     `xml:"updated"`
	Author  opdsAuthor `xml:"author"`
	Content string     `xml:"content"`
	Links   []opdsLink `xml:"link"`
}

type opdsAuthor struct {
	Name string `xml:"name"`
}

type opdsLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr"`
}

// Image link relations
const (
	relImage          = "http://opds-spec.org/image"
	relImageThumbnail = "http://opds-spec.org/image/thumbnail"
)

// OPDSIntegration implements read-only OPDS 1.x catalog support
type OPDSIntegration struct {
	feedURL string
	headers []model.HeaderConfig
	client  *http.Client
}

// newOPDS creates a new OPDS integration
func newOPDS(in model.IntegrationConfig) *OPDSIntegration {
	logrus.Tracef("%snew client, feedURL: %s, headers: %d", loggerOPDS, in.FeedURL, len(in.Headers))
	return &OPDSIntegration{
		feedURL: in.FeedURL,
		headers: in.Headers,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with configured headers
func (o *OPDSIntegration) doRequest(targetURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}

	// Add configured headers for authentication
	for _, h := range o.headers {
		req.Header.Set(h.Name, h.Value)
	}

	return o.client.Do(req)
}

// fetchFeed retrieves and parses an OPDS feed
func (o *OPDSIntegration) fetchFeed(feedURL string) (*opdsFeed, error) {
	logrus.Tracef("%sfetching feed: %s", loggerOPDS, feedURL)

	resp, err := o.doRequest(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	var feed opdsFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	return &feed, nil
}

// resolveURL resolves a relative URL against the feed base URL
func (o *OPDSIntegration) resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}
	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return baseURL.ResolveReference(refURL).String()
}

// List returns the contents of an OPDS catalog folder
func (o *OPDSIntegration) List(folderID string, depth int) (*messages.IntegrationFolder, error) {
	var feedURL string
	var folderName string

	if folderID == rootFolder {
		feedURL = o.feedURL
		folderName = "OPDS Catalog"
	} else {
		decoded, err := decodeName(folderID)
		if err != nil {
			return nil, err
		}
		feedURL = decoded
		folderName = path.Base(feedURL)
	}

	logrus.Infof("%squery for: %s depth: %d", loggerOPDS, feedURL, depth)

	feed, err := o.fetchFeed(feedURL)
	if err != nil {
		return nil, err
	}

	if feed.Title != "" {
		folderName = feed.Title
	}

	response := messages.NewIntegrationFolder(folderID, folderName)

	for _, entry := range feed.Entries {
		// Check for navigation links (subfolders)
		navLink := o.findNavigationLink(entry.Links)
		if navLink != nil {
			resolvedURL := o.resolveURL(feedURL, navLink.Href)
			subfolder := messages.NewIntegrationFolder(encodeName(resolvedURL), entry.Title)
			response.SubFolders = append(response.SubFolders, subfolder)
			logrus.Tracef("%ssubfolder added: %s", loggerOPDS, entry.Title)
			continue
		}

		// Check for acquisition links (downloadable files)
		acqLink := o.findAcquisitionLink(entry.Links)
		if acqLink != nil {
			file := o.entryToFile(entry, acqLink, feedURL)
			if file != nil {
				response.Files = append(response.Files, file)
				logrus.Tracef("%sfile added: %s", loggerOPDS, entry.Title)
			}
		}
	}

	return response, nil
}

// findNavigationLink finds a navigation link in an entry
func (o *OPDSIntegration) findNavigationLink(links []opdsLink) *opdsLink {
	for i, link := range links {
		if link.Rel == relSubsection {
			return &links[i]
		}
		// Also treat atom feeds as navigation
		if strings.Contains(link.Type, "atom+xml") &&
			!strings.Contains(link.Rel, "acquisition") {
			return &links[i]
		}
	}
	return nil
}

// findAcquisitionLink finds a downloadable file link
func (o *OPDSIntegration) findAcquisitionLink(links []opdsLink) *opdsLink {
	// Priority: epub > pdf
	var pdfLink, epubLink *opdsLink

	for i, link := range links {
		if !strings.Contains(link.Rel, "acquisition") {
			continue
		}
		switch link.Type {
		case "application/epub+zip":
			epubLink = &links[i]
		case "application/pdf":
			pdfLink = &links[i]
		}
	}

	if epubLink != nil {
		return epubLink
	}
	return pdfLink
}

// entryToFile converts an OPDS entry to an IntegrationFile
func (o *OPDSIntegration) entryToFile(entry opdsEntry, link *opdsLink, baseURL string) *messages.IntegrationFile {
	contentType := link.Type
	if contentType != "application/pdf" && contentType != "application/epub+zip" {
		return nil
	}

	var ext, fileType string
	if contentType == "application/pdf" {
		ext = "pdf"
		fileType = "pdf"
	} else {
		ext = "epub"
		fileType = "epub"
	}

	resolvedURL := o.resolveURL(baseURL, link.Href)

	// Include thumbnail URL in the encoded ID if available
	thumbnailURL := o.findThumbnailLink(entry.Links, baseURL)
	encodedID := encodeFileWithThumbnail(resolvedURL, thumbnailURL)

	updated := time.Now()
	if entry.Updated != "" {
		if t, err := time.Parse(time.RFC3339, entry.Updated); err == nil {
			updated = t
		}
	}

	return &messages.IntegrationFile{
		ID:               encodedID,
		FileID:           encodedID,
		Name:             entry.Title,
		FileExtension:    ext,
		FileType:         fileType,
		ProvidedFileType: contentType,
		SourceFileType:   contentType,
		DateChanged:      updated,
		Size:             0, // Size unknown from OPDS metadata
	}
}

// findThumbnailLink finds an image/thumbnail link in an entry
func (o *OPDSIntegration) findThumbnailLink(links []opdsLink, baseURL string) string {
	// Prefer thumbnail over full image
	for _, link := range links {
		if link.Rel == relImageThumbnail {
			return o.resolveURL(baseURL, link.Href)
		}
	}
	for _, link := range links {
		if link.Rel == relImage {
			return o.resolveURL(baseURL, link.Href)
		}
	}
	// Some feeds use "image" directly
	for _, link := range links {
		if strings.Contains(link.Type, "image/") && link.Href != "" {
			return o.resolveURL(baseURL, link.Href)
		}
	}
	return ""
}

// encodeFileWithThumbnail encodes file URL and optional thumbnail URL together
func encodeFileWithThumbnail(fileURL, thumbnailURL string) string {
	if thumbnailURL == "" {
		return encodeName(fileURL)
	}
	return encodeName(fileURL + "\n" + thumbnailURL)
}

// decodeFileWithThumbnail decodes a file ID into file URL and thumbnail URL
func decodeFileWithThumbnail(encoded string) (fileURL, thumbnailURL string, err error) {
	decoded, err := decodeName(encoded)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(decoded, "\n", 2)
	fileURL = parts[0]
	if len(parts) > 1 {
		thumbnailURL = parts[1]
	}
	return fileURL, thumbnailURL, nil
}

// GetMetadata returns metadata for a specific file
func (o *OPDSIntegration) GetMetadata(fileID string) (*messages.IntegrationMetadata, error) {
	fileURL, thumbnailURL, err := decodeFileWithThumbnail(fileID)
	if err != nil {
		return nil, err
	}

	// Determine content type from URL
	contentType := "application/pdf"
	lowerURL := strings.ToLower(fileURL)
	if strings.HasSuffix(lowerURL, ".epub") || strings.Contains(lowerURL, "epub") {
		contentType = "application/epub+zip"
	}

	ext := path.Ext(fileURL)
	name := path.Base(fileURL)
	if ext != "" {
		name = strings.TrimSuffix(name, ext)
	}

	// Fetch thumbnail if available
	var thumbnail []byte
	if thumbnailURL != "" {
		thumbnail = o.fetchThumbnail(thumbnailURL)
	}

	return &messages.IntegrationMetadata{
		ID:               fileID,
		Name:             name,
		Thumbnail:        thumbnail,
		SourceFileType:   contentType,
		ProvidedFileType: contentType,
		FileType:         strings.TrimPrefix(ext, "."),
	}, nil
}

// fetchThumbnail downloads and returns thumbnail image data
func (o *OPDSIntegration) fetchThumbnail(thumbnailURL string) []byte {
	logrus.Tracef("%sfetching thumbnail: %s", loggerOPDS, thumbnailURL)

	resp, err := o.doRequest(thumbnailURL)
	if err != nil {
		logrus.Warnf("%sfailed to fetch thumbnail: %v", loggerOPDS, err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("%sthumbnail returned status %d", loggerOPDS, resp.StatusCode)
		return nil
	}

	// Limit thumbnail size to 1MB
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		logrus.Warnf("%sfailed to read thumbnail: %v", loggerOPDS, err)
		return nil
	}

	return data
}

// Download fetches a file from the OPDS catalog
func (o *OPDSIntegration) Download(fileID string) (io.ReadCloser, int64, error) {
	fileURL, _, err := decodeFileWithThumbnail(fileID)
	if err != nil {
		return nil, 0, err
	}

	logrus.Tracef("%sdownloading: %s", loggerOPDS, fileURL)

	resp, err := o.doRequest(fileURL)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, 0, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return resp.Body, resp.ContentLength, nil
}

// Upload is not supported for OPDS (read-only)
func (o *OPDSIntegration) Upload(folderID, name, fileType string, reader io.ReadCloser) (string, error) {
	if reader != nil {
		reader.Close()
	}
	return "", errors.New("OPDS integration is read-only, upload not supported")
}
