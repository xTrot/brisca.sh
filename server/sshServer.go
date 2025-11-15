package server

import (
	"context"
	"errors"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"brisca.sh/server/screens"
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
	Env             Environment
	keyPath         = ".ssh/id_ed25519"
)

type Environment struct {
	Host  string `default:"localhost"`
	Port  string `default:"23234"`
	Log   string `default:"brisca.log"`
	Debug bool   `default:"false"`
	Key   string `default:""`

	BrowserServer string `default:"http://browser:9000"`
	GameServer    string `default:"http://games:8000"`
}

func Start() {
	os.Setenv("GLAMOUR_STYLE", "dracula")
	err := envconfig.Process("brisca", &Env)
	if err != nil {
		log.Fatal(err.Error())
		panic("defaults loading failed.")
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(Env.Host, Env.Port)),
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

		//
	)

	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	if Env.Debug {
		log.SetLevel(log.DebugLevel)
		log.Helper()
		log.SetReportCaller(true)
		log.Debug("Debug Started")
	}

	log.Debug("Env:", "env", Env)

	s.IdleTimeout = 15 * time.Minute
	log.Info("Server", "s.IdleTimeout", s.IdleTimeout)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", Env.Host, "port", Env.Port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	sig := <-done
	log.Info("Initiating graceful shutdown, Received signal: ", "sig", sig)
	log.Info("Performing cleanup operations...")

	screens.Retired = true
	log.Info("Retired set to true.")

	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Minute)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}

	log.Info("Application shut down gracefully.")

	if log.GetLevel() == log.DebugLevel {

		log.Debug("Let me read the logs before shutting down.")
		time.Sleep(5 * time.Second)

	}

	//
}

func skipThis() ssh.KeyboardInteractiveHandler {
	return func(ctx ssh.Context, challenger gossh.KeyboardInteractiveChallenge) bool {
		return true
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {

	quit, _ := screens.HasRetired()
	if quit != nil {
		return quit, []tea.ProgramOption{}
	}

	m := screens.NewRegisterScreen(&s, Env.BrowserServer, Env.GameServer)
	return m, []tea.ProgramOption{tea.WithAltScreen(), tea.WithoutSignalHandler()}
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
