package src

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/adrg/xdg"
	"github.com/gorilla/handlers"
	"github.com/spf13/afero"
	"github.com/xescugc/rubberducking/assets"
	"github.com/xescugc/rubberducking/log"
	"go.uber.org/atomic"
	"golang.design/x/hotkey"
	"golang.design/x/hotkey/mainthread"
)

const (
	AppName = "rubberducking"
)

var (
	// This 2 amoics are only available on the main run, not on the subcommands
	isGameRunning atomic.Bool
	gamePort      atomic.String

	dataFile = path.Join(xdg.DataHome, AppName, "data.json")

	gameStarted     = make(chan struct{})
	shouldGameStart = make(chan struct{})
)

func initFs(fs afero.Fs) error {
	err := fs.MkdirAll(filepath.Dir(dataFile), 0700)
	if err != nil {
		return fmt.Errorf("failed to MkdirAll: %w", err)
	}
	return nil
}

type Data struct {
	ManagerURL string `json:"manager_url"`
}

type CreateMessageRequest struct {
	Message string `json:"message"`
}
type ErrorResponse struct {
	Error string `json:"error"`
}

func Manager(ctx context.Context, fs afero.Fs, v bool) error {
	err := initFs(fs)
	if err != nil {
		return fmt.Errorf("could not initialize FS")
	}

	go mainthread.Init(hotkeysFn(ctx, fs))

	tmpdir, err := afero.TempDir(fs, "", AppName)
	if err != nil {
		return fmt.Errorf("failed to initialize temp dir: %w", err)
	}
	err = fs.MkdirAll(tmpdir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create a temp dir: %w", err)
	}

	ctx, cfn := context.WithCancel(ctx)
	cleanup := func() {
		fs.RemoveAll(tmpdir)
		fs.Remove(dataFile)
		cfn()
		select {
		case <-gameStarted:
		default:
			close(gameStarted)
		}
	}
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c

		// signal caught, cleanup
		cleanup()
	}()
	// Just in case it panics then we also have to clean stuff
	defer cleanup()

	gamepath := filepath.Join(tmpdir, "game")

	err = afero.WriteFile(fs, gamepath, assets.Game, 7777)
	if err != nil {
		return fmt.Errorf("failed to write Game: %w", err)
	}

	sp, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get a free port for the server: %w", err)
	}

	ssp := strconv.Itoa(sp)
	data := Data{ManagerURL: "http://localhost:" + ssp}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = afero.WriteFile(fs, dataFile, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", dataFile, err)
	}

	go startHandler(ssp, v)

	err = startGame(ctx, data.ManagerURL, gamepath, v)
	if err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	return nil
}

func buildProxyRequest(req *http.Request) *http.Request {
	proxyReq := req.Clone(context.Background())

	// create a new url from the raw RequestURI sent by the client
	u := "http://localhost:" + gamePort.Load() + "/messages"

	proxyReq.URL, _ = url.Parse(u)
	proxyReq.RequestURI = ""

	// We may want to filter some headers, otherwise we could just use a shallow copy
	// proxyReq.Header = req.Header
	proxyReq.Header = make(http.Header)
	for h, val := range req.Header {
		proxyReq.Header[h] = val
	}
	return proxyReq
}

