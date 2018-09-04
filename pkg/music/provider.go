package music

// Provider is the interface for an implementation that can actually play songs
type Provider interface {
	CanPlay(song *Song) bool
	SetVolume(percentage int) error
	GetVolume() (int, error)
	PlaySong(song *Song) error
	Play() error
	Pause() error
	Skip() error
	// wait for the current song to end
	Wait()

	// stop the player.
	Stop()
}
