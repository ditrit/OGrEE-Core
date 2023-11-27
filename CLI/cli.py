#!/usr/bin/env python
from prompt_toolkit.completion import Completer, Completion
from prompt_toolkit.completion.base import Completion
from prompt_toolkit.formatted_text import ANSI
from prompt_toolkit import PromptSession
from prompt_toolkit.history import FileHistory
from prompt_toolkit.patch_stdout import patch_stdout
from prompt_toolkit.key_binding import KeyBindings
import grpc
import grpc._channel
import com_pb2_grpc as pb2_grpc
import com_pb2 as pb2
from google.protobuf.empty_pb2 import Empty
import asyncio
import sys
import subprocess
import os
import psutil
import nest_asyncio
from config import read_config
from icecream import ic

nest_asyncio.apply()

def check_process_running(process_name):
    for proc in psutil.process_iter(['pid', 'name']):
        if os.path.splitext(proc.info['name'])[0].lower() == process_name.lower():
            return True
    return False

async def get_completions_async(stub, command):
    response = await stub.Completion(pb2.CompletionRequest(command=command))
    return response.completions

class CustomCompleter(Completer):
    def __init__(self, stub, event_loop):
        self.stub = stub
        self.event_loop = event_loop  

    def get_completions(self, document, complete_event):
        task = self.event_loop.create_task(get_completions_async(self.stub, document.text))
        completions = self.event_loop.run_until_complete(task)
        return [Completion(text, 0) for text in completions]


async def unity_loop(stub):
    try:
        stream = stub.UnityStream(Empty())
        async for message in stream:
            if message.newConnection:
                print("Established connection with OGrEE-3D!")
            else:
                print("Received from OGrEE-3D:", message.message)
    except grpc.aio._call.AioRpcError:
        return


async def prompt_iter(session, stub, completer, prompt):
    try:
        text = await session.prompt_async(
            ANSI(prompt),
            completer=completer, complete_while_typing=False, is_password=False,
        )
        response = await stub.ProcessCommand(pb2.ProcessCommandRequest(command=text))
        print(response.message, end='')
        return response.prompt
    except KeyboardInterrupt:
        return prompt
    

async def run(conf):
    our_history = FileHistory(conf.HistPath)
    session = PromptSession(enable_history_search=False, history=our_history)
    loop = asyncio.get_event_loop()
    try:
        async with grpc.aio.insecure_channel(addr) as channel:
            stub = pb2_grpc.CLIStub(channel)

            try:
                init_response = await stub.Init(conf)
                print(init_response.message, end='')
            except grpc.RpcError as e:
                print(e.details())
                return

            login_session = PromptSession()
            if conf.User == "":
                 conf.User = login_session.prompt("Login : ", is_password=False)
            password_session = PromptSession()
            if conf.Password == "":
                conf.Password = password_session.prompt("Password : ", is_password=True)

            try:
                response = await stub.ConnectAPI(pb2.ConnectAPIRequest(
                    username=conf.User, password = conf.Password))
                prompt = response.prompt
            except grpc.RpcError as e:
                print(e.details())
                return
            
            completer = CustomCompleter(stub, loop)
            loop.create_task(unity_loop(stub))
            
            while True:
                prompt = await prompt_iter(session, stub, completer, prompt)

    except (EOFError, grpc.aio._call.AioRpcError, KeyboardInterrupt):
        pass

if __name__ == "__main__":
    go_process = None
    addr = '127.0.0.1:50051'
    if not check_process_running("ogree-cli-service"):
        script_dir = os.path.dirname(os.path.abspath(__file__))
        go_process = subprocess.Popen([os.path.join(script_dir, 'ogree-cli-service')])

    conf = read_config()
    
    def handle_exception(exc_type, exc_value, exc_traceback):
        if go_process is not None:
            go_process.terminate()
        sys.__excepthook__(exc_type, exc_value, exc_traceback)
    
    sys.excepthook = handle_exception
    with patch_stdout(raw=True):
        asyncio.run(run(conf))

    if go_process is not None:
        go_process.terminate()
    
