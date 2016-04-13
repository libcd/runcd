package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/libcd/libcd"
	"github.com/libcd/libcd/docker"
	"github.com/libcd/libyaml"
	"github.com/libcd/libyaml/builtin"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "runcd"
	app.Usage = "runcd provides command line tools for the libcd runtime"
	app.Commands = []cli.Command{
		{
			Name:   "compile",
			Usage:  "compile the yaml file",
			Action: compileCmd,
		},
		{
			Name:   "exec",
			Usage:  "execute the compiled file",
			Action: executeCmd,
		},
	}

	app.Run(os.Args)
}

func compileCmd(c *cli.Context) {
	filename := c.Args().First()
	filedata, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Unable to open file %s. %s", filename, err)
		return
	}

	trans := []libyaml.Transform{
		builtin.NewWorkspaceOp("/drone", "drone/src"),
		builtin.NewNormalizeOp("plugins"),
		builtin.NewPullOp(false),
		builtin.NewEnvOp(map[string]string{"CI": "true"}),
		builtin.NewValidateOp(false, []string{"plugins/*"}),
		builtin.NewShellOp(builtin.Linux_adm64),
		builtin.NewArgsOp(),
		builtin.NewPodOp("drone_"),
		// builtin.NewCloneOp("plugins/drone-git"),
		// builtin.NewCacheOp("plugins/drone-cache", "/var/lib/drone/cache"),
	}

	compiler := libyaml.New()
	compiler.Transforms(trans)
	spec, err := compiler.Compile(filedata)
	if err != nil {
		fmt.Printf("Unable to compile file %s. %s", filename, err)
		return
	}

	out, _ := json.MarshalIndent(spec, "", "  ")
	os.Stdout.Write(out)
}

func executeCmd(c *cli.Context) {
	filename := c.Args().First()
	filedata, err := readFileOrStdin(filename)
	if err != nil {
		fmt.Printf("Unable to read file from disk or stdin. %s", err)
		return
	}

	spec, err := libcd.Parse(filedata)
	if err != nil {
		fmt.Printf("Unable to open file %s. %s", filename, err)
		return
	}

	conf := libcd.Config{
		Engine: docker.MustEnv(),
	}
	runner := conf.Runner(libcd.NoContext, spec)
	if err := runner.Run(); err != nil {
		fmt.Println(err)
		return
	}

	pipe := runner.Pipe()
	for {
		line := pipe.Next()
		if line == nil {
			break
		}
		fmt.Println(line)
	}

	if err := runner.Wait(); err != nil {
		fmt.Println(err)
		return
	}
}

func readFileOrStdin(filename string) ([]byte, error) {
	if filename == "" {
		return ioutil.ReadAll(os.Stdin)
	} else {
		return ioutil.ReadFile(filename)
	}
}
