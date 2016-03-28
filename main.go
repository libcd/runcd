package main

import (
	// "encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/libcd/libcd"
	"github.com/libcd/libcd/docker"
	"github.com/libcd/libcd/graph"
	"github.com/pkg/browser"
	// "github.com/libcd/libyaml"

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
	// filename := c.Args().First()
	// filedata, err := ioutil.ReadFile(filename)
	// if err != nil {
	// 	fmt.Printf("Unable to open file %s. %s", filename, err)
	// 	return
	// }

	// parsed, err := libyaml.Parse(filedata)
	// if err != nil {
	// 	fmt.Printf("Unable to parse file %s. %s", filename, err)
	// 	return
	// }

	// spec, err := parsed.Compiler().Compile()
	// if err != nil {
	// 	fmt.Printf("Unable to compile file %s. %s", filename, err)
	// 	return
	// }

	// out, _ := json.MarshalIndent(spec, "", "  ")
	// os.Stdout.Write(out)
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
