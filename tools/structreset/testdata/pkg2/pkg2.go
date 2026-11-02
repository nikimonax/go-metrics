package pkg2

type Resettable interface {
	Reset()
}

type (
	NotResettableEmbed struct {
		EmbedStr string
	}

	// generate:reset
	ResettableEmbed struct {
		EmbedChild Resettable
	}
)
