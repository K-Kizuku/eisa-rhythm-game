package main

import (
	"bytes"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"log"
	"math/rand"
	"time"

	raudio "github.com/K-Kizuku/eisa-rhythm-game/resources/audio"
	riaudio "github.com/K-Kizuku/eisa-rhythm-game/resources/image"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 1000
	screenHeight = 800

	sampleRate = 48000
)

var (
	playerBarColor     = color.RGBA{0x80, 0x80, 0x80, 0xff}
	playerCurrentColor = color.RGBA{0xff, 0xff, 0xff, 0xff}
	tapPoints          = []float64{
		3.88,
		7.76,
		11.64,
		15.52,
		19.4,
		23.28,
		27.17,
		31.05,
		34.93,
		38.81,
		42.69,
		46.57,
		50.45,
		54.33,
		58.21,
		62.09,
		65.97,
		69.85,
		73.74,
		77.62,
		81.5,
		85.38,
		89.26,
		93.14,
		97.02,
		100.9,
		104.78,
		108.66,
		112.54,
		116.42,
		120.3,
		124.19,
		128.07,
		131.95,
		135.83,
		139.71,
		143.59,
		147.47,
		151.35,
		155.23,
		159.11,
		162.99,
		166.87,
		170.76,
		174.64,
		178.52,
		182.4,
		186.28,
		190.16,
	}
	tapEisaPoints = []float64{
		0.0,
		1.06,
		2.12,
		3.18,
		4.24,
		5.3,
		6.36,
		7.42,
		8.48,
		9.54,
		10.6,
		11.66,
		12.72,
		13.78,
		14.84,
		15.9,
		16.96,
		18.02,
		19.08,
		20.14,
		21.2,
		22.26,
		23.32,
		24.38,
		25.44,
		26.5,
		27.56,
		28.62,
		29.68,
		30.74,
		31.8,
		32.86,
		33.92,
		34.98,
		36.04,
		37.1,
		38.16,
		39.22,
		40.28,
		41.34,
		42.4,
		43.46,
		44.52,
		45.58,
		46.64,
		47.7,
		48.76,
		49.82,
		50.88,
		51.94,
		53.0,
		54.06,
		55.12,
		56.18,
		57.24,
		58.3,
		59.36,
		60.42,
		61.48,
		62.54,
		63.6,
		64.66,
		65.72,
		66.78,
		67.84,
		68.9,
		69.96,
		71.02,
		72.08,
		73.14,
		74.2,
		75.26,
		76.32,
		77.38,
		78.44,
		79.5,
		80.56,
		81.62,
		82.68,
		83.74,
		84.8,
		85.86,
		86.92,
		87.98,
		89.04,
		90.1,
		91.16,
		92.22,
		93.28,
		94.34,
		95.4,
		96.46,
		97.52,
		98.58,
		99.64,
		100.7,
		101.76,
		102.82,
		103.88,
		104.94,
	}
)

var (
	yanbarukuinaImage *ebiten.Image
	chinsukouImage    *ebiten.Image
	alertButtonImage  *ebiten.Image
	backGround        *ebiten.Image
	currentTime       time.Duration
)

func init() {
	img, _, err := image.Decode(bytes.NewReader(riaudio.Yanbarukuina))
	if err != nil {
		panic(err)
	}
	yanbarukuinaImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(riaudio.Chinsukou))
	if err != nil {
		panic(err)
	}
	chinsukouImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(riaudio.Chinsukou))
	if err != nil {
		panic(err)
	}
	alertButtonImage = ebiten.NewImageFromImage(img)

	img, _, err = image.Decode(bytes.NewReader(riaudio.Background))
	if err != nil {
		panic(err)
	}
	backGround = ebiten.NewImageFromImage(img)
}

type musicType int

const (
	typeOgg musicType = iota
	typeMP3
)

func (t musicType) String() string {
	switch t {
	case typeOgg:
		return "Ogg"
	case typeMP3:
		return "MP3"
	default:
		panic("not reached")
	}
}

// Player represents the current audio state.
type Player struct {
	game         *Game
	audioContext *audio.Context
	audioPlayer  *audio.Player
	current      time.Duration
	total        time.Duration
	seBytes      []byte
	seCh         chan []byte
	volume128    int
	musicType    musicType

	playButtonPosition  image.Point
	alertButtonPosition image.Point
}

func playerBarRect() (x, y, w, h int) {
	w, h = 600, 8
	x = (screenWidth - w) / 2
	y = screenHeight - h - 16
	return
}

