package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
)

var version string

var showVersion = flag.Bool("version", false, "print version and exit")
var socketPath = flag.String("docker-socket", "/var/run/docker.sock", "path to socket for Docker")
var outputJson = flag.Bool("json", false, "output in JSON format")
var outputLong = flag.Bool("l", false, "output in detailed format")

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `
Usage: lsdvol [OPTION]... [ID or NAME]
Lists volumes in use by a Docker container.

If no ID or NAME is specified, the program is assumed to be
executed from within a container and the container ID will be
autodetected.

Options:`)
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("lsdvol %s\n", version)
		os.Exit(0)
	}

	docker, err := NewDockerClient(*socketPath)
	if err != nil {
		log.Fatalln(err)
	}

	var containerId = flag.Arg(0)
	if containerId == "" {
		containerId, err = detectContainerId()
		if err != nil {
			log.Fatalln(err)
		}
	}

	volumes, err := docker.VolumesFor(containerId)
	if err != nil {
		log.Fatalln(err)
	}
	if *outputLong {
		printVolumesLong(volumes)
		os.Exit(0)
	}
	if *outputJson {
		printVolumesJson(volumes)
		os.Exit(0)
	}
	printVolumes(volumes)
}

func printVolumesLong(volumes []DockerVolume) {
	fmt.Printf("%d volume(s)\n", len(volumes))
	for _, v := range volumes {
		var w = ""
		if v.Writable {
			w = "w"
		}
		fmt.Printf("r%s  %s\n", w, v.Path)
	}
}

func printVolumesJson(volumes []DockerVolume) {
	var enc = json.NewEncoder(os.Stdout)
	var err = enc.Encode(volumes)
	if err != nil {
		log.Fatalln(err)
	}
}

func printVolumes(volumes []DockerVolume) {
	for _, v := range volumes {
		fmt.Println(v.Path)
	}
}

func detectContainerId() (string, error) {
	var file, err = os.Open("/proc/self/cgroup")
	defer file.Close()
	if err != nil {
		return "", err
	}
	var id = regexp.MustCompile("[a-f0-9]{64}")
	var scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		if x := id.FindString(scanner.Text()); x != "" {
			return x, nil
		}
	}
	return "", fmt.Errorf("Unable to determine running container id")
}

const DockerAPIVersion = "v1.14"

type DockerClient struct {
	http *http.Client
}

type DockerVolume struct {
	Path     string `json:"path"`
	Writable bool   `json:"writable"`
}

func NewDockerClient(socketPath string) (client *DockerClient, err error) {
	info, err := os.Stat(socketPath)
	if err != nil {
		return
	}
	if info.Mode()&os.ModeSocket != os.ModeSocket {
		return nil, fmt.Errorf("%s is not a socket", socketPath)
	}
	client = &DockerClient{
		http: &http.Client{Transport: &http.Transport{Dial: func(network string, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}}}}

	isCompatible, err := client.hasCompatibleVersion()
	if !isCompatible {
		err = fmt.Errorf("Docker Engine not compatible with Remote API version %s", DockerAPIVersion)
	}
	return
}

func (c *DockerClient) get(url string) (*http.Response, error) {
	return c.http.Get(fmt.Sprintf("http://docker/%s%s", DockerAPIVersion, url))
}

func (c *DockerClient) hasCompatibleVersion() (bool, error) {
	var res, err = c.get("/info")
	if err != nil {
		return false, err
	}
	return res.StatusCode == http.StatusOK, nil
}

func (c *DockerClient) VolumesFor(containerId string) ([]DockerVolume, error) {
	type Info struct {
		VolumesRW map[string]bool
	}
	var res, err = c.get(fmt.Sprintf("/containers/%s/json", containerId))
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("No container with id %s was found.", containerId)
	}
	var dec = json.NewDecoder(res.Body)
	var info Info
	err = dec.Decode(&info)
	if err != nil {
		return nil, err
	}
	var volumes = []DockerVolume{}
	for path, w := range info.VolumesRW {
		volumes = append(volumes, DockerVolume{path, w})
	}
	return volumes, nil
}
