from __future__ import print_function

import asyncio
import json
import os
import re
import sys

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
    return result[1], result[2]


class PluginInfo():
    def __init__(self, name, version):
        self.name = name
        self.version = version

    def json(self):
        return json.dumps(self, default=lambda o: o.__dict__)


class Plugin(object):
    def __init__(self, env=None, version=""):
        if env == None:
            env = os.environ

        self.connect_addr = json.loads(env.get(ENV_PLUGIN_CONNECT_ADDRESS))
        self.protocols = env.get(ENV_SUPPORT_RPC_PROTOCOL)
        self.install_version = env.get(ENV_INSTALL_VERSION)

        if self.connect_addr == None or self.protocols == None or self.install_version == None:
            raise NotImplementedError('You must run this program as plugin.')
        if self.install_version != version:
            raise RuntimeError('version not support')

        self.protocol, addr = self.config(self.protocols, self.connect_addr)
        self.schema, self.addr = parse_address(addr)

    @asyncio.coroutine
    def _run(self):
        if self.protocol == 'JsonRPCProtocol':
            server = Server('ws://' + self.addr + '/plugin')
            server.Plugin.Ping = Plugin._pong.__get__(self, Plugin)
            server.Plugin.GetPluginInfo = self._get_plugin_info
            server.Plugin.Command = Plugin._command.__get__(self, Plugin)
            server.Plugin.GetHelp = Plugin._get_help.__get__(self, Plugin)
            server.Plugin.ListCommand = Plugin._list_command.__get__(self, Plugin)
            try:
                yield from server.ws_connect()
                yield from asyncio.sleep(3)
            finally:
                yield from server.close()
                yield from server.session.close()

    def _pong(self, *args, **kwargs):
        # print(args)
        # print(kwargs)
        return 'pong'

    def _command(self, *args, **kwargs):
        #print(args)
        #print(kwargs)
        self.get_plugin_info()


    def _get_help(self, *args, **kwargs):
        # print(args)
        # print(kwargs)
        self.get_plugin_info()


    def _list_command(self, *args, **kwargs):
        # print(args)
        # print(kwargs)
        self.get_plugin_info()


    def _get_plugin_info(self, *args, **kwargs):
        print(args)
        print(kwargs)
        return self.get_plugin_info().json()

    def run(self):
        asyncio.get_event_loop().run_until_complete(self._run())

    def stop(self, code):
        pass

    def config(self, protocols, connect_addr):
        raise NotImplementedError('Method not implemented!')

    def command(self, name):
        raise NotImplementedError('Method not implemented!')

    def get_help(self, name):
        raise NotImplementedError('Method not implemented!')

    def list_command(self):
        raise NotImplementedError('Method not implemented!')

    def get_plugin_info(self):
        raise NotImplementedError('Method not implemented!')
