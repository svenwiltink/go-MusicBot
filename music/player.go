package music

// Player is the wrapper around MusicProviders. This should keep track of the queue and control
// the MusicProviders
type Player interface {
	Start()
	AddSong(song *Song) error
	Next() error
	Stop()
}
