package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

type HostsEntry struct {
	Ip       string `yaml:"ip"`
	Hostname string `yaml:"hostname"`
}

var rootCmd = &cobra.Command{
	Use:   "sethosts",
	Short: "sethosts write a Windows hosts file from a YAML input",
	Long:  "sethosts write a Windows hosts file from a YAML input",
	Run:   rootRun,
}

var dryRun bool
var pause bool
var verbose bool

func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do not actually write the result to hosts file, just display it")
	rootCmd.Flags().BoolVarP(&pause, "pause", "p", false, "Request user to type Enter to exit")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log more information on console")
}

func rootRun(cmd *cobra.Command, args []string) {
	entries := strings.Join(args[:], " ")
	if len(entries) == 0 {
		if _, err := fmt.Scanln(&entries); err != nil {
			log.Printf("Cannot read YAML hosts entries from standard input: %s\n", err)
			leave(1)
		}
	}
	if len(entries) == 0 {
		log.Println("Missing YAML hosts entries")
		leave(1)
	}
	if err := run(entries); err != nil {
		log.Printf("Cause: %s\n", err)
		leave(1)
	}
	leave(0)
}

func run(yamlEntries string) (rerr error) {
	if verbose {
		log.Printf("YAML hosts entries: %s\n", yamlEntries)
	}
	var entries []HostsEntry
	if err := yaml.Unmarshal([]byte(yamlEntries), &entries); err != nil {
		log.Println("Bad YAML hosts entries")
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if dryRun {
		for _, entry := range entries {
			if len(entry.Ip) > 0 && len(entry.Hostname) > 0 {
				fmt.Printf("%s\t%s\n", entry.Ip, entry.Hostname)
			} else {
				log.Printf("Ignored malformed YAML host entry IP: %s and HostName: %s\n", entry.Ip, entry.Hostname)
			}
		}
	} else {
		dir := path.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc")
		hosts := path.Join(dir, "hosts")
		hostsBak := path.Join(dir, "hosts.bak")
		f1, err1 := os.Open(hosts)
		if err1 != nil {
			log.Println("Cannot open hosts file")
			return err1
		}
		defer func() {
			if err := f1.Close(); err != nil {
				log.Println("Unexpected error")
				rerr = err
			}
		}()
		f2, err2 := os.Create(hostsBak)
		if err2 != nil {
			log.Println("Cannot create backup hosts file")
			return err2
		}
		defer func() {
			if err := f2.Close(); err != nil {
				log.Println("Unexpected error")
				rerr = err
			}
		}()
		if _, err := io.Copy(f2, f1); err != nil {
			log.Println("Cannot backup hosts file")
			return err
		}
		f3, err3 := os.Create(hosts)
		if err3 != nil {
			log.Println("Cannot create hosts file")
			return err3
		}
		defer func() {
			if err := f3.Close(); err != nil {
				log.Println("Unexpected error")
				rerr = err
			}
		}()
		w := bufio.NewWriter(f3)
		for _, entry := range entries {
			if len(entry.Ip) > 0 && len(entry.Hostname) > 0 {
				if _, err := fmt.Fprintf(w, "%s\t%s\n", entry.Ip, entry.Hostname); err != nil {
					log.Println("Cannot write hosts file")
					return err
				}
			} else {
				log.Printf("Ignored malformed YAML host entry IP: %s and HostName: %s\n", entry.Ip, entry.Hostname)
			}
		}
		if err := w.Flush(); err != nil {
			log.Println("Cannot write hosts file")
			return err
		}
	}
	return nil
}

func leave(exitCode int) {
	if pause {
		fmt.Println("Type <ENTER> to exit")
		fmt.Scanln()
	}
	os.Exit(exitCode)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
