#!/usr/bin/env python
import time
start = time.time()
from typing import Iterable
from termcolor import colored

end = time.time()
print(end-start)

from prompt_toolkit.completion import Completer, Completion
from prompt_toolkit.completion.base import Completion
from prompt_toolkit.formatted_text import ANSI
from prompt_toolkit import PromptSession
from prompt_toolkit.history import FileHistory
from prompt_toolkit.patch_stdout import patch_stdout

end = time.time()
print(end-start)
import grpc
end = time.time()
print(end-start)
import com_pb2_grpc as pb2_grpc
import com_pb2 as pb2
from google.protobuf.empty_pb2 import Empty



import asyncio
import sys
import nest_asyncio
nest_asyncio.apply()


async def get_completions_async(stub, line):
    response = await stub.Completion(pb2.Request(line=line))
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
            completer=completer, complete_while_typing=False
        )
        response = await stub.ProcessLine(pb2.Request(line=text))
        print(response.message, end='')
        return response.prompt
    except KeyboardInterrupt:
        return prompt


async def run():
    our_history = FileHistory(".history")
    session = PromptSession(history=our_history)
    loop = asyncio.get_event_loop()
    try:
        addr = 'localhost:50051'
        async with grpc.aio.insecure_channel(addr) as channel:
            stub = pb2_grpc.CLIStub(channel)
            completer = CustomCompleter(stub, loop)
            loop.create_task(unity_loop(stub))
            response = await stub.ProcessLine(pb2.Request(line=""))
            prompt = response.prompt
            end = time.time()
            print(end-start)
            while True:
                prompt = await prompt_iter(session, stub, completer, prompt)
    except (EOFError, grpc.aio._call.AioRpcError):
        return


def handle_exception(exc_type, exc_value, exc_traceback):
    sys.__excepthook__(exc_type, exc_value, exc_traceback)


if __name__ == "__main__":
    sys.excepthook = handle_exception
    with patch_stdout(raw=True):
        asyncio.run(run())
