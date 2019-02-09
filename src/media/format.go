package media

type MediaType int

const (
	Audio MediaType = iota
	Video
)


type Media struct {
	Url string
	Format string
	Type MediaType
}

func GetTypeFromFormat(format string) MediaType {
	if format == "mp3" {
		return Audio
	}
	return Video
}