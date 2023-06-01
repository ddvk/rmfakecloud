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

const googleMimeTypeFolder = "application/vnd.google-apps.folder"

var (
	gdrive_config       *oauth2.Config
	gdrive_oauth_states OAuthStates = OAuthStates{}
)

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

type OAuthStates map[string]*model.IntegrationConfig

func (o OAuthStates) create_state(c *gin.Context, int *model.IntegrationConfig) {
	uid := c.GetString("UserID")

	int.ID = uuid.NewString()
	o[uid] = int
}

func (o OAuthStates) get_state(c *gin.Context) (*model.IntegrationConfig, error) {
	uid := c.GetString("UserID")

	if state, ok := o[uid]; !ok {
		return nil, fmt.Errorf("No state registered for this user")
	} else {
		return state, nil
	}
}

func (o OAuthStates) compare_state(c *gin.Context, cmp_state string) bool {
	state, err := o.get_state(c)
	return err == nil && state.ID == cmp_state
}

func (o OAuthStates) drop_state(c *gin.Context) {
	uid := c.GetString("UserID")
	delete(o, uid)
}

func GDriveOAuthRedirect(c *gin.Context, int *model.IntegrationConfig) {
	if gdrive_config == nil {
		c.AbortWithStatusJSON(http.StatusNotAcceptable, gin.H{"error": "Google Drive integration is not available, please read the documentation in order to enable it."})
		return
	}

	gdrive_oauth_states.create_state(c, int)

	c.JSON(http.StatusOK, gin.H{"redirect": gdrive_config.AuthCodeURL(int.ID, oauth2.AccessTypeOffline)})
}

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

type GDrive struct {
	client *drive.Service
}

func newGDrive(i model.IntegrationConfig) *GDrive {
	t := &oauth2.Token{}
	err := json.Unmarshal([]byte(i.Accesstoken), t)
	if err != nil {
		logrus.Error("Unable to unmarshal gdrive token:", err.Error())
		return nil
	}

	client, err := drive.NewService(context.Background(), option.WithHTTPClient(gdrive_config.Client(context.Background(), t)))
	if err != nil {
		logrus.Errorf("An error occurred creating Drive client: %v\n", err)
		return nil
	}

	return &GDrive{
		client,
	}
}

type GFile struct {
	*drive.File
}

func (f GFile) Name() string {
	return f.File.Name
}

func (f GFile) Id() string {
	return f.File.Id
}

func (f GFile) ContentType() string {
	return f.File.MimeType
}

func (f GFile) Ext() string {
	return f.File.FullFileExtension
}

func (f GFile) Size() int64 {
	return f.File.Size
}

func (f GFile) Mode() fs.FileMode {
	return 0
}

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

func (f GFile) IsDir() bool {
	return f.MimeType == googleMimeTypeFolder
}

func (f GFile) Sys() interface{} {
	return nil
}

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
