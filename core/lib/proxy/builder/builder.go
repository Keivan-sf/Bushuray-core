package builder

import "errors"

type Builder struct {
	coreJSON []byte
}

var ErrCoreIsNill = errors.New("core config is nil")

func NewBuilder(core []byte) Builder {
	coreCopy := make([]byte, len(core))
	copy(coreCopy, core)

	return Builder{coreCopy}
}

func (b *Builder) Build() []byte {
	coreCopy := make([]byte, len(b.coreJSON))
	copy(coreCopy, b.coreJSON)

	return coreCopy
}
