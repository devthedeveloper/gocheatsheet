// channels: uploaded files, checksummed as they land — hand to hand.
package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	dir, _ := os.MkdirTemp("", "uploads")
	defer os.RemoveAll(dir)

	uploads := make(chan string) // the hand-off point: file paths

	go func() { // goroutine A: receives the uploads
		docs := map[string]string{
			"invoice_jan.pdf": "invoice January contents",
			"invoice_feb.pdf": "invoice February contents",
			"salary_slip.pdf": "salary slip contents",
		}
		for name, body := range docs {
			path := filepath.Join(dir, name)
			os.WriteFile(path, []byte(body), 0o644) // REAL disk write
			fmt.Println("upload complete:", name)
			uploads <- path // hand the real file to the hasher
		}
		close(uploads) // no more uploads today
	}()

	for path := range uploads { // main: hashes each real file
		f, _ := os.Open(path)
		h := sha256.New()
		io.Copy(h, f) // REAL bytes read off disk, REAL hash computed
		f.Close()
		fmt.Printf("  sha256 %.12x  %s\n",
			h.Sum(nil), filepath.Base(path))
	}
	fmt.Println("all uploads processed ✓")
}