func NewPlayer(game *Game, audioContext *audio.Context, musicType musicType) (*Player, error) {
	type audioStream interface {
		io.ReadSeeker
		Length() int64
	}

	const bytesPerSample = 4 // TODO: This should be defined in audio package

	var s audioStream

	switch musicType {
	case typeOgg:
		var err error
		s, err = mp3.DecodeWithoutResampling(bytes.NewReader(raudio.NagoNight))
		if err != nil {
			return nil, err
		}
	case typeMP3:
		var err error
		s, err = mp3.DecodeWithoutResampling(bytes.NewReader(raudio.EisaBeat))
		if err != nil {
			return nil, err
		}
	default:
		panic("not reached")
	}
	p, err := audioContext.NewPlayer(s)
	if err != nil {
		return nil, err
	}
	player := &Player{
		game:         game,
		audioContext: audioContext,
		audioPlayer:  p,
		total:        time.Second * time.Duration(s.Length()) / bytesPerSample / sampleRate,
		volume128:    128,
		seCh:         make(chan []byte),
		musicType:    musicType,
	}
	if player.total == 0 {
		player.total = 1
	}

	const buttonPadding = 16
	w := yanbarukuinaImage.Bounds().Dx()
	player.playButtonPosition.X = (screenWidth - w*2 + buttonPadding*1) / 2
	player.playButtonPosition.Y = screenHeight - 160

	player.alertButtonPosition.X = player.playButtonPosition.X + w + buttonPadding
	player.alertButtonPosition.Y = player.playButtonPosition.Y

	player.audioPlayer.Play()
	go func() {
		// s, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(raudio.Jab_wav))
		s, err := mp3.DecodeWithoutResampling(bytes.NewReader(raudio.Taiko3))
		if err != nil {
			log.Fatal(err)
			return
		}
		b, err := io.ReadAll(s)
		if err != nil {
			log.Fatal(err)
			return
		}
		player.seCh <- b
	}()
	return player, nil
}

func (p *Player) Close() error {
	return p.audioPlayer.Close()
}

func (p *Player) update() error {
	select {
	case p.seBytes = <-p.seCh:
		close(p.seCh)
		p.seCh = nil
	default:
	}

	if p.audioPlayer.IsPlaying() {
		p.current = p.audioPlayer.Position()
	}
	if err := p.seekBarIfNeeded(); err != nil {
		return err
	}
	p.switchPlayStateIfNeeded()
	p.playSEIfNeeded()
	p.updateVolumeIfNeeded()

	if inpututil.IsKeyJustPressed(ebiten.KeyU) {
		b := ebiten.IsRunnableOnUnfocused()
		ebiten.SetRunnableOnUnfocused(!b)
	}
	currentTime = p.current
	return nil
}

func (p *Player) shouldPlaySE() bool {
	if p.seBytes == nil {
		// Bytes for the SE is not loaded yet.
		return false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		return true
	}
	r := image.Rectangle{
		Min: p.alertButtonPosition,
		Max: p.alertButtonPosition.Add(alertButtonImage.Bounds().Size()),
	}
	if image.Pt(ebiten.CursorPosition()).In(r) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			return true
		}
	}
	for _, id := range p.game.justPressedTouchIDs {
		if image.Pt(ebiten.TouchPosition(id)).In(r) {
			return true
		}
	}
	return false
}

func (p *Player) playSEIfNeeded() {
	if !p.shouldPlaySE() {
		return
	}
	sePlayer := p.audioContext.NewPlayerFromBytes(p.seBytes)
	sePlayer.Play()
}

func (p *Player) updateVolumeIfNeeded() {
	if ebiten.IsKeyPressed(ebiten.KeyZ) {
		p.volume128--
	}
	if ebiten.IsKeyPressed(ebiten.KeyX) {
		p.volume128++
	}
	if p.volume128 < 0 {
		p.volume128 = 0
	}
	if 128 < p.volume128 {
		p.volume128 = 128
	}
	p.audioPlayer.SetVolume(float64(p.volume128) / 128)
}

func (p *Player) shouldSwitchPlayStateIfNeeded() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		return true
	}
	r := image.Rectangle{
		Min: p.playButtonPosition,
		Max: p.playButtonPosition.Add(yanbarukuinaImage.Bounds().Size()),
	}
	if image.Pt(ebiten.CursorPosition()).In(r) {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			return true
		}
	}
	for _, id := range p.game.justPressedTouchIDs {
		if image.Pt(ebiten.TouchPosition(id)).In(r) {
			return true
		}
	}
	return false
}

func (p *Player) switchPlayStateIfNeeded() {
	if !p.shouldSwitchPlayStateIfNeeded() {
		return
	}
	if p.audioPlayer.IsPlaying() {
		p.audioPlayer.Pause()
		return
	}
	p.audioPlayer.Play()
}

func (p *Player) justPressedPosition() (int, int, bool) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return x, y, true
	}

	if len(p.game.justPressedTouchIDs) > 0 {
		x, y := ebiten.TouchPosition(p.game.justPressedTouchIDs[0])
		return x, y, true
	}

	return 0, 0, false
}

