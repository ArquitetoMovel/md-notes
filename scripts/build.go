package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Target struct {
	OS   string
	Arch string
}

var allTargets = []Target{
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
	{OS: "windows", Arch: "arm64"},
}

func getGitInfo() (string, string) {
	cmdVersion := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	outVersion, err := cmdVersion.Output()
	version := "dev"
	if err == nil {
		version = strings.TrimSpace(string(outVersion))
	}

	cmdCommit := exec.Command("git", "rev-parse", "--short", "HEAD")
	outCommit, err := cmdCommit.Output()
	commit := "none"
	if err == nil {
		commit = strings.TrimSpace(string(outCommit))
	}

	return version, commit
}

func main() {
	targetOS := flag.String("os", "all", "Target OS: darwin, linux, windows, all")
	targetArch := flag.String("arch", "all", "Target Arch: amd64, arm64, all")
	versionFlag := flag.String("version", "", "Version string (default from git)")
	doPackage := flag.Bool("package", false, "Package release archives (.tar.gz, .zip) with checksums")
	doClean := flag.Bool("clean", false, "Clean bin and dist directories")
	flag.Parse()

	defaultVer, gitCommit := getGitInfo()
	version := defaultVer
	if *versionFlag != "" {
		version = *versionFlag
	}
	buildDate := time.Now().UTC().Format(time.RFC3339)

	if *doClean {
		fmt.Println("==> Cleaning bin/ and dist/...")
		_ = os.RemoveAll("bin")
		_ = os.RemoveAll("dist")
	}

	var targets []Target
	for _, t := range allTargets {
		if *targetOS != "all" && *targetOS != t.OS {
			continue
		}
		if *targetArch != "all" && *targetArch != t.Arch {
			continue
		}
		targets = append(targets, t)
	}

	if len(targets) == 0 {
		fmt.Printf("Error: No targets matched os=%s arch=%s\n", *targetOS, *targetArch)
		os.Exit(1)
	}

	fmt.Println("======================================================")
	fmt.Println(" md-notes Cross-Platform Builder")
	fmt.Printf(" Version:    %s\n", version)
	fmt.Printf(" Git Commit: %s\n", gitCommit)
	fmt.Printf(" Build Date: %s\n", buildDate)
	fmt.Printf(" Targets:    %d platform(s)\n", len(targets))
	fmt.Println("======================================================")

	ldflags := fmt.Sprintf("-s -w -X main.Version=%s -X main.GitCommit=%s -X main.BuildDate=%s",
		version, gitCommit, buildDate)

	type BuiltBinary struct {
		Target Target
		Path   string
		Name   string
	}
	var builtBinaries []BuiltBinary

	for _, t := range targets {
		outDir := filepath.Join("bin", fmt.Sprintf("%s_%s", t.OS, t.Arch))
		if err := os.MkdirAll(outDir, 0755); err != nil {
			fmt.Printf("Error creating dir %s: %v\n", outDir, err)
			os.Exit(1)
		}

		binName := "mdn"
		if t.OS == "windows" {
			binName = "mdn.exe"
		}
		outFile := filepath.Join(outDir, binName)

		fmt.Printf("==> Compiling %s/%s -> %s\n", t.OS, t.Arch, outFile)

		cmd := exec.Command("go", "build", "-trimpath", "-ldflags", ldflags, "-o", outFile, "./cmd/mdn")
		cmd.Env = append(os.Environ(),
			"CGO_ENABLED=0",
			fmt.Sprintf("GOOS=%s", t.OS),
			fmt.Sprintf("GOARCH=%s", t.Arch),
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error compiling for %s/%s: %v\n", t.OS, t.Arch, err)
			os.Exit(1)
		}

		builtBinaries = append(builtBinaries, BuiltBinary{
			Target: t,
			Path:   outFile,
			Name:   binName,
		})
	}

	// Universal macOS binary (when running on macOS and both darwin targets were built)
	if runtime.GOOS == "darwin" {
		var darwinAmd64, darwinArm64 string
		for _, b := range builtBinaries {
			if b.Target.OS == "darwin" && b.Target.Arch == "amd64" {
				darwinAmd64 = b.Path
			}
			if b.Target.OS == "darwin" && b.Target.Arch == "arm64" {
				darwinArm64 = b.Path
			}
		}
		if darwinAmd64 != "" && darwinArm64 != "" {
			universalDir := filepath.Join("bin", "darwin_universal")
			_ = os.MkdirAll(universalDir, 0755)
			universalOut := filepath.Join(universalDir, "mdn")
			lipoCmd := exec.Command("lipo", "-create", "-output", universalOut, darwinAmd64, darwinArm64)
			if err := lipoCmd.Run(); err == nil {
				fmt.Printf("==> Created macOS Universal Binary at %s\n", universalOut)
				builtBinaries = append(builtBinaries, BuiltBinary{
					Target: Target{OS: "darwin", Arch: "universal"},
					Path:   universalOut,
					Name:   "mdn",
				})
			}
		}
	}

	fmt.Println("\nAll binaries built successfully!")

	if *doPackage {
		fmt.Println("\n======================================================")
		fmt.Println(" Packaging Release Archives & Checksums")
		fmt.Println("======================================================")

		distDir := "dist"
		if err := os.MkdirAll(distDir, 0755); err != nil {
			fmt.Printf("Error creating %s: %v\n", distDir, err)
			os.Exit(1)
		}

		type ChecksumEntry struct {
			FileName string
			SHA256   string
		}
		var checksums []ChecksumEntry

		for _, b := range builtBinaries {
			archiveBase := fmt.Sprintf("mdn-%s-%s-%s", version, b.Target.OS, b.Target.Arch)
			var archivePath string
			var err error

			if b.Target.OS == "windows" {
				archivePath = filepath.Join(distDir, archiveBase+".zip")
				fmt.Printf("==> Creating %s...\n", archivePath)
				err = createZip(archivePath, b.Path, b.Name)
			} else {
				archivePath = filepath.Join(distDir, archiveBase+".tar.gz")
				fmt.Printf("==> Creating %s...\n", archivePath)
				err = createTarGz(archivePath, b.Path, b.Name)
			}

			if err != nil {
				fmt.Printf("Error packaging %s: %v\n", archivePath, err)
				os.Exit(1)
			}

			hash, err := fileSHA256(archivePath)
			if err != nil {
				fmt.Printf("Error hashing %s: %v\n", archivePath, err)
				os.Exit(1)
			}
			checksums = append(checksums, ChecksumEntry{
				FileName: filepath.Base(archivePath),
				SHA256:   hash,
			})
		}

		checksumFilePath := filepath.Join(distDir, "checksums.txt")
		checksumFile, err := os.Create(checksumFilePath)
		if err != nil {
			fmt.Printf("Error creating checksums.txt: %v\n", err)
			os.Exit(1)
		}
		for _, c := range checksums {
			_, _ = fmt.Fprintf(checksumFile, "%s  %s\n", c.SHA256, c.FileName)
		}
		checksumFile.Close()

		fmt.Printf("==> Generated checksums in %s\n", checksumFilePath)
		fmt.Println("\nRelease packaging complete! Files in dist/:")
		entries, _ := os.ReadDir(distDir)
		for _, e := range entries {
			info, _ := e.Info()
			fmt.Printf("  - %-40s (%d bytes)\n", e.Name(), info.Size())
		}
	}
}

func createTarGz(archivePath, srcFilePath, nameInArchive string) error {
	outFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(srcInfo, "")
	if err != nil {
		return err
	}
	header.Name = nameInArchive
	header.Mode = 0755

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, srcFile)
	return err
}

func createZip(archivePath, srcFilePath, nameInArchive string) error {
	outFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(srcInfo)
	if err != nil {
		return err
	}
	header.Name = nameInArchive
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, srcFile)
	return err
}

func fileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
