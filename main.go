package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/libcd/libcd"
	"github.com/libcd/libcd/docker"
	"github.com/libcd/libcd/graph"
	"github.com/libcd/libyaml"
	"github.com/libcd/libyaml/builtin"
	"github.com/pkg/browser"

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
			Action: errorCommand(compileCmd),
		},
		{
			Name:   "exec",
			Usage:  "execute the compiled file",
			Action: errorCommand(executeCmd),
		},
		{
			Name:   "graph",
			Usage:  "generate the graphviz file",
			Action: errorCommand(graphCmd),
		},
	}

	app.Run(os.Args)
}

func compileCmd(c *cli.Context) error {
	filename := c.Args().First()
	filedata, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
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
		return err
	}

	out, _ := json.MarshalIndent(spec, "", "  ")
	os.Stdout.Write(out)

	return nil
}

func executeCmd(c *cli.Context) error {
	filename := c.Args().First()
	filedata, err := readFileOrStdin(filename)
	if err != nil {
		return err
	}

	spec, err := libcd.Parse(filedata)
	if err != nil {
		return err
	}

	conf := libcd.Config{
		Engine: docker.MustEnv(),
	}
	runner := conf.Runner(libcd.NoContext, spec)
	if err := runner.Run(); err != nil {
		return err
	}

	pipe := runner.Pipe()
	for {
		line := pipe.Next()
		if line == nil {
			break
		}
		fmt.Println(line)
	}

	return runner.Wait()
}

func graphCmd(c *cli.Context) error {
	filename := c.Args().First()
	filedata, err := readFileOrStdin(filename)
	if err != nil {
		return err
	}

	spec, err := libcd.Parse(filedata)
	if err != nil {
		return err
	}

	cmd := exec.Command("dot", "-Tsvg")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	println(string(graph.Create(spec)))

	graph.WriteTo(spec, in)
	in.Close()

	browser.OpenReader(out)
	return cmd.Wait()
}

func readFileOrStdin(filename string) ([]byte, error) {
	if filename == "" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(filename)
}

func errorCommand(fn func(*cli.Context) error) func(*cli.Context) {
	return func(c *cli.Context) {
		err := fn(c)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
