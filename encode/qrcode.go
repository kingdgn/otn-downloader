package encode

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"rsc.io/qr"
)

type Key string

const (
	KeyMeta Key = "m"
	KeyData Key = "d"
)

type Value struct {
	Key   Key
	Index string
	Value string
}

type MetaValue struct {
	Filename  string `json:"filename"`
	Total     int    `json:"total"`
	FileSize  int    `json:"file_size"`
	ChunkSize int    `json:"chunk_size"`
}

func (v Value) Text() string {
	return fmt.Sprintf("%s:%s:%s", v.Key, v.Index, v.Value)
}

func printQRCode(v Value) {
	fmt.Printf("\033[0;0H")
	qrterminal.Generate(v.Text(), qrterminal.L, os.Stdout)
}

func loadValues(filePath string, cfg Config) ([]Value, []Value, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	datas := make([]Value, 0)
	buf := make([]byte, cfg.ChunkSize)
	s := 0
	i := 0

	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, nil, err
		}

		s += n
		value := base64.StdEncoding.EncodeToString(buf[:n])
		data := Value{
			Key:   KeyData,
			Index: strconv.FormatInt(int64(i), 10),
			Value: value,
		}
		datas = append(datas, data)
		i++
	}

	filename := filepath.Base(filePath)
	meta := MetaValue{
		Filename:  filename,
		Total:     i,
		FileSize:  s,
		ChunkSize: cfg.ChunkSize,
	}
	metaData, _ := json.Marshal(meta)
	metas := []Value{
		{
			Key:   KeyMeta,
			Index: "json",
			Value: string(metaData),
		},
	}

	return metas, datas, nil
}

const metaSleep = 5 * time.Second

var sliceSpecPattern = regexp.MustCompile(`\d+(?:\s*[-~～]\s*\d+)?`)

func ParseSliceSpecs(strValues []string) (map[int]bool, error) {
	result := map[int]bool{}

	for _, raw := range strValues {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		matches := sliceSpecPattern.FindAllString(raw, -1)
		if len(matches) == 0 {
			return nil, fmt.Errorf("no slice indexes found in %q", raw)
		}

		for _, match := range matches {
			token := strings.Join(strings.Fields(match), "")
			sep := ""
			for _, candidate := range []string{"-", "~", "～"} {
				if strings.Contains(token, candidate) {
					sep = candidate
					break
				}
			}

			if sep == "" {
				val, err := strconv.Atoi(token)
				if err != nil {
					return nil, fmt.Errorf("invalid slice index %q: %w", token, err)
				}
				result[val] = true
				continue
			}

			parts := strings.SplitN(token, sep, 2)
			start, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid slice range %q: %w", token, err)
			}
			end, err := strconv.Atoi(parts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid slice range %q: %w", token, err)
			}
			if end < start {
				return nil, fmt.Errorf("invalid slice range %q: end is smaller than start", token)
			}
			for i := start; i <= end; i++ {
				result[i] = true
			}
		}
	}

	return result, nil
}

func formatSliceList(slices map[int]bool, limit int) string {
	if len(slices) == 0 {
		return ""
	}
	values := make([]int, 0, len(slices))
	for i := range slices {
		values = append(values, i)
	}
	sort.Ints(values)
	if limit > 0 && len(values) > limit {
		values = values[:limit]
	}

	parts := make([]string, 0, len(values)+1)
	for _, val := range values {
		parts = append(parts, strconv.Itoa(val))
	}
	if limit > 0 && len(slices) > limit {
		parts = append(parts, "...")
	}
	return strings.Join(parts, " ")
}

func validateSlices(slices map[int]bool, total int) error {
	if len(slices) == 0 {
		return nil
	}

	invalid := map[int]bool{}
	for index := range slices {
		if index < 0 || index >= total {
			invalid[index] = true
		}
	}
	if len(invalid) == 0 {
		return nil
	}

	validRange := "none"
	if total > 0 {
		validRange = fmt.Sprintf("0-%d", total-1)
	}
	return fmt.Errorf("slice index out of range: %s; total=%d, valid range=%s", formatSliceList(invalid, 20), total, validRange)
}

