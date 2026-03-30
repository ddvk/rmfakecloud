package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/ddvk/rmfakecloud/internal/config"
	"github.com/ddvk/rmfakecloud/internal/messages"
	"github.com/ddvk/rmfakecloud/internal/model"
	"github.com/ddvk/rmfakecloud/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// googleMimeTypeFolder is the MIME type used by Google Drive to identify folder objects.
const googleMimeTypeFolder = "application/vnd.google-apps.folder"

var (
	// gdrive_config holds the OAuth2 configuration for Google Drive integration.
	gdrive_config *oauth2.Config

	// gdrive_oauth_states stores temporary OAuth state information during the authentication flow.
	gdrive_oauth_states OAuthStates = OAuthStates{}
)

// ConfigGDrive initializes the Google Drive OAuth2 configuration.
func ConfigGDrive(cfg *config.Config) {
	gdrive_clientID := os.Getenv("GDRIVE_CLIENT_ID")
	gdrive_clientSecret := os.Getenv("GDRIVE_CLIENT_SECRET")

	if gdrive_clientID == "" || gdrive_clientSecret == "" {
		logrus.Warn("GDRIVE_CLIENT_ID or GDRIVE_CLIENT_SECRET not filled, Google Drive integration will not be available.")
	}

	redirect_url := cfg.StorageURL
	if redirect_url[len(redirect_url)-1] != '/' {
		redirect_url += "/"
	}
	redirect_url += "ui/api/integrations/google/complete"

	gdrive_config = &oauth2.Config{
		Scopes:      []string{"https://www.googleapis.com/auth/drive"},
		RedirectURL: redirect_url,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
		ClientID:     gdrive_clientID,
		ClientSecret: gdrive_clientSecret,
	}
}

// OAuthStates maintains a mapping of user IDs to their pending OAuth integration configurations.
type OAuthStates map[string]*model.IntegrationConfig

// create_state generates a new OAuth state for the user and stores the integration configuration.
func (o OAuthStates) create_state(c *gin.Context, int *model.IntegrationConfig) {
	uid := c.GetString("UserID")

	int.ID = uuid.NewString()
	o[uid] = int
}

// get_state retrieves the stored OAuth state for the current user.
func (o OAuthStates) get_state(c *gin.Context) (*model.IntegrationConfig, error) {
	uid := c.GetString("UserID")

	if state, ok := o[uid]; !ok {
		return nil, fmt.Errorf("No state registered for this user")
	} else {
		return state, nil
	}
}

// compare_state validates that the provided state matches the stored state for the current user.
func (o OAuthStates) compare_state(c *gin.Context, cmp_state string) bool {
	state, err := o.get_state(c)
	return err == nil && state.ID == cmp_state
}

// drop_state removes the OAuth state for the current user, cleaning up after the OAuth flow completes.
func (o OAuthStates) drop_state(c *gin.Context) {
	uid := c.GetString("UserID")
	delete(o, uid)
}

// GDriveOAuthRedirect initiates the OAuth2 flow for Google Drive integration.
func GDriveOAuthRedirect(c *gin.Context, int *model.IntegrationConfig) {
	if gdrive_config == nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": "Google Drive integration is not available, please read the documentation in order to enable it."})
		return
	}

	gdrive_oauth_states.create_state(c, int)

	c.JSON(http.StatusOK, gin.H{"redirect": gdrive_config.AuthCodeURL(int.ID, oauth2.AccessTypeOffline)})
}

// GDriveOAuthComplete handles the OAuth2 callback from Google.
func GDriveOAuthComplete(userStorer storage.UserStorer, uid string, c *gin.Context) {
	if gdrive_config == nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": "Google Drive integration is not available, please read the documentation in order to enable it."})
		return
	}

	state := c.Request.URL.Query().Get("state")
	if !gdrive_oauth_states.compare_state(c, state) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Authentication request expired"})
		return
	}

	oauth2Token, err := gdrive_config.Exchange(c.Request.Context(), c.Request.URL.Query().Get("code"))
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errmsg": "Failed to exchange token: " + err.Error()})
		return
	}

	int, err := gdrive_oauth_states.get_state(c)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Register the integration
	user, err := userStorer.GetUser(uid)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	at, err := json.Marshal(oauth2Token)
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	int.Accesstoken = string(at)

	user.Integrations = append(user.Integrations, *int)

	err = userStorer.UpdateUser(user)
	if err != nil {
		logrus.Error("error updating user", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusFound, "/integrations")
}

// GDrive provides an interface to Google Drive API operations.
type GDrive struct {
	client *drive.Service
}

// PersistentTokenSource wraps an oauth2.TokenSource and persists refreshed tokens to storage.
type PersistentTokenSource struct {
	userID      string
	userStorer  storage.UserStorer
	integration *model.IntegrationConfig
	source      oauth2.TokenSource
}

