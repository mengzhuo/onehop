package onehop

const BUFFER_LENGTH = 1024

func NewLeakyBuffer(length, size int) *LeakyBuffer {
	return &LeakyBuffer{make(chan []byte, length), size}
}

type LeakyBuffer struct {
	buf  chan []byte
	size int
}

func (l *LeakyBuffer) Get() []byte {
	select {
	case p := <-l.buf:
		return p
	default:
		p := make([]byte, l.size)
		return p
	}
}

func (l *LeakyBuffer) Put(p []byte) {

	select {
	case l.buf <- p:
	default:

	}

}