func encodeToQRCode(filename string, cfg Config) error {
	metas, datas, err := loadValues(filename, cfg)
	if err != nil {
		return err
	}
	if err := validateSlices(cfg.Slices, len(datas)); err != nil {
		return err
	}

	if !cfg.SkipMeta {
		for _, v := range metas {
			printQRCode(v)
			time.Sleep(metaSleep)
		}
	}

	d := 1 * time.Second / time.Duration(cfg.Fps)

	for i, v := range datas {
		if len(cfg.Slices) > 0 && !cfg.Slices[i] {
			continue
		}

		printQRCode(v)
		time.Sleep(d)
	}

	return nil
}

func encodeToQRCodeImages(filename string, cfg Config) error {
	metas, datas, err := loadValues(filename, cfg)
	if err != nil {
		return err
	}
	if cfg.ImageScale <= 0 {
		cfg.ImageScale = 8
	}
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return err
	}
	if err := validateSlices(cfg.Slices, len(datas)); err != nil {
		return err
	}

	manifestCount := 0
	if !cfg.SkipMeta {
		for i, v := range metas {
			name := "manifest.png"
			if i > 0 {
				name = fmt.Sprintf("manifest_%03d.png", i)
			}
			if err := writeQRCodePNG(filepath.Join(cfg.OutputDir, name), v, cfg.ImageScale); err != nil {
				return err
			}
			if v.Key == KeyMeta && v.Index == "json" {
				if err := os.WriteFile(filepath.Join(cfg.OutputDir, "manifest.json"), []byte(v.Value), 0644); err != nil {
					return err
				}
			}
			manifestCount++
		}
	}

	written := 0
	for i, v := range datas {
		if len(cfg.Slices) > 0 && !cfg.Slices[i] {
			continue
		}
		name := fmt.Sprintf("frame_%06d.png", i)
		if err := writeQRCodePNG(filepath.Join(cfg.OutputDir, name), v, cfg.ImageScale); err != nil {
			return err
		}
		written++
	}

	fmt.Printf("wrote %d data QR images and %d manifest QR image(s) to %s\n", written, manifestCount, cfg.OutputDir)
	return nil
}

func writeQRCodePNG(filename string, v Value, scale int) error {
	code, err := qr.Encode(v.Text(), qr.L)
	if err != nil {
		return err
	}
	code.Scale = scale
	return os.WriteFile(filename, code.PNG(), 0644)
}

type Config struct {
	Fps                int
	ChunkSize          int
	Loop               int
	Slices             map[int]bool
	OutputDir          string
	ImageScale         int
	SkipMeta           bool
	InteractiveMissing bool
}

func EncodToQRCode(filename string, cfg Config) {
	if cfg.OutputDir != "" {
		err := encodeToQRCodeImages(filename, cfg)
		if err != nil {
			log.Fatalf("encode file to QRCode images failed %+v", err)
		}
		return
	}

	if cfg.InteractiveMissing {
		err := encodeInteractiveMissing(filename, cfg)
		if err != nil {
			log.Fatalf("encode file to QRCode failed %+v", err)
		}
		return
	}

	err := encodeLoop(filename, cfg)
	if err != nil {
		log.Fatalf("encode file to QRCode failed %+v", err)
	}
}

func encodeLoop(filename string, cfg Config) error {
	for range cfg.Loop {
		err := encodeToQRCode(filename, cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func encodeInteractiveMissing(filename string, cfg Config) error {
	reader := bufio.NewReader(os.Stdin)

	for round := 1; ; round++ {
		if len(cfg.Slices) == 0 {
			fmt.Printf("\ninteractive round %d: streaming the full sequence for %d loop(s).\n", round, cfg.Loop)
		} else {
			fmt.Printf("\ninteractive round %d: streaming %d selected slice(s) for %d loop(s): %s\n", round, len(cfg.Slices), cfg.Loop, formatSliceList(cfg.Slices, 60))
		}

		if err := encodeLoop(filename, cfg); err != nil {
			return err
		}

		for {
			fmt.Print("\nPaste missing slice indexes for next round, or press Enter to exit: ")
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF {
				return err
			}

			line = strings.TrimSpace(line)
			if line == "" {
				fmt.Println("interactive missing mode finished")
				return nil
			}

			slices, parseErr := ParseSliceSpecs([]string{line})
			if parseErr != nil {
				fmt.Printf("invalid slice list: %v\n", parseErr)
				if err == io.EOF {
					return nil
				}
				continue
			}

			cfg.Slices = slices
			cfg.SkipMeta = true
			if err == io.EOF {
				return nil
			}
			break
		}
	}
}
