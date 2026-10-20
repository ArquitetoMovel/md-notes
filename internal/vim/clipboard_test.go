package vim

import (
	"errors"
	"testing"
)

type fakeClipboardProvider struct {
	text     string
	readErr  error
	writeErr error
	writes   int
}

func (f *fakeClipboardProvider) ReadAll() (string, error) { return f.text, f.readErr }
func (f *fakeClipboardProvider) WriteAll(text string) error {
	f.writes++
	if f.writeErr == nil {
		f.text = text
	}
	return f.writeErr
}

func TestClipboard_StrictReadWriteErrors(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("clipboard unavailable")
	cases := []struct {
		name     string
		provider fakeClipboardProvider
		read     bool
	}{
		{"write failure", fakeClipboardProvider{writeErr: wantErr}, false},
		{"read failure", fakeClipboardProvider{readErr: wantErr}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			provider := tc.provider
			cb := NewClipboardWithProvider(&provider)
			cb.Set("legacy", false)
			if tc.read {
				_, err := cb.ReadStrict()
				if !errors.Is(err, wantErr) {
					t.Fatalf("ReadStrict error = %v, want %v", err, wantErr)
				}
			} else {
				err := cb.WriteStrict("new", false)
				if !errors.Is(err, wantErr) {
					t.Fatalf("WriteStrict error = %v, want %v", err, wantErr)
				}
			}
			if reg := cb.GetRegister(); reg.Content != "legacy" {
				t.Fatalf("strict failure changed legacy register: %+v", reg)
			}
		})
	}
}

// TC11: Sincronização com o Clipboard do Sistema Operacional e Registrador Local
func TestClipboard_SetAndGet(t *testing.T) {
	cb := NewClipboard()
	if cb == nil {
		t.Fatal("NewClipboard() retornou nil")
	}

	testText := "Texto para cópia e colagem\n"
	cb.Set(testText, true)

	content, isLineWise := cb.Get()
	if content != testText {
		t.Errorf("Esperado '%s', obtido '%s'", testText, content)
	}
	if !isLineWise {
		t.Error("Esperado isLineWise == true")
	}

	reg := cb.GetRegister()
	if reg.Content != testText || !reg.IsLineWise {
		t.Errorf("Registrador local inconsistente: %+v", reg)
	}
}
