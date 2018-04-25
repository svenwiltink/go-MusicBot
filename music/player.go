package music

// MusicPlayer is the wrapper around MusicProviders. This should keep track of the queue and control
// the MusicProviders
type MusicPlayer interface {
	Start()
	AddSong(song *Song) error
}
