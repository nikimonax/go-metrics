package pkg2

type Resetable interface {
	Reset()
}

type (
	NotResetableEmbed struct {
		EmbedStr string
	}

	// generate:reset
	ResetableEmbed struct {
		EmbedChild Resetable
	}
)
