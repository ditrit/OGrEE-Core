package main

import (
	"bytes"
	"cli/config"
	cmd "cli/controllers"
	l "cli/logger"
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
	server := cliServer{}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	RegisterCLIServer(grpcServer, &server)
	grpcServer.Serve(lis)
}

func SetPrompt(user string) string {
	cmd.State.Prompt = "\u001b[1m\u001b[32m" + user + "@" + cmd.State.Customer + ":" +
		"\u001b[37;1m" + cmd.State.CurrPath + "\u001b[1m\u001b[32m>\u001b[0m "
	cmd.State.BlankPrompt = user + "@" + cmd.State.Customer + cmd.State.CurrPath + "> "
	return cmd.State.Prompt
}

func InterpretLine(str string) {
	root, parseErr := Parse(str)
	if parseErr != nil {
		Println(parseErr.Error())
		return
	}
	if root == nil {
		return
	}
	_, err := root.execute()
	if err != nil {
		manageError(err, true)
	}
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

type cliServer struct{}

func (s *cliServer) Init(ctx context.Context, e *empty.Empty) (*ProcessResponse, error) {
	Output.Reset()
	conf := config.ReadConfig()

	l.InitLogs()
	cmd.InitConfigFilePath(conf.ConfigPath)
	cmd.InitHistoryFilePath(conf.HistPath)
	cmd.InitDebugLevel(conf.Verbose)
	cmd.InitTimeout(conf.UnityTimeout)
	cmd.InitURLs(conf.APIURL, conf.UnityURL)

	if !cmd.PingAPI() {
		Println("Cannot reach API at", cmd.State.APIURL)
	}
	var apiKey string
	user, apiKey, err := cmd.Login(conf.User, conf.Password)
	if err != nil {
		Println(err.Error())
	} else {
		Printf("Successfully connected to %s\n", cmd.State.APIURL)
	}
	cmd.State.User = *user
	cmd.InitKey(apiKey)
	err = cmd.InitState(conf)
	if err != nil {
		Println(err.Error())
	}
	err = InitVars(conf.Variables)
	if err != nil {
		Println("Error while initializing variables :", err.Error())
	}
	//Execute Script if provided as arg and exit
	if conf.Script != "" {
		if strings.Contains(conf.Script, ".ocli") {
			LoadFile(conf.Script)
			os.Exit(0)
		}
	}
	userShort := strings.Split(cmd.State.User.Email, "@")[0]
	response := &ProcessResponse{
		Message: Output.String(),
		Prompt:  SetPrompt(userShort),
	}
	return response, nil
}

func (s *cliServer) ProcessLine(ctx context.Context, r *Request) (*ProcessResponse, error) {
	Output.Reset()
	InterpretLine(r.Line)
	userShort := strings.Split(cmd.State.User.Email, "@")[0]
	response := &ProcessResponse{
		Message: Output.String(),
		Prompt:  SetPrompt(userShort),
	}
	return response, nil
}

func (s *cliServer) Completion(ctx context.Context, r *Request) (*CompletionResponse, error) {
	completionResponse := &CompletionResponse{
		Completions: Complete(r.Line),
	}
	return completionResponse, nil
}
func (s *cliServer) UnityStream(e *empty.Empty, stream CLI_UnityStreamServer) error {
	err := cmd.Connect3D(cmd.State.Ogree3DURL)
	if err == nil {
		unityMsg := UnityMessage{
			NewConnection: true,
			Message:       "",
		}
		stream.Send(&unityMsg)
	}
	for message := range models.Ogree3D.MessageChan() {
		unityMsg := UnityMessage{
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
func (s *cliServer) mustEmbedUnimplementedCLIServer() {
}
