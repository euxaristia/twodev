package git

import (
	"bytes"
	"fmt"
	"io"
)

// WritePktLine writes a Git pkt-line record.
func WritePktLine(w io.Writer, payload []byte) error {
	if len(payload) > 0xffff {
		return fmt.Errorf("pkt-line payload too large")
	}
	if _, err := fmt.Fprintf(w, "%04x", len(payload)+4); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return nil
}

// WritePktFlush writes a pkt-line flush marker.
func WritePktFlush(w io.Writer) error {
	_, err := w.Write([]byte("0000"))
	return err
}

// AdvertiseService prepends the service announcement pkt-lines before git output.
func AdvertiseService(service string, gitOutput []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := WritePktLine(&buf, []byte("# service="+service+"\n")); err != nil {
		return nil, err
	}
	if err := WritePktFlush(&buf); err != nil {
		return nil, err
	}
	if _, err := buf.Write(gitOutput); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}