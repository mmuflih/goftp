package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jlaffaye/ftp"
	"github.com/mmuflih/envgo/conf"
)

func uploadDirectory(conn *ftp.ServerConn, localDir string, remoteDir string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create remote directory if it is a directory
		if info.IsDir() {
			remotePath := filepath.Join(remoteDir, strings.TrimPrefix(path, localDir))
			return conn.MakeDir(remotePath)
		}

		// Upload file if it is a file
		if !info.IsDir() {
			localFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer localFile.Close()

			remotePath := filepath.Join(remoteDir, strings.TrimPrefix(path, localDir))
			remotePath = filepath.ToSlash(remotePath) // Ensure remote path is Unix-style

			err = conn.Stor(remotePath, localFile)
			if err != nil {
				return err
			}
			fmt.Printf("Uploaded %s to %s\n", path, remotePath)
		}

		return nil
	})
}

func GetCurrentDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(dir), nil
}

func main() {
	conf := conf.NewConfig()

	ftpServer := conf.GetString("server")
	user := conf.GetString("user")
	password := conf.GetString("password")
	ftpDir := conf.GetString("ftp_dir")

	conn, err := ftp.Dial(ftpServer)
	if err != nil {
		fmt.Println("Error connecting to FTP server:", err)
		return
	}
	defer conn.Quit()

	err = conn.Login(user, password)
	if err != nil {
		fmt.Println("Error logging in to FTP server:", err)
		return
	}

	curDir, err := GetCurrentDir()
	if err != nil {
		fmt.Println("Curr dir error", err)
		return
	}

	localDir := "./"
	remoteDir := ftpDir + "/" + curDir

	err = uploadDirectory(conn, localDir, remoteDir)
	if err != nil {
		fmt.Println("Error uploading directory:", err)
		return
	}

	fmt.Println("Directory uploaded successfully")
}

func ensureRemoteDirExists(conn *ftp.ServerConn, remoteDir string) error {
	err := conn.MakeDir(remoteDir)
	if err != nil && !strings.Contains(err.Error(), "550") {
		return err
	}
	return nil
}
