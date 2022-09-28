package music

type DataProvider interface {
	CanProvideData(song Song) bool
	Search(string) ([]Song, error)
	// fill the song object with data
	ProvideData(song *Song) error
	AddPlaylist(string) (*Playlist, error)
}
