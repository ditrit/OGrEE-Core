package main

import (
	"bytes"
	"cli/com"
	cmd "cli/controllers"
	l "cli/logger"
	"cli/models"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

var Output *bytes.Buffer

func Printf(format string, a ...any) {
	fmt.Fprintf(Output, format, a...)
}

func Println(a ...any) {
	fmt.Fprintln(Output, a...)
}

func main() {
	cmd.Printf = Printf
	cmd.Println = Println
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		fmt.Println(err.Error())
	}
	Output = &bytes.Buffer{}
	server := &cliServer{}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	com.RegisterCLIServer(grpcServer, server)
	grpcServer.Serve(lis)
}

func SetPrompt(user string) string {
	cmd.State.Prompt = "\u001b[1m\u001b[32m" + user + "@" + cmd.State.Customer + ":" +
		"\u001b[37;1m" + cmd.State.CurrPath + "\u001b[1m\u001b[32m>\u001b[0m "
	cmd.State.BlankPrompt = user + "@" + cmd.State.Customer + cmd.State.CurrPath + "> "
	return cmd.State.Prompt
}

func initCLI(conf *com.Config) error {
	l.InitLogs()
	cmd.InitConfigFilePath(conf.ConfigPath)
	cmd.InitHistoryFilePath(conf.HistPath)
	cmd.InitDebugLevel(conf.Verbose)
	cmd.InitTimeout(conf.UnityTimeout)
	cmd.InitURLs(conf.APIURL, conf.UnityURL)
	err := InitVars(conf.Variables)
	if err != nil {
		return fmt.Errorf("Error while initializing variables : %s", err.Error())
	}
	return nil
}

func connectCLI(conf *com.Config, username string, password string) error {
	var apiKey string
	user, apiKey, err := cmd.Login(username, password)
	if err != nil {
		return err
	} else {
		Printf("Successfully connected to %s\n", cmd.State.APIURL)
	}
	cmd.State.User = *user
	cmd.InitKey(apiKey)
	return cmd.InitState(conf)
}

type cliServer struct {
	com.UnimplementedCLIServer
	conf *com.Config
}

func (s *cliServer) Init(ctx context.Context, conf *com.Config) (*com.InitResponse, error) {
	Output.Reset()
	err := initCLI(conf)
	if err != nil {
		return nil, err
	}
	if !cmd.PingAPI() {
		return nil, fmt.Errorf("cannot reach API at %s", cmd.State.APIURL)
	}
	response := &com.InitResponse{
		Message: Output.String(),
	}
	s.conf = conf
	return response, nil
}

func (s *cliServer) ConnectAPI(ctx context.Context, r *com.ConnectAPIRequest) (*com.ProcessResponse, error) {
	Output.Reset()
	err := connectCLI(s.conf, r.Username, r.Password)
	if err != nil {
		return nil, err
	}
	userShort := strings.Split(cmd.State.User.Email, "@")[0]
	response := &com.ProcessResponse{
		Message: Output.String(),
		Prompt:  SetPrompt(userShort),
	}
	return response, nil
}

func (s *cliServer) ProcessCommand(ctx context.Context, r *com.ProcessCommandRequest) (*com.ProcessResponse, error) {
	Output.Reset()
	root, parseErr := Parse(r.Command)
	if parseErr != nil {
		Println(parseErr.Error())
	}
	if root != nil {
		_, err := root.execute()
		if err != nil {
			manageError(err, true)
		}
	}
	userShort := strings.Split(cmd.State.User.Email, "@")[0]
	response := &com.ProcessResponse{
		Message: Output.String(),
		Prompt:  SetPrompt(userShort),
	}
	return response, nil
}

func (s *cliServer) ProcessFile(r *com.ProcessFileRequest, stream com.CLI_ProcessFileServer) error {
	Output.Reset()
	if strings.Contains(r.FilePath, ".ocli") {
		LoadFile(r.FilePath)
		os.Exit(0)
	}
	return nil
}

func manageError(err error, addErrorPrefix bool) {
	l.GetErrorLogger().Println(err.Error())
	if cmd.State.DebugLvl > cmd.NONE {
		if traceErr, ok := err.(*stackTraceError); ok {
			Println(traceErr.Error())
		} else if errWithInternalErr, ok := err.(cmd.ErrorWithInternalError); ok {
			printError(errWithInternalErr.UserError, addErrorPrefix)
			if cmd.State.DebugLvl > cmd.ERROR {
				Println(errWithInternalErr.InternalError.Error())
			}
		} else {
			printError(err, addErrorPrefix)
		}
	}
}

func printError(err error, addErrorPrefix bool) {
	errMsg := err.Error()
	if !addErrorPrefix || strings.Contains(strings.ToLower(errMsg), "error") {
		Println(errMsg)
	} else {
		Println("Error:", errMsg)
	}
}

func (s *cliServer) Completion(ctx context.Context, r *com.CompletionRequest) (*com.CompletionResponse, error) {
	completionResponse := &com.CompletionResponse{
		Completions: Complete(r.Command),
	}
	return completionResponse, nil
}
func (s *cliServer) UnityStream(e *empty.Empty, stream com.CLI_UnityStreamServer) error {
	err := cmd.Connect3D(cmd.State.Ogree3DURL)
	if err == nil {
		unityMsg := com.UnityMessage{
			NewConnection: true,
			Message:       "",
		}
		stream.Send(&unityMsg)
	}
	for message := range models.Ogree3D.MessageChan() {
		unityMsg := com.UnityMessage{
			NewConnection: false,
			Message:       message,
		}
		stream.Send(&unityMsg)
	}
	if err != nil {
		manageError(err, false)
	}
	return nil
}