func startHandler(sp string, v bool) {
	hmux := http.NewServeMux()

	hmux.HandleFunc("/game/start", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			isGameRunning.Store(true)
			log.Logger.Info("Game Started signal send")
			gameStarted <- struct{}{}
			log.Logger.Info("Game Started signal read")
		}
		return
	})
	hmux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// This handler basically proxies any request to the Game if it's running
		// if it's not running then it waits for it to run before proxing the request
		if !isGameRunning.Load() {
			log.Logger.Info("Game is not running, pushing to channel 'shouldGameStart'")
			shouldGameStart <- struct{}{}

			log.Logger.Info("Waiting for the Game to start on 'gameStarted'")
			<-gameStarted
		}

		log.Logger.Info("Game is running", "isGameRunning", isGameRunning.Load())
		proxyReq := buildProxyRequest(req)
		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			retryTime := time.Second / 30
			log.Logger.Info("Retrying HTTP request")
			time.Sleep(retryTime)

			proxyReq = buildProxyRequest(req)
			resp, err = http.DefaultClient.Do(proxyReq)
			if err != nil {
				log.Logger.Error("HTTP request failed", "error", err, "url", proxyReq.URL.String())
				em := ErrorResponse{Error: fmt.Errorf("Failed to connect to the game: %w", err).Error()}
				w.WriteHeader(http.StatusBadRequest)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(em)
			}
			return
		}
		defer resp.Body.Close()
	})

	out := io.Discard
	if v {
		out = os.Stdout
	}
	svr := &http.Server{
		Addr:    fmt.Sprintf(":%s", sp),
		Handler: handlers.LoggingHandler(out, hmux),
	}

	log.Logger.Info("Starting server", "port", sp)
	if err := svr.ListenAndServe(); err != nil {
		panic(fmt.Errorf("server error: %w", err).Error())
	}
}

func startGame(ctx context.Context, managerURL, gamePath string, v bool) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-shouldGameStart:
			log.Logger.Debug("Start game")
			port, err := getFreePort()
			if err != nil {
				return err
			}
			log.Logger.Info("Game port", "port", port)

			cmd := exec.CommandContext(ctx, gamePath)
			cmd.Env = append(os.Environ(), []string{
				"PORT=" + strconv.Itoa(port),
				fmt.Sprintf("VERBOSE=%t", v),
				"MANAGER_URL=" + managerURL,
			}...)

			stderr, err := cmd.StderrPipe()
			if err != nil {
				log.Logger.Error("Failed to connect to StderrPipe", "error", err)
				return fmt.Errorf("Failed to connect to StderrPipe: %w", err)
			}
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Logger.Error("Failed to connect to StdoutPipe", "error", err)
				return fmt.Errorf("Failed to connect to StdoutPipe: %w", err)
			}

			gamePort.Store(strconv.Itoa(port))
			log.Logger.Info("Starting to run game")
			err = cmd.Start()
			if err != nil {
				log.Logger.Info("Failed to start game")
				return fmt.Errorf("failed to run game")
			}

			isGameRunning.Store(true)

			errs, _ := io.ReadAll(stderr)
			outs, _ := io.ReadAll(stdout)

			log.Logger.Info("Waiting for the game to end")
			err = cmd.Wait()
			if err != nil {
				log.Logger.Error("Failed to Wait for cmd", "error", err)
			}

			isGameRunning.Store(false)

			log.Logger.Info("Game ended")
			if v {
				log.Logger.Info("Game output", "CMD", cmd.String(), "STDOUT", outs, "STDERR", errs)
			}
		}
	}
}

func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func hotkeysFn(ctx context.Context, fs afero.Fs) func() {
	return func() {
		hk := hotkey.New([]hotkey.Modifier{hotkey.Mod4, hotkey.ModShift}, hotkey.KeyD)
		err := hk.Register()
		if err != nil {
			log.Logger.Error("hotkey: failed to register hotkey", "error", err)
			return
		}
		for {
			select {
			case <-ctx.Done():
				hk.Unregister()
				return
			case <-hk.Keydown():
				log.Logger.Info("hotkey: Keydown", "key", hk)
			case <-hk.Keyup():
				log.Logger.Info("hotkey: Keyup", "key", hk)
				if !isGameRunning.Load() {
					err = SendMessage(ctx, fs, "Quack!")
				} else {
					err = SendMessage(ctx, fs, "")
				}
				if err != nil {
					log.Logger.Error("hotkey: failed to send message", "error", err)
					return
				}
			}
		}
	}
}
