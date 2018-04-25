package music

// MusicProvider is the interface for an implementation that can actually play songs
type MusicProvider interface {
	CanPlay(song *Song) bool
	PlaySong(song *Song) error
	Play() error
	Pause() error
	// wait for the current song to end
	Wait()
}