// Token returns a valid OAuth2 token, refreshing if necessary and persisting the refreshed token.
func (p *PersistentTokenSource) Token() (*oauth2.Token, error) {
	token, err := p.source.Token()
	if err != nil {
		return nil, err
	}

	// Save the (potentially refreshed) token back to storage
	at, err := json.Marshal(token)
	if err != nil {
		logrus.Warn("Failed to marshal refreshed token:", err)
		return token, nil
	}

	// Update the integration config in storage
	user, err := p.userStorer.GetUser(p.userID)
	if err != nil {
		logrus.Warn("Failed to get user for token refresh:", err)
		return token, nil
	}

	// Find and update the integration
	for i := range user.Integrations {
		if user.Integrations[i].ID == p.integration.ID {
			oldToken := user.Integrations[i].Accesstoken
			user.Integrations[i].Accesstoken = string(at)

			err = p.userStorer.UpdateUser(user)
			if err != nil {
				logrus.Warn("Failed to persist refreshed token:", err)
			} else if oldToken != string(at) {
				logrus.Info("[gdrive] Token refreshed and persisted for user:", p.userID)
			}
			break
		}
	}

	return token, nil
}

// newGDrive creates a new GDrive client instance for the specified user and integration.
func newGDrive(userID string, userStorer storage.UserStorer, i model.IntegrationConfig) *GDrive {
	t := &oauth2.Token{}
	err := json.Unmarshal([]byte(i.Accesstoken), t)
	if err != nil {
		logrus.Error("Unable to unmarshal gdrive token:", err.Error())
		return nil
	}

	// Create a token source that persists refreshed tokens
	tokenSource := &PersistentTokenSource{
		userID:      userID,
		userStorer:  userStorer,
		integration: &i,
		source:      gdrive_config.TokenSource(context.Background(), t),
	}

	client, err := drive.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		logrus.Errorf("An error occurred creating Drive client: %v\n", err)
		return nil
	}

	return &GDrive{
		client,
	}
}

// GFile wraps a Google Drive File and implements the fs.FileInfo interface.
type GFile struct {
	*drive.File
}

// Name returns the file name.
func (f GFile) Name() string {
	return f.File.Name
}

// Id returns the unique Google Drive file ID.
func (f GFile) Id() string {
	return f.File.Id
}

// ContentType returns the MIME type of the file.
func (f GFile) ContentType() string {
	return f.File.MimeType
}

// Ext returns the full file extension.
func (f GFile) Ext() string {
	return f.File.FullFileExtension
}

// Size returns the file size in bytes.
// For Google Workspace files (Docs, Sheets, etc.), this may be 0.
func (f GFile) Size() int64 {
	return f.File.Size
}

// Mode returns the file mode. Always returns 0 for Google Drive files.
func (f GFile) Mode() fs.FileMode {
	return 0
}

// ModTime returns the last modification time of the file.
func (f GFile) ModTime() time.Time {
	modifiedTime := f.ModifiedTime
	if modifiedTime == "" {
		modifiedTime = f.CreatedTime
	}
	if modifiedTime == "" {
		return time.Time{}
	}

	modTime, _ := time.Parse(time.RFC3339, modifiedTime)
	return modTime
}

// IsDir returns true if the file is a Google Drive folder.
func (f GFile) IsDir() bool {
	return f.MimeType == googleMimeTypeFolder
}

// Sys returns the underlying data source. Always returns nil for Google Drive files.
func (f GFile) Sys() interface{} {
	return nil
}

// GDriveGetProvidedContentType maps Google Workspace MIME types to their export formats.
// Google Docs, Sheets, and Presentations are exported as PDF.
// Other file types are returned unchanged.
func GDriveGetProvidedContentType(contentType string) string {
	providedContentType := contentType

	switch providedContentType {
	case "application/vnd.google-apps.spreadsheet":
		providedContentType = "application/pdf"
	case "application/vnd.google-apps.presentation":
		providedContentType = "application/pdf"
	case "application/vnd.google-apps.document":
		providedContentType = "application/pdf"
	}

	return providedContentType
}

// toIntegrationFile converts a GFile to an IntegrationFile message format.
// This method prepares file metadata for transmission to reMarkable devices.
func (d GFile) toIntegrationFile() *messages.IntegrationFile {
	ext := d.Ext()
	contentType := d.ContentType()

	return &messages.IntegrationFile{
		ProvidedFileType: GDriveGetProvidedContentType(contentType),
		DateChanged:      d.ModTime(),
		FileExtension:    ext,
		FileType:         ext,
		ID:               d.Id(),
		FileID:           d.Id(),
		Name:             d.Name(),
		Size:             d.Size(),
		SourceFileType:   contentType,
	}
}

