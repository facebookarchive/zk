package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const defaultZkURL = "https://archive.apache.org/dist/zookeeper/zookeeper-3.6.2/apache-zookeeper-3.6.2-bin.tar.gz"
const defaultConfigPath = "zk.cfg"
const defaultArchiveName = "server.tar.gz"

func main() {
	zkURL := flag.String("url", defaultZkURL, "Download URL for Apache Zookeeper server installation")
	zkConfigPath := flag.String("config", defaultConfigPath, "Zookeeper config filename")
	archiveName := flag.String("archive", defaultArchiveName, "Downloaded archive name")
	flag.Parse()

	workdir, err := os.Getwd()
	if err != nil {
		log.Fatalf("error getting working directory: %s", err.Error())
	}
	*zkConfigPath = filepath.Join(workdir, *zkConfigPath) // zk config file is expected to be in the project's root dir

	err = downloadToFile(*zkURL, filepath.Join(workdir, *archiveName))
	if err != nil {
		log.Fatalf("error downloading file: %s\n", err)
	}

	log.Printf("successfully downloaded archive %s\n", *archiveName)
	root, err := extractTarGz(filepath.Join(workdir, *archiveName))
	if err != nil {
		log.Fatalf("error extracting file: %s\n", err)
	}

	serverScriptPath := filepath.Join(workdir, root, "bin/zkServer.sh")
	err = os.Chmod(serverScriptPath, 0777)
	if err != nil {
		log.Fatalf("error changing server script permissions: %s\n", err)
	}

	cmd := exec.Command("/bin/bash", serverScriptPath, "start-foreground", *zkConfigPath)
	stdout, err := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		log.Fatalf("error executing server command: %s\n", err)
	}

	stdoutScanner := bufio.NewScanner(stdout)
	stdoutScanner.Split(bufio.ScanLines)
	for stdoutScanner.Scan() {
		m := stdoutScanner.Text()
		log.Println(m)
	}
	cmd.Wait()
}

func downloadToFile(sourceURL, filepath string) error {
	proxyClient := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	response, err := proxyClient.Get(sourceURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("wrong status code while downloading from %s: %d", sourceURL, response.StatusCode)
	}

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func extractTarGz(src string) (string, error) {
	var rootPath string
	isRootDir := true

	file, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("error opening archive file: %v", err.Error())
	}

	reader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("error creating gzip reader: %v", err.Error())
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("error reading from tar archive: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if isRootDir {
				rootPath = getRootFromPath(header.Name)
				isRootDir = false
			}
			if err := os.MkdirAll(header.Name, os.ModePerm); err != nil {
				return "", fmt.Errorf("mkdir failed: %s", err.Error())
			}
		case tar.TypeReg:
			if err = ensureBaseDir(header.Name); err != nil {
				return "", fmt.Errorf("error creating file from tar header: %s", err.Error())
			}
			outFile, err := os.Create(header.Name)
			if err != nil {
				return "", fmt.Errorf("error creating file from tar header: %s", err.Error())
			}

			if _, err = io.Copy(outFile, tarReader); err != nil {
				return "", fmt.Errorf("error copying file from archive: %s", err.Error())
			}
			outFile.Close()
		default:
			return "", fmt.Errorf("unknown header type detected while extracting from archive %s", src)
		}
	}

	return rootPath, nil
}

func getRootFromPath(filepath string) string {
	idx := strings.Index(filepath, "/")
	if idx == -1 {
		return filepath
	}

	return filepath[:idx]
}

func ensureBaseDir(filepath string) error {
	baseDir := path.Dir(filepath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}
