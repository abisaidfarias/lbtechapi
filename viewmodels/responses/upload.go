package responses

// UploadFileResponse is returned by POST /api/v1/upload/images.
type UploadFileResponse struct {
	URL              string `json:"url"`
	OriginalFileName string `json:"originalFileName"`
	StoredFileName   string `json:"storedFileName,omitempty"`
}
