package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/mitchellh/cli"
)

const (
	ExitCodeOK        int = iota //0
	ExitCodeError     int = iota //1
	ExitCodeFileError int = iota //2
)

type BrowseCommand struct {
}

func (c *BrowseCommand) Synopsis() string {
	return "Browse repository"
}

func (c *BrowseCommand) Help() string {
	return "Usage: lab brewse [option]"
}

func (c *BrowseCommand) Run(args []string) int {
	var debug bool

	flags := flag.NewFlagSet("add", flag.ContinueOnError)
	flags.BoolVar(&debug, "debug", false, "Run as DEBUG mode")

	// Get remote repositorys
	remotes := gitOutputs("git", []string{"remote"})

	// Remote repository is not registered
	if len(remotes) == 0 {
		fmt.Println("No remote setting in this repository")
		return ExitCodeError
	}

	var gitlabUrls []string
	for _, remote := range remotes {
		url := gitOutput("git", []string{"remote", "get-url", remote})
		}

		remoteUrl, err := NewRemoteUrl(url)
		if err != nil {
			fmt.Println("No remote setting in this repository.")
			return ExitCodeError
		}

		if strings.HasPrefix(remoteUrl.Domain, "gitlab") {
			gitlabUrls = append(gitlabUrls, remoteUrl.ConcatUrl())
		}
	}

	var gitlabUrl string
	if len(gitlabUrls) > 0 {
		gitlabUrl = gitlabUrls[0]
	} else {
		fmt.Println("Not a cloned repository from gitlab.")
		return ExitCodeError
	}

	browser := searchBrowserLauncher(runtime.GOOS)
	cmdOutput(browser, []string{gitlabUrl})

	return ExitCodeOK
}

type RemoteUrl struct {
	Url        string
	Domain     string
	User       string
	Repository string
}

func (r *RemoteUrl) ConcatUrl() string {
	params := strings.Join([]string{r.Domain, r.User, r.Repository}, "/")
	return "https://" + params
}

func NewRemoteUrl(url string) (*RemoteUrl, error) {
	var (
		otherScheme string
		domain      string
		user        string
		repository  string
	)

	if strings.HasPrefix(url, "ssh") {
		// ssh://git@gitlab.com/lighttiger2505/lab.git
		otherScheme = strings.Split(url, "@")[1]
		otherScheme = strings.TrimSuffix(otherScheme, ".git")
	} else if strings.HasPrefix(url, "https") {
		// https://github.com/lighttiger2505/lab
		otherScheme = strings.Split(url, "//")[1]
	} else {
		return nil, errors.New(fmt.Sprintf("Invalid remote url: %s", url))
	}

	splitUrl := strings.Split(otherScheme, "/")
	domain = splitUrl[0]
	user = splitUrl[1]
	repository = splitUrl[2]

	return &RemoteUrl{
		Url:        url,
		Domain:     domain,
		User:       user,
		Repository: repository,
	}, nil
}

func gitOutput(name string, args []string) string {
	return gitOutputs(name, args)[0]
}

func gitOutputs(name string, args []string) []string {
	var out = cmdOutput(name, args)
	var outs []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) != "" {
			outs = append(outs, string(line))
		}
	}
	return outs
}

func cmdOutput(name string, args []string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	return string(out)
}

func searchBrowserLauncher(goos string) (browser string) {
	switch goos {
	case "darwin":
		browser = "open"
	case "windows":
		browser = "cmd /c start"
	default:
		candidates := []string{
			"xdg-open",
			"cygstart",
			"x-www-browser",
			"firefox",
			"opera",
			"mozilla",
			"netscape",
		}
		for _, b := range candidates {
			path, err := exec.LookPath(b)
			if err == nil {
				browser = path
				break
			}
		}
	}
	return browser
}

func main() {
	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]

	c.Commands = map[string]cli.CommandFactory{
		"browse": func() (cli.Command, error) {
			return &BrowseCommand{}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
