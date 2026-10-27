package riff

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeckSaveWithWriter(t *testing.T) {
	for _, fail := range []string{"", ".cards", ".deck"} {
		t.Run(fail, func(t *testing.T) {
			dir := t.TempDir()
			deck, err := LoadDeck(dir, "20261027000000-abcdefg", 0.9, 36500, "")
			if err != nil {
				t.Fatal(err)
			}
			deck.AddCard("card", "block")
			want := errors.New("injected source write failure")
			var writes []string
			err = deck.SaveWithWriter(func(path string, data []byte) error {
				writes = append(writes, filepath.Ext(path))
				if fail != "" && strings.HasSuffix(path, fail) {
					return want
				}
				return os.WriteFile(path, data, 0600)
			})
			if fail != "" {
				if !errors.Is(err, want) || (fail == ".cards" && len(writes) != 1) {
					t.Fatalf("write failure was swallowed or followed by another write: %v, %v", writes, err)
				}
				return
			}
			if err != nil || len(writes) != 2 {
				t.Fatalf("save did not write both source files: %v, %v", writes, err)
			}
			loaded, err := LoadDeck(dir, deck.ID, 0.9, 36500, "")
			if err != nil || loaded.GetCard("card") == nil {
				t.Fatalf("existing reader cannot load writer output: %v", err)
			}
		})
	}
}
