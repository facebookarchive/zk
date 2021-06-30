package integration

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
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

const zkFormatURL = "https://archive.apache.org/dist/zookeeper/zookeeper-%s/apache-zookeeper-%s-bin.tar.gz"
const defaultArchiveName = "server.tar.gz"

// RunZookeeperServer downloads the specified Zookeeper version from Apache's website,
// extracts it and runs it using the config specified by configPath.
// ReadyChan is used to notify the caller that the server has been initialized,
// and the caller can send to exitChan in order to terminate the server.
func RunZookeeperServer(version, configPath string, readyChan, exitChan chan struct{}) error {
	zkURL := fmt.Sprintf(zkFormatURL, version, version)

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %s", err.Error())
	}
	configPath = filepath.Join(workdir, configPath) // zk config file is expected to be in the project's root dir
	archivePath := filepath.Join(workdir, defaultArchiveName)
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		err = downloadToFile(zkURL, archivePath)
		if err != nil {
			return fmt.Errorf("error downloading file: %s\n", err)
		}
		log.Printf("successfully downloaded archive %s\n", defaultArchiveName)
	}

	root, err := extractTarGz(archivePath)
	if err != nil {
		return fmt.Errorf("error extracting file: %s\n", err)
	}

	serverScriptPath := filepath.Join(workdir, root, "bin/zkServer.sh")
	err = os.Chmod(serverScriptPath, 0777)
	if err != nil {
		return fmt.Errorf("error changing server script permissions: %s\n", err)
	}

	cmd := exec.Command(serverScriptPath, "start-foreground", configPath)
	stdout, err := cmd.StdoutPipe()
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("error executing server command: %s\n", err)
	}

	go handleOutput(stdout)

	readyChan <- struct{}{} // notify caller that server has started successfully

	// wait for exit signal from caller
	select {
	case <-exitChan:
		log.Printf("got exit signal, shutting down")
		break
	}

	return nil
}

func handleOutput(stdout io.ReadCloser) {
	stdoutScanner := bufio.NewScanner(stdout)
	stdoutScanner.Split(bufio.ScanLines)
	for stdoutScanner.Scan() {
		m := stdoutScanner.Text()
		log.Println(m)
	}
}

func downloadToFile(sourceURL, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	response, err := http.Get(sourceURL)
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