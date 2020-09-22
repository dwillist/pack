package dive

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/spf13/cobra"

	"github.com/buildpacks/pack/internal/commands"
	"github.com/buildpacks/pack/internal/config"
	"github.com/buildpacks/pack/logging"
)

var (
	once         sync.Once
	appSingleton *app
)

// CreateBuilder creates a builder image, based on a builder config
func Dive(logger logging.Logger, cfg config.Config, client commands.PackClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dive <image-name>",
		Args:  cobra.ExactArgs(1),
		Short: "interactive exploration of image",
		RunE: commands.LogError(logger, func(cmd *cobra.Command, args []string) error {
			initConfig()
			imgName := args[0]

			// TODO: deleteme
			logfile, err := os.Create(filepath.Join("/tmp", "output", "pack.txt"))
			if err != nil {
				return err
			}
			logrus.SetOutput(logfile)

			logger.Infof("Building structures for %s\n", imgName)
			diveResult, err := client.Dive(imgName, true)
			if err != nil {
				return err
			}
			// create a GUI
			logger.Info("creating GUI")
			g, err := gocui.NewGui(gocui.OutputNormal)
			logger.Info("GUI created!!")
			if err != nil {
				return err
			}
			defer g.Close()

			logger.Info("starting app!")

			_, err = newApp(AppOptions{
				DiveResult: diveResult,
				GUI:        g,
				//Debug: true, doesn't work currently
			})
			if err != nil {
				panic(err)
			}

			if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
				logger.Errorf("main loop error: ", err)
				return err
			}

			return nil
		}),
	}
	return cmd
}

//
// Debug methods
//

func repeat(s string, c int) string {
	result := ""
	for i := 0; i < c; i++ {
		result += s
	}
	return result
}

func prettyPrint(input interface{}) (string, error) {
	buf := bytes.NewBuffer(nil)
	_, err := fmt.Fprintf(buf, "%+v", input)
	if err != nil {
		return "", err
	}
	inputString := buf.String()
	result := ""
	indent := 0
	for _, c := range inputString {
		character := string(c)
		if character == "{" || character == "[" {
			thisIndent := repeat("  ", indent)
			indent++
			nextIndent := repeat("  ", indent)
			result += fmt.Sprintf("\n%s%s\n%s", thisIndent, character, nextIndent)
		} else if character == " " {
			result += fmt.Sprintf("\n%s", repeat("  ", indent))
		} else if character == "}" || character == "]" {
			indent--
			nextIndent := repeat("  ", indent)
			result += fmt.Sprintf("\n%s%s%s", nextIndent, character, nextIndent)
		} else {
			result += character
		}
	}
	return result, nil
}
