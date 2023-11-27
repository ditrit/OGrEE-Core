import argparse
import os
import sys
import toml
import com_pb2 as pb2
from icecream import ic

def exe_dir():
    if getattr(sys, 'frozen', False):  # Check if the application is frozen (e.g., PyInstaller)
        return os.path.dirname(sys.executable)
    elif __file__:
        return os.path.dirname(__file__)
    

class Vardef:
    def __init__(self, name, value):
        self.Name = name
        self.Value = value

class Config:
    def __init__(self):
        self.Verbose = "ERROR"
        self.APIURL = ""
        self.UnityURL = ""
        self.UnityTimeout = "10ms"
        self.ConfigPath = exe_dir() + "/../config.toml"
        self.HistPath = "./.history"
        self.Script = ""
        self.Drawable = ["all"]
        self.DrawableJson = {}
        self.DrawLimit = 50
        self.Updates = ["all"]
        self.User = ""
        self.Password = ""

def marshal_to_protobuf(config_instance):
    config_proto = pb2.Config()
    config_proto.Verbose = config_instance.Verbose
    config_proto.APIURL = config_instance.APIURL
    config_proto.UnityURL = config_instance.UnityURL
    config_proto.UnityTimeout = config_instance.UnityTimeout
    config_proto.ConfigPath = config_instance.ConfigPath
    config_proto.HistPath = config_instance.HistPath
    config_proto.Script = config_instance.Script
    config_proto.DrawLimit = config_instance.DrawLimit
    config_proto.User = config_instance.User
    config_proto.Password = config_instance.Password
    config_proto.Drawable.extend(config_instance.Drawable)
    config_proto.Updates.extend(config_instance.Updates)
    return config_proto

class ArgumentParser(argparse.ArgumentParser):
    def error(self, message):
        self.print_help(sys.stderr)
        self.exit(2, '%s: error: %s\n' % (self.prog, message))

def read_config():
    conf = Config()
    parser = ArgumentParser()
    parser.add_argument('--conf_path', '-c', default=conf.ConfigPath, dest='ConfigPath',
                        help='Indicate the location of the Shell\'s config file')
    parser.add_argument('--verbose', '-v', dest='Verbose',
                        help='Indicates level of debugging messages. The levels are of in ascending order: {NONE,ERROR,WARNING,INFO,DEBUG}.')
    parser.add_argument('--unity_url', '-u', dest='UnityURL',
                        help='Unity URL')
    parser.add_argument('--api_url', '-a', dest='APIURL',
                        help='API URL')
    parser.add_argument('--history_path', dest='HistPath',
                        help='Indicate the location of the Shell\'s history file')
    parser.add_argument('--file', '-f', dest='Script',
                        help='Launch the shell as an interpreter by only executing an OCLI script file')
    parser.add_argument('--user', dest='User',
                        help='User email')
    parser.add_argument('--password', dest='Password',
                        help='Password')

    args = parser.parse_args()

    try:
        with open(args.ConfigPath, 'r') as f:
            toml_data = toml.load(f).get('OGrEE-CLI', {})
            conf.__dict__.update(toml_data)
    except FileNotFoundError as e:
        raise FileNotFoundError(f"cannot read config file {args.conf_path}: {e}\n"
                                f"Please ensure that you have a properly formatted config file saved as 'config.toml' in the parent directory\n"
                                f"For more details please refer to: https://github.com/ditrit/OGrEE-Core/blob/main/README.md")
    
    for arg, value in vars(args).items():
        if value is not None:
            setattr(conf, arg, value)
    
    conf.ConfigPath = os.path.abspath(conf.ConfigPath)
    conf.HistPath = os.path.abspath(conf.HistPath)

    return marshal_to_protobuf(conf)