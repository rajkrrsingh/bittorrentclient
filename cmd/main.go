package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"torrent-client/client"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "BitTorrent Client v%s (built: %s)\n", Version, BuildTime)
		fmt.Fprintf(os.Stderr, "Usage: %s <torrent-file> [output-path]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --help\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "       %s --version\n", os.Args[0])
		os.Exit(1)
	}

	if os.Args[1] == "--version" || os.Args[1] == "-v" {
		fmt.Printf("BitTorrent Client v%s (built: %s)\n", Version, BuildTime)
		return
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Printf("BitTorrent Client v%s (built: %s)\n\n", Version, BuildTime)
		fmt.Printf("USAGE:\n")
		fmt.Printf("    %s <torrent-file> [output-path]\n\n", os.Args[0])
		fmt.Printf("ARGUMENTS:\n")
		fmt.Printf("    <torrent-file>    Path to the .torrent file to download\n")
		fmt.Printf("    [output-path]     Optional output path (defaults to torrent name)\n\n")
		fmt.Printf("FLAGS:\n")
		fmt.Printf("    -h, --help        Show this help message\n")
		fmt.Printf("    -v, --version     Show version information\n\n")
		fmt.Printf("EXAMPLES:\n")
		fmt.Printf("    %s example.torrent\n", os.Args[0])
		fmt.Printf("    %s example.torrent ./downloads/\n", os.Args[0])
		fmt.Printf("    %s example.torrent /path/to/output/file.txt\n", os.Args[0])
		return
	}

	torrentPath := os.Args[1]
	var outputPath string

	if len(os.Args) >= 3 {
		outputPath = os.Args[2]
	}

	torrent, err := client.Open(torrentPath)
	if err != nil {
		log.Fatalf("Failed to open torrent: %v", err)
	}

	if outputPath == "" {
		// Use the name from the torrent file
		outputPath = torrent.Name
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if outputDir != "" && outputDir != "." {
		err = os.MkdirAll(outputDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}
	}

	log.Printf("Starting download of '%s' (%d bytes)", torrent.Name, torrent.Length)
	log.Printf("Found %d peers", len(torrent.Peers))
	log.Printf("File will be saved as '%s'", outputPath)

	err = torrent.DownloadToFile(outputPath)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	log.Printf("Download completed successfully! File saved as '%s'", outputPath)
}