func (p *Player) seekBarIfNeeded() error {
	// Calculate the next seeking position from the current cursor position.
	x, y, ok := p.justPressedPosition()
	if !ok {
		return nil
	}
	bx, by, bw, bh := playerBarRect()
	const padding = 4
	if y < by-padding || by+bh+padding <= y {
		return nil
	}
	if x < bx || bx+bw <= x {
		return nil
	}
	pos := time.Duration(x-bx) * p.total / time.Duration(bw)
	p.current = pos
	if err := p.audioPlayer.SetPosition(pos); err != nil {
		return err
	}
	return nil
}

func (p *Player) draw(screen *ebiten.Image) {
	// Draw the bar.
	x, y, w, h := playerBarRect()
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), playerBarColor, true)

	// Draw the cursor on the bar.
	// c := p.current
	cx := float32(x) + float32(w)*float32(p.current)/float32(p.total)
	cy := float32(y) + float32(h)/2
	vector.DrawFilledCircle(screen, cx, cy, 12, playerCurrentColor, true)

	// Compose the current time text.
	// m := (c / time.Minute) % 100
	// s := (c / time.Second) % 60
	// ms := (c / time.Millisecond) % 1000
	// currentTimeStr := fmt.Sprintf("%02d:%02d.%03d", m, s, ms)

	// Draw buttons
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(p.playButtonPosition.X), float64(p.playButtonPosition.Y))
	if p.audioPlayer.IsPlaying() {
		// screen.DrawImage(chinsukouImage, op)
	} else {
		// screen.DrawImage(yanbarukuinaImage, op)
	}
	op.GeoM.Reset()
	op.GeoM.Translate(float64(p.alertButtonPosition.X), float64(p.alertButtonPosition.Y))
	// screen.DrawImage(alertButtonImage, op)

	// Draw the debug message.
	// 	msg := fmt.Sprintf(`TPS: %0.2f
	// Press S to toggle Play/Pause
	// Press P to play SE
	// Press Z or X to change volume of the music
	// Press U to switch the runnable-on-unfocused state
	// Press A to switch Ogg and MP3 (Current: %s)
	// Current Time: %s
	// Current Volume: %d/128
	// Type: %s`, ebiten.ActualTPS(), p.musicType,
	// 		currentTimeStr, int(p.audioPlayer.Volume()*128), p.musicType)
	// ebitenutil.DebugPrint(screen, msg)
}

type Game struct {
	musicPlayer   *Player
	musicPlayerCh chan *Player
	errCh         chan error

	justPressedTouchIDs []ebiten.TouchID
}

func NewGame() (*Game, error) {
	audioContext := audio.NewContext(sampleRate)

	g := &Game{
		musicPlayerCh: make(chan *Player),
		errCh:         make(chan error),
	}

	m, err := NewPlayer(g, audioContext, typeOgg)
	if err != nil {
		return nil, err
	}

	g.musicPlayer = m
	return g, nil
}

func (g *Game) Update() error {
	select {
	case p := <-g.musicPlayerCh:
		g.musicPlayer = p
	case err := <-g.errCh:
		return err
	default:
	}

	g.justPressedTouchIDs = inpututil.AppendJustPressedTouchIDs(g.justPressedTouchIDs[:0])

	if g.musicPlayer != nil && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		var t musicType
		switch g.musicPlayer.musicType {
		case typeOgg:
			t = typeMP3
		case typeMP3:
			t = typeOgg
		default:
			panic("not reached")
		}

		if err := g.musicPlayer.Close(); err != nil {
			return err
		}
		g.musicPlayer = nil

		go func() {
			p, err := NewPlayer(g, audio.CurrentContext(), t)
			if err != nil {
				g.errCh <- err
				return
			}
			g.musicPlayerCh <- p
		}()
	}

	if g.musicPlayer != nil {
		if err := g.musicPlayer.update(); err != nil {
			return err
		}
	}
	return nil
}

// 0~4のランダムな整数を返す関数
func randam() int {
	return rand.Intn(5)
}

func numPos(n int) int {
	return (n*107/(51+n) + n) % 5
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(backGround, nil)
	op := &ebiten.DrawImageOptions{}
	for i, v := range tapEisaPoints {
		op.GeoM.Reset()
		op.GeoM.Translate(float64(375+numPos(i)*50), float64(-v*300+float64((currentTime/time.Millisecond))))
		screen.DrawImage(chinsukouImage, op)
		// if i%2 == 0 {
		// 	op.GeoM.Reset()
		// 	op.GeoM.Translate(v, 0)
		// 	screen.DrawImage(yanbarukuinaImage, op)
		// }
	}

	if g.musicPlayer != nil {
		g.musicPlayer.draw(screen)
	}
	for n := range 5 {
		op.GeoM.Reset()
		op.GeoM.Translate(float64(375+numPos(n)*50), 700)
		// screen.DrawImage(ebitenutil.DrawRect(), op)
		vector.DrawFilledCircle(screen, float32(400+n*50), 700, 25, color.RGBA{0xff, 0xff, 0xff, 0xff}, false)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("No Eisa, No Life!")
	g, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
