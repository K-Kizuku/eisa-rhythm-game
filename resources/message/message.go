package message

type Message int

const (
	E Message = iota
	D
	C
	B
	A
)

func (m Message) String() string {
	switch m {
	case E:
		return "なんぎーだったねぇ…" // 難しかったねぇ…
	case D:
		return "まーさんど！" // 頑張ったね！
	case C:
		return "あともうちょい！ぐぶりーさびら！" // あともう少し！頑張って！
	case B:
		return "んじ、かーぎ！次も楽しみさぁ！" // ほんと、上手！次も楽しみだね！
	case A:
		return "めっさ、かーぎ！あんたは天才さぁ！" // とても上手！あなたは天才だね！
	}
	return ""
}
