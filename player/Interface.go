package player

type MusicPlayerInterface interface {
	Name() (name string)
	CanPlay(url string) (canPlay bool)
	GetItems(url string) (items []ListItem, err error)
	SearchItems(searchStr string, limit int) (items []ListItem, err error)
	Play(url string) (err error)
	Pause(pauseState bool) (err error)
	Stop() (err error)
}
