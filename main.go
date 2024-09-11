package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/logging"
	"github.com/spf13/cobra"
)

func getEnv(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	return value
}

var sshHost = getEnv("SSH_HOST", "localhost")
var sshPort = getEnv("SSH_PORT", "23235")
var sshPubKeyFile = getEnv("SSH_PUBKEY_FILE", ".pubkeys")

func getPubKeysFromFile() []string {
	dat, err := os.ReadFile(sshPubKeyFile)
	if os.IsNotExist(err) {
		println("ERROR: SSH pubkey file not found! Nobody will be able to adjust the counter!")
		return []string{}
	}
	if err != nil {
		panic(err)
	}
	return strings.Split(string(dat), "\n")
}

const (
	counterMax = 100
)

func getCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "get",
		Short: "Get the current counter",
		Long:  `Get the current counter`,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			config, err := LoadConfig()
			if err != nil {
				panic(err)
			}
			cmd.Println(config.Counter)
		},
	}
	return cmd
}

func incrCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "incr",
		Short: "Increment the counter",
		Long:  `Increment the counter`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			incr, err := strconv.Atoi(args[0])
			if err != nil {
				// ... handle error
				panic(err)
			}
			oldVal := getCounter()
			newVal := oldVal + incr
			setCounter(newVal)
			cmd.Println("Incremented the counter by " + strconv.Itoa(incr) + " to " + strconv.Itoa(newVal))
		},
	}
	return cmd
}

func setCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "set [new counter value]",
		Short: "Set the counter",
		Long:  `Set the counter`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			newVal, err := strconv.Atoi(args[0])
			if err != nil {
				// ... handle error
				panic(err)
			}
			oldCounterVal := getCounter()
			setCounter(newVal)
			cmd.Println("Set the counter from " + strconv.Itoa(oldCounterVal) + " to " + strconv.Itoa(newVal))
		},
	}
	return cmd
}

func getCounter() int {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	return config.Counter
}

func setCounter(newVal int) {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	config.Counter = newVal
	err = SaveConfig(config)
	if err != nil {
		panic(err)
	}
}

func main() {
	pubKeys := getPubKeysFromFile()

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(sshHost, sshPort)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithPublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
			return key.Type() == "ssh-ed25519"
		}),
		wish.WithMiddleware(
			func(next ssh.Handler) ssh.Handler {
				return func(sess ssh.Session) {
					var rootCmd = &cobra.Command{}
					rootCmd.SetArgs(sess.Command())
					rootCmd.SetIn(sess)
					rootCmd.SetOut(sess)
					rootCmd.SetErr(sess.Stderr())

					for _, pubkey := range pubKeys {
						parsed, _, _, _, _ := ssh.ParseAuthorizedKey(
							[]byte(pubkey),
						)
						if ssh.KeysEqual(sess.PublicKey(), parsed) {
							// register the commands which are only available with authenciation
							rootCmd.AddCommand(incrCmd())
							rootCmd.AddCommand(setCmd())
						}
					}

					// register the commands which are available with and without authentication
					rootCmd.AddCommand(getCmd())

					rootCmd.CompletionOptions.DisableDefaultCmd = true

					rootCmd.Execute()
					/*if err := rootCmd.Execute(); err != nil {
						_ = sess.Exit(1)
						return
					}*/

					next(sess)
				}
			},
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", sshHost, "port", sshPort)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
