package music

type DataProvider interface {
	CanProvideData(song *Song) bool

	// fill the song object with data
	ProvideData(song *Song) error
}
