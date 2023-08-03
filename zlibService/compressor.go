package zlibService

type Compressor struct{}

func NewZlibCompressor() *Compressor {
	return &Compressor{}
}
