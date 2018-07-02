from __future__ import print_function

import os
import sys
import asyncio
import json
import re
from jsonrpc_websocket import Server

ENV_PLUGIN_CONNECT_ADDRESS = "ENV_PLUGIN_CONNECT_ADDRESS"
ENV_SUPPORT_RPC_PROTOCOL = "ENV_SUPPORT_RPC_PROTOCOL"
ENV_INSTALL_VERSION = "ENV_INSTALLED_VERSION"

GRPCProtocol = 1

if sys.version_info[:2] < (3, 3):
    old_print = print


    def print(*args, **kwargs):
        flush = kwargs.pop('flush', False)
        old_print(*args, **kwargs)
        if flush:
            file = kwargs.get('file', sys.stdout)
            # Why might file=None? IDK, but it works for print(i, file=None)
            file.flush() if file is not None else sys.stdout.flush()

def parse_address(addr):
    prog = re.compile('(.*)://(.*)')
    result = prog.match(addr)
    return result[0], result[1]

def client_method(arg1, arg2):
    return arg1 + arg2

class Plugin(object):
    def __init__(self, env=None, version=""):
        if env == None:
            env = os.environ

        self.connect_addr = json.loads(env.get(ENV_PLUGIN_CONNECT_ADDRESS))
        self.rpc_protocols = env.get(ENV_SUPPORT_RPC_PROTOCOL)
        self.install_version = env.get(ENV_INSTALL_VERSION)

        if self.connect_addr == None or self.rpc_protocols == None or self.install_version == None:
            raise NotImplementedError('You must run this program as plugin.')
        if self.install_version != version:
            raise RuntimeError('version not support')

        self.protocol, addr = self.config()
        self.schema, self.addr = parse_address(addr)

    @asyncio.coroutine
    def run(self):
        if self.protocol == 'JsonRPCProtocol':
            server = Server('ws://' + self.addr)
            server.Plugin.Command = client_method
            server.Plugin.GetHelp = client_method
            server.Plugin.ListCommand = client_method
            try:
                yield from server.ws_connect()
            finally:
                yield from server.close()
                yield from server.session.close()

    def stop(self, code):
        pass

    def config(self):
        raise NotImplementedError('Method not implemented!')

    def command(self, name):
        raise NotImplementedError('Method not implemented!')

    def get_help(self, name):
        raise NotImplementedError('Method not implemented!')

    def list_command(self):
        raise NotImplementedError('Method not implemented!')
