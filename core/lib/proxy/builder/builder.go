package builder

import "errors"

type Builder struct {
	coreJSON []byte
	// There will be TUN config in future I guess
}

var ErrCoreIsNill = errors.New("core config is nil")

func NewBuilder(core []byte) Builder {
	// I want to copy there to dont change source bytes
	// I think it's not necessarry, but better be safe
	coreCopy := make([]byte, len(core))
	copy(coreCopy, core)

	return Builder{coreCopy}
}

func (b *Builder) Build() []byte {
	// Copy there is same as above
	coreCopy := make([]byte, len(b.coreJSON))
	copy(coreCopy, b.coreJSON)

	return coreCopy
}
