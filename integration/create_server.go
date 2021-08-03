package integration

import (
	"archive/tar"
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
	"time"

	"github.com/facebookincubator/zk/flw"
)

const zkFormatURL = "https://archive.apache.org/dist/zookeeper/zookeeper-%s/apache-zookeeper-%s-bin.tar.gz"
const defaultArchiveName = "server.tar.gz"
const defaultConfigName = "zk.cfg"
const maxRetries = 10

type ZKServer struct {
	Version string
	Config  *ServerConfig

	cmd *exec.Cmd
}

// NewZKServer creates a new ZKServer instance using the given version and config map.
// Contents of the config struct are written to a file which will later be used by the ZK binary.
func NewZKServer(version string, config *ServerConfig) (*ZKServer, error) {
	file, err := os.Create(defaultConfigName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err = config.Marshall(file); err != nil {
		return nil, fmt.Errorf("error writing config to file: %v", err)
	}

	return &ZKServer{
		Version: version,
		Config:  config,
	}, nil
}

// Run downloads the specified Zookeeper version from Apache's website,
// extracts it and runs it using the config specified by configPath.
// The server runs in the background until Shutdown is called.
func (server *ZKServer) Run() error {
	zkURL := fmt.Sprintf(zkFormatURL, server.Version, server.Version)

	workdir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %s", err.Error())
	}
	archivePath := filepath.Join(workdir, defaultArchiveName)
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		err = downloadToFile(zkURL, archivePath)
		if err != nil {
			return fmt.Errorf("error downloading file: %s\n", err)
		}
		log.Printf("successfully downloaded archive %s\n", defaultArchiveName)
	}

	dirName := "apache-zookeeper-" + server.Version + "-bin"
	if _, err := os.Stat(filepath.Join(workdir, dirName)); os.IsNotExist(err) {
		_, err := extractTarGz(archivePath)
		if err != nil {
			return fmt.Errorf("error extracting file: %s\n", err)
		}
	}

	serverScriptPath := filepath.Join(workdir, dirName, "bin/zkServer.sh")
	err = os.Chmod(serverScriptPath, 0777)
	if err != nil {
		return fmt.Errorf("error changing server script permissions: %s\n", err)
	}

	server.cmd = exec.Command(serverScriptPath, "start-foreground", filepath.Join(workdir, defaultConfigName))
	server.cmd.Stdout = os.Stdout
	server.cmd.Stderr = os.Stderr

	err = server.cmd.Start()
	if err != nil {
		return fmt.Errorf("error executing server command: %s\n", err)
	}

	if err = waitForStart([]string{"0.0.0.0"}, maxRetries, time.Second); err != nil {
		return err
	}

	return nil
}

func (server *ZKServer) Shutdown() error {
	log.Printf("Shutdown() called, killing server process")
	return server.cmd.Process.Kill()
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
	defer file.Close()

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

// waitForStart blocks until all servers from the specified addresses slice are up
func waitForStart(addresses []string, maxRetry int, interval time.Duration) error {
	client := &flw.Client{Timeout: time.Second}
	for i := 0; i < maxRetry; i++ {
		_, ok := client.Srvr(addresses)
		if ok {
			return nil
		}
		time.Sleep(interval)
	}

	return fmt.Errorf("unable to verify health of servers")
}
