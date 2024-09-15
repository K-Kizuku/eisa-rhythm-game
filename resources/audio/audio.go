package audio

import (
	_ "embed"
)

var (
	//go:embed Echoing_in_the_Nago_night.mp3
	NagoNight []byte

	//go:embed The_Beat_of_Eisa.mp3
	EisaBeat []byte

	//go:embed taiko1.mp3
	Taiko1 []byte

	//go:embed taiko2.mp3
	Taiko2 []byte

	//go:embed taiko3.mp3
	Taiko3 []byte

	//go:embed shamisen.mp3
	Shamisen []byte

	//go:embed shamisen2.mp3
	Shamisen2 []byte
)
