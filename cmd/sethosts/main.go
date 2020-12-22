package main

import (
	"fmt"
	"github.com/goodhosts/hostsfile"
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
var merge bool

func init() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Do not actually write the result to hosts file, just display it")
	rootCmd.Flags().BoolVarP(&pause, "pause", "p", false, "Request user to type Enter to exit")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Log more information on console")
	rootCmd.Flags().BoolVarP(&merge, "merge", "m", false, "Merge new entries with already existing ones in hosts file")
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
	if !dryRun {
		err := backupHosts()
		if err != nil {
			return err
		}
	}
	hosts, err := hostsfile.NewHosts()
	if err != nil {
		log.Println("Cannot read hosts file")
		return err
	}
	for _, entry := range entries {
		if len(entry.Ip) > 0 && len(entry.Hostname) > 0 {
			err := hosts.Add(entry.Ip, entry.Hostname)
			if err != nil {
				log.Println("Malformed YAML host entry")
				return err
			}
		} else {
			log.Printf("Ignored malformed YAML host entry IP: %s and HostName: %s\n", entry.Ip, entry.Hostname)
		}
	}
	hosts.Clean()
	if verbose {
		for _, line := range hosts.Lines {
			log.Println(line.Raw)
		}
	}
	if !dryRun {
		err := hosts.Flush()
		if err != nil {
			log.Println("Cannot write hosts file")
			return err
		}
	}
	return nil
}

func backupHosts() (rerr error) {
	dir := path.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc")
	hostsFile := path.Join(dir, "hosts")
	hostsBakFile := path.Join(dir, "hosts.bak")
	f1, err1 := os.Open(hostsFile)
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
	f2, err2 := os.Create(hostsBakFile)
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
	if !merge {
		f3, err3 := os.Create(hostsFile)
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
