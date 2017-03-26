package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
	"github.com/rwcarlsen/goexif/exif"
)

const (
	ExpandDimension = 800
)

func getConnection() (*ssh.Client, error) {
	key, err := ioutil.ReadFile("/home/artur/.ssh/id_rsa")
	if err != nil {
		log.Fatal(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("parse key failed:%v", err)
	}
	config := &ssh.ClientConfig{
		User: "ubuntu",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	return ssh.Dial("tcp", "52.73.15.72:22", config)
}

func execRemote(sshClient *ssh.Client, command string) error {
	session, sessionErr := sshClient.NewSession()
	defer session.Close()
	defer func() {
		if e := recover(); e != nil {
			log.Println("recovered", e)
			// whatever
		}
	}()

	if sessionErr != nil {
		return sessionErr
	}
	mkdirErr := session.Run(command)
	if mkdirErr != nil {
		return mkdirErr
	}
	return nil
}

var (
	startAt     = flag.Int("start", 0, "start at image id")
	concurrency = flag.Int("concurrency", 1, "number of concurrent goroutines uploading images")
)

var dirMade = map[string]bool{}

func init() {
	flag.Parse()

}

func main() {
	fmt.Println("Connecting to server")
	sshClient, sshErr := getConnection()

	if sshErr != nil {
		panic(sshErr)
	}

	// Load desired start id if we haven't set it with a flag

	if *startAt == 0 {
		session, sessionErr := sshClient.NewSession()
		if sessionErr != nil {
			panic(sessionErr)
		}

		var (
			stdout bytes.Buffer
		)

		session.Stdout = &stdout

		runErr := session.Run("./arturphotos_getstart")

		session.Close()

		if runErr != nil {
			panic(runErr)
		}

		var result struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Id    int `json:"id"`
		}

		jErr := json.Unmarshal(stdout.Bytes(), &result)

		if jErr == nil {
			*startAt = result.Id + 1
		}

		if result.Id == 0 {
			panic("Unable to fetch latest id")
		}

		fmt.Println("Starting at id", *startAt)
	}

	images, imagesGlobErr := filepath.Glob("/run/user/*/gvfs/*/DCIM/*/*.JPG")

	if imagesGlobErr != nil {
		panic(imagesGlobErr)
	}

	if len(images) == 0 {
		panic("Unable to find photos on disk")
	}

	var wg sync.WaitGroup

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go syncImages(sshClient, images, *concurrency, i, &wg)
	}

	wg.Wait()
}

func upload(sshClient *ssh.Client, src, base, dest string) {
	sftpClient, sftpErr := sftp.NewClient(sshClient)

	if sftpErr != nil {
		panic(sftpErr)
	}

	// Resize to album size
	outPathTmp := filepath.Join("/tmp/artur.co/upload", base+".tmp")
	outFile, outErr := sftpClient.Create(outPathTmp)
	srcFile, srcErr := os.Open(src)

	if outErr != nil {
		panic(outErr)
	}
	if srcErr != nil {
		panic(srcErr)
	}

	log.Println("Copying", dest)

	_, copyErr := io.Copy(outFile, srcFile)
	if copyErr != nil {
		panic(copyErr)
	}

	srcFile.Close()
	outFile.Close()
	sftpClient.Close()

	mvErr := execRemote(sshClient, "mv "+outPathTmp+" "+dest)
	if mvErr != nil {
		log.Println(mvErr)
	}

}

func syncImages(sshClient *ssh.Client, images []string, ofn int, offset int, wg *sync.WaitGroup) {

	for _, path := range images {
		base := filepath.Base(path)

		if base[0:3] != "IMG" {
			log.Println("Skipping weird filename " + path)
			continue
		}

		id, iderr := strconv.ParseInt(base[4:8], 10, 32)
		if iderr != nil {
			panic(iderr)
		}

		// Each worker is only responsible for their modulo value
		if int(id)%ofn != offset {
			continue
		}

		if id < int64(*startAt) {
			continue
		}

		f, ferr := os.Open(path)
		if ferr != nil {
			log.Println(ferr)
			continue
		}
		ex, err := exif.Decode(f)
		if err != nil {
			log.Println(err)
			continue
		}

		datetime, dterr := ex.Get(exif.DateTime)
		if dterr != nil {
			continue
		}

		timestamp, parseErr := time.Parse("2006:01:02 15:04:05", string(datetime.Val)[0:19])
		if parseErr != nil {
			panic(parseErr)
		}

		photoDir := fmt.Sprintf("/mnt/raw/photos/%d/%02d", timestamp.Year(), timestamp.Month())
		fmt.Printf("%s (%s) -> %s\n", base, timestamp, photoDir)

		mkdirErr := execRemote(sshClient, "mkdir -p "+photoDir)
		if mkdirErr != nil {
			panic(mkdirErr)
		}

		log.Println("Opening", path)

		// Upload raw
		upload(sshClient, path, base, filepath.Join(photoDir, base))

		log.Println("Done", base)
	}

	wg.Done()
}
