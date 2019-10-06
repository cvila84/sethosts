package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path"
)

type HostsEntry struct {
	IP       string
	HostName string
}

var rootCmd = &cobra.Command{
	Use:   "sethosts",
	Short: "sethosts write a Windows hosts file from a JSON input",
	Long:  "sethosts write a Windows hosts file from a JSON input",
	Run:   rootRun,
}

func rootRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatalln("Missing JSON hosts entries")
	}
	var entries []HostsEntry
	if err := json.Unmarshal([]byte(args[0]), &entries); err != nil {
		log.Println("Bad JSON hosts entries")
		log.Fatalln(err)
	}
	if len(entries) == 0 {
		os.Exit(0)
	}
	dir := path.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc")
	hosts := path.Join(dir, "hosts")
	hostsBak := path.Join(dir, "hosts.bak")
	f1, _ := os.Open(hosts)
	defer f1.Close()
	os.Remove(hostsBak)
	f2, _ := os.Create(hostsBak)
	defer f2.Close()
	io.Copy(f2, f1)
	os.Remove(hosts)
	f3, _ := os.Create(hosts)
	defer f3.Close()
	w := bufio.NewWriter(f3)
	for _, entry := range entries {
		if len(entry.IP) > 0 && len(entry.HostName) > 0 {
			fmt.Printf("%s\t%s\n", entry.IP, entry.HostName)
			fmt.Fprintf(w, "%s\t%s\n", entry.IP, entry.HostName)
		} else {
			log.Printf("Ignored malformed JSON host entry IP: %s and HostName: %s\n", entry.IP, entry.HostName)
		}
	}
	w.Flush()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
