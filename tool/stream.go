package tool

import "bytes"

type Stream struct {
	Length     int
	BytesRead  int
	BytesWrote int
	buffer     *bytes.Buffer
}

func NewStream(buf []byte) *Stream {
	return &Stream{
		buffer: bytes.NewBuffer(buf),
		Length: len(buf),
	}
}

// ReadBytes reads and returns the next byte from the buffer.
func (s *Stream) ReadByte() (b byte, err error) {
	b, err = s.buffer.ReadByte()
	s.BytesRead += 1
	return b, nil
}

// Read returns a slice containing the next n bytes from the buffer.
func (s *Stream) Read(n int) []byte {
	s.BytesRead += n
	return s.buffer.Next(n)
}

// Len returns the number of bytes of the unread portion of the buffer;
func (s *Stream) Len() int {
	return s.buffer.Len()
}

func (s *Stream) Bytes() []byte {
	return s.buffer.Bytes()
}

func (s *Stream) String() string {
	return s.buffer.String()
}

func (s *Stream) Write(buf []byte) (n int, err error) {
	n, err = s.buffer.Write(buf)
	s.BytesWrote += n
	return
}

func (s *Stream) WriteByte(c byte) (err error) {
	err = s.buffer.WriteByte(c)
	s.BytesWrote += 1
	return nil
}
