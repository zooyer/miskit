package ssh

import "io"

type reader struct {
	size   int
	onRead func(int)
	reader io.Reader
}

type writer struct {
	size    int
	onWrite func(int)
	writer  io.Writer
}

func newReader(r io.Reader, onRead func(size int)) *reader {
	return &reader{
		size:   0,
		onRead: onRead,
		reader: r,
	}
}

func newWriter(w io.Writer, onWrite func(size int)) *writer {
	return &writer{
		size:    0,
		onWrite: onWrite,
		writer:  w,
	}
}

func (r *reader) Read(buf []byte) (n int, err error) {
	if n, err = r.reader.Read(buf); err != nil {
		return
	}
	r.size += n
	if r.onRead != nil {
		r.onRead(r.size)
	}
	return
}

func (w *writer) Write(buf []byte) (n int, err error) {
	if n, err = w.writer.Write(buf); err != nil {
		return
	}
	w.size += n
	if w.onWrite != nil {
		w.onWrite(w.size)
	}
	return
}
