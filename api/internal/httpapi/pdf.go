package httpapi

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func montarPDF(linhas []string) []byte {
	var body strings.Builder
	body.WriteString("BT /F1 11 Tf 50 760 Td\n")
	for i, linha := range linhas {
		if i > 0 {
			body.WriteString("0 -14 Td\n")
		}
		body.WriteString("(")
		body.WriteString(pdfEsc(linha))
		body.WriteString(") Tj\n")
	}
	body.WriteString("ET\n")
	stream := body.String()

	objs := []string{
		"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n",
		"2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n",
		"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >> endobj\n",
		fmt.Sprintf("4 0 obj << /Length %d >> stream\n%s\nendstream endobj\n", len(stream), stream),
		"5 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n",
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offs := make([]int, len(objs)+1)
	for i, obj := range objs {
		offs[i+1] = buf.Len()
		buf.WriteString(obj)
	}
	startxref := buf.Len()
	buf.WriteString("xref\n0 ")
	buf.WriteString(strconv.Itoa(len(objs) + 1))
	buf.WriteString("\n0000000000 65535 f \n")
	for i := 1; i <= len(objs); i++ {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offs[i]))
	}
	buf.WriteString("trailer << /Size ")
	buf.WriteString(strconv.Itoa(len(objs) + 1))
	buf.WriteString(" /Root 1 0 R >>\nstartxref\n")
	buf.WriteString(strconv.Itoa(startxref))
	buf.WriteString("\n%%EOF\n")
	return buf.Bytes()
}

func pdfEsc(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	s = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return '?'
		}
		return r
	}, s)
	if len(s) > 90 {
		s = s[:90]
	}
	return s
}
