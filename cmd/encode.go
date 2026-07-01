package cmd

import (
	"log"

	"github.com/kingdgn/otn-downloader/encode"
	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "encode data to the output",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if *filename == "" {
			log.Fatalf("miss input file")
		}
		if *fps <= 0 {
			log.Fatalf("fps must be greater than 0")
		}
		if *chunkSize <= 0 {
			log.Fatalf("chunk-size must be greater than 0")
		}
		if *loop <= 0 {
			log.Fatalf("loop must be greater than 0")
		}

		sliceSpecs := append([]string{}, *slices...)
		if *missingSlices != "" {
			sliceSpecs = append(sliceSpecs, *missingSlices)
		}
		parsedSlices, err := encode.ParseSliceSpecs(sliceSpecs)
		if err != nil {
			log.Fatalf("invalid slice list: %v", err)
		}

		cfg := encode.Config{
			Fps:                *fps,
			ChunkSize:          *chunkSize,
			Loop:               *loop,
			Slices:             parsedSlices,
			OutputDir:          *outputDir,
			ImageScale:         *imageScale,
			SkipMeta:           *skipMeta,
			InteractiveMissing: *interactiveMissing,
		}
		encode.EncodToQRCode(*filename, cfg)
	},
}

var (
	fps                *int
	chunkSize          *int
	filename           *string
	loop               *int
	slices             *[]string
	missingSlices      *string
	outputDir          *string
	imageScale         *int
	skipMeta           *bool
	interactiveMissing *bool
)

func init() {
	rootCmd.AddCommand(encodeCmd)
	fps = encodeCmd.Flags().Int("fps", 30, "the data encode fps")
	loop = encodeCmd.Flags().Int("loop", 3, "the number of times process")
	chunkSize = encodeCmd.Flags().IntP("chunk-size", "c", 60, "the chunk byte size of the input file")
	filename = encodeCmd.Flags().StringP("input-file", "f", "", "the source files")
	slices = encodeCmd.Flags().StringSliceP("slices", "s", []string{}, "only emit these data slice indexes; accepts repeated flags, pasted lists, comma separated values, and ranges like 37-40")
	missingSlices = encodeCmd.Flags().String("missing-slices", "", "same as --slices; paste the missing slice list from the receiver, for example \"0 37 40 46\"")
	outputDir = encodeCmd.Flags().StringP("output-dir", "o", "", "write QR code PNG files to this directory instead of streaming to terminal")
	imageScale = encodeCmd.Flags().Int("image-scale", 8, "PNG pixel scale for each QR module when output-dir is set")
	skipMeta = encodeCmd.Flags().Bool("skip-meta", false, "do not emit the manifest QR code; use after the receiver has already scanned the manifest")
	interactiveMissing = encodeCmd.Flags().Bool("interactive-missing", false, "after each pass, prompt for the next missing slice list; empty input exits")
}
