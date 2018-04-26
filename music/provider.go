package music

// Provider is the interface for an implementation that can actually play songs
type Provider interface {
	CanPlay(song *Song) bool
	PlaySong(song *Song) error
	Play() error
	Pause() error
	// wait for the current song to end
	Wait()
}
