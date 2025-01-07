package objectstorage

type ObjectStorage interface {
	SaveFromURL(url, podcastGUID, episodeGUID string) error
	// Load(podcastGUID, episodeGUID string) (io.Reader, error)
}