// visitDir recursively traverses a Google Drive folder up to the specified depth.
func (g *GDrive) visitDir(currentPath string, depth int, parentFolder *messages.IntegrationFolder) error {
	if depth < 1 {
		return nil
	}

	logrus.Trace(loggerfs, "visiting: ", currentPath)
	fs, err := g.readdir(currentPath)
	if err != nil {
		return err
	}

	for _, f := range fs {
		if d, ok := f.(GFile); ok {
			entryName := d.Name()
			if d.IsDir() {
				folder := messages.NewIntegrationFolder(d.Id(), entryName)

				err = g.visitDir(d.Id(), depth-1, folder)
				if err != nil {
					return err
				}

				parentFolder.SubFolders = append(parentFolder.SubFolders, folder)
				logrus.Trace(loggerfs, "dir added: ", d.Name())
			} else {
				file := d.toIntegrationFile()
				parentFolder.Files = append(parentFolder.Files, file)
				logrus.Trace(loggerfs, "file added: ", path.Join(currentPath, d.Name()))
			}
		}
	}

	return nil
}

// readdir lists all files and folders in the specified Google Drive directory.
func (g *GDrive) readdir(dirname string) ([]fs.FileInfo, error) {
	q := g.client.Files.List()
	query := fmt.Sprintf("'%s' in parents and trashed=false", dirname)
	q.Q(query)
	q.Fields("files(id, name, appProperties, mimeType, size, modifiedTime, createdTime, fullFileExtension, thumbnailLink)")

	r, err := q.Do()
	if err != nil {
		return nil, err
	}

	files := make([]fs.FileInfo, len(r.Files))

	for i := range files {
		files[i] = GFile{r.Files[i]}
	}

	return files, nil
}

// List retrieves the folder structure from Google Drive with the specified depth.
func (g *GDrive) List(folder string, depth int) (*messages.IntegrationFolder, error) {
	response := messages.NewIntegrationFolder(folder, "")

	if folder == rootFolder {
		folder = "root"
		response.Name = "Google Drive root"

		// Add shared spaces as directory
		drives, err := g.client.Files.List().Q("sharedWithMe").Do()
		if err != nil {
			return nil, err
		}
		for _, drive := range drives.Files {
			if drive.MimeType == googleMimeTypeFolder {
				folder := messages.NewIntegrationFolder(drive.Id, drive.Name)

				err = g.visitDir(drive.Id, depth-1, folder)
				if err != nil {
					return nil, err
				}

				response.SubFolders = append(response.SubFolders, folder)
			} else {
				d := GFile{drive}

				file := d.toIntegrationFile()

				response.Files = append(response.Files, file)
			}
		}
	} else {
		file, err := g.client.Files.Get(folder).Do()
		if err != nil {
			response.Name = path.Base(folder)
		} else {
			response.Name = file.Name
		}
	}
	logrus.Info("[gdrive] query for: ", folder, " depth: ", depth)

	err := g.visitDir(folder, depth, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetMetadata retrieves detailed metadata for a specific file, including thumbnail image.
func (g *GDrive) GetMetadata(fileID string) (*messages.IntegrationMetadata, error) {
	q := g.client.Files.Get(fileID)
	q.Fields("id, name, appProperties, mimeType, size, modifiedTime, createdTime, fullFileExtension, thumbnailLink")
	fileprop, err := q.Do()
	if err != nil {
		return nil, err
	}

	metadata := messages.IntegrationMetadata{
		ID:               fileID,
		Name:             fileprop.Name,
		Thumbnail:        []byte{},
		SourceFileType:   fileprop.MimeType,
		ProvidedFileType: GDriveGetProvidedContentType(fileprop.MimeType),
		FileType:         path.Ext(fileprop.Name),
	}

	fmt.Println(fileprop.ThumbnailLink)

	if fileprop.ThumbnailLink != "" {
		resp, err := http.Get(fileprop.ThumbnailLink)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve thumbnail: %w", err)
		}
		defer resp.Body.Close()

		metadata.Thumbnail, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("unable to read thumbnail: %w", err)
		}
	}

	return &metadata, nil
}

// Download retrieves a file from Google Drive.
func (g *GDrive) Download(fileID string) (io.ReadCloser, int64, error) {
	fileprop, err := g.client.Files.Get(fileID).Do()
	if err != nil {
		return nil, 0, err
	}

	var res *http.Response
	if fileprop.MimeType != "application/pdf" && fileprop.MimeType != "application/epub+zip" {
		res, err = g.client.Files.Export(fileID, "application/pdf").Download()
	} else {
		res, err = g.client.Files.Get(fileID).Download()
	}

	if err != nil {
		return nil, 0, err
	}
	return res.Body, res.ContentLength, nil
}

// Upload uploads a file to the specified Google Drive folder.
func (g *GDrive) Upload(folderID, name, fileType string, reader io.ReadCloser) (string, error) {
	dstFile := &drive.File{
		Name:    name,
		Parents: []string{folderID},
	}

	res, err := g.client.Files.Create(dstFile).Media(reader).Do()
	if err != nil {
		return "", err
	}

	return res.Id, nil
}
