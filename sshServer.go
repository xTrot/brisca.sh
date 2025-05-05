package main

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gossh "golang.org/x/crypto/ssh"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/kelseyhightower/envconfig"
)

var (
	db = map[string]string{
		"Enddy": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCa4MMeQKYeAyEekyGOA0WlD/vFvlxQXu/yZV81wMWEKeplLTvMQGjfpsgA51BvmLAAzFhBnPloAFS6+dUvKlYJ27HbdI/6hLAxUPH19OwYD9Aks0utGIXqPRwPl+TVCM+4OZkvsd18rKueWJ/SYBeVudNzXECx9UmB2n33Rz4OLC+tKLuoYVgAd3LGIHsRS29o67OpPwHdW8zosCfQ6ZLD4oAinHqSIZCqfXXtrfee1F5hQNTxPTU47zyVNUshm9JaqrYQyFn5AWKjolcqOr3zt176rULsphZYcha9XpZ6u2M0YeJEkLUIrKAVoY3aTG0ZqBKMryLA4G89L+AQCcX1lxvMnU1SOotQ57C/CDC+iiqWF1VguU/23H80LVANYEenJYqgPhN3A42d7HchcaW8VTAwjLrrBSPT9F336oi+jQNGTPWfNndp9i0dlbPZbSRW89hOX+N5EHodWZPf09Wb5pvNm2Hyd2GCIoIhmF+kMyJnP3X6LX9ebn5jKK9mxiqsdjvZKlQLlhaLGHZmMJKyqWzOPKcBFPzU+h9/pUrqcmNQCP8djM8Al7z/JLeUqP9TIXv7W/yyVvMcAA/TgR/GvcXwV+JQFqi4x2EwaD+VMQeSBvGu01f0CVPZFepmDLBpSztzr0adk6f49aCNwXY+52s/Z3VxTjNblStqLFYAFw== enddyygf93@live.com",
	}
	allowedKeyTypes = "ssh-rsa, "
	env             Environment
	keyPath         = ".ssh/id_ed25519"
)

type Environment struct {
	Host   string `default:"localhost"`
	Server string `default:"localhost"`
	Port   string `default:"23234"`
	Log    string `default:"brisca.log"`
	Debug  bool   `default:"false"`
	Key    string `default:""`
}

func main() {
	os.Setenv("GLAMOUR_STYLE", "dracula")
	err := envconfig.Process("brisca", &env)
	if err != nil {
		log.Fatal(err.Error())
		panic("defaults loading failed.")
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(env.Host, env.Port)),
		wish.WithHostKeyPath(keyPath),
		wish.WithPublicKeyAuth(keyHandler),           // This should be optional but isn't.
		wish.WithKeyboardInteractiveAuth(skipThis()), // If this isn't added the PubKeyAuth will require a key.
		// This makes make PubKey Auth optional
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			AuthMiddleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	if env.Debug {
		log.SetLevel(log.DebugLevel)
		log.Helper()
		log.SetReportCaller(true)
		log.Debug("Debug Started")
	}

	log.Debug("Env:", "env", env)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", env.Host, "port", env.Port)
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

func skipThis() ssh.KeyboardInteractiveHandler {
	return func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
		return true
	}
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	// renderer := bubbletea.MakeRenderer(s)

	m := newModel(&s)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func keyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	if !strings.Contains(allowedKeyTypes, key.Type()) {
		allowedKeyTypes += key.Type() + ", "
		log.Info("newKeyTypeAdded:", "allowedKeyTypes", allowedKeyTypes)
	}
	return true
}

func AuthMiddleware() wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			keyUserGave := sess.PublicKey()

			if keyUserGave == nil {
				log.Info("AuthMiddleware: No key provided.")
				next(sess)
				return
			}

			if !strings.Contains(allowedKeyTypes, keyUserGave.Type()) {
				allowedKeyTypes += keyUserGave.Type() + ", "
				log.Info("newKeyTypeAdded:", "allowedKeyTypes", allowedKeyTypes)
			}

			var found bool
			for name, pubkey := range db {
				keyStored, _, _, _, _ := ssh.ParseAuthorizedKey([]byte(pubkey))
				if ssh.KeysEqual(keyUserGave, keyStored) {
					log.Info("AuthMiddleWare: I remember,", "name", name)
					found = true
				}
			}

			if !found {
				log.Info("AuthMiddleware: I don't remember, I can offer to remember you!", "name", sess.User())
			}

			next(sess)
		}
	}
}
