package pkg

type Podcast struct{}

type Episode struct{}

type PodcastService interface {
	GetPodcast(id string) (Podcast, error)
	ListPodcasts() ([]Podcast, error)
	AddPodcast(podcast Podcast) error
	ListEpisodes(id string) ([]Episode, error)
	GetEpisode(podcastID string, episodeID string) error
}
