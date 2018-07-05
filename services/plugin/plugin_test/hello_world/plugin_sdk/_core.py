from __future__ import print_function

import asyncio
import json
import os
import re
import sys
import logging

import aiohttp
from aiohttp import ClientError
from aiohttp.http_exceptions import HttpProcessingError
from jsonrpc import Dispatcher, JSONRPCResponseManager

logging.basicConfig(level=logging.DEBUG)

ENV_PLUGIN_CONNECT_ADDRESS = "ENV_PLUGIN_CONNECT_ADDRESS"
ENV_SUPPORT_RPC_PROTOCOL = "ENV_SUPPORT_RPC_PROTOCOL"
ENV_INSTALL_VERSION = "ENV_INSTALLED_VERSION"

if sys.version_info[:2] < (3, 3):
    old_print = print


    def print(*args, **kwargs):
        flush = kwargs.pop('flush', False)
        old_print(*args, **kwargs)
        if flush:
            file = kwargs.get('file', sys.stdout)
            # Why might file=None? IDK, but it works for print(i, file=None)
            file.flush() if file is not None else sys.stdout.flush()


class TransportError(Exception):

    def __init__(self, exception_text, message=None, *args):
        if message:
            super(TransportError, self).__init__(
                '%s: %s' % (message.transport_error_text, exception_text), *args)
        else:
            super(TransportError, self).__init__(exception_text, *args)


class ProtocolError(Exception):
    pass


def parse_address(addr):
    prog = re.compile('(.*)://(.*)')
    result = prog.match(addr)
    return result[1], result[2]


class PluginInfo:
    def __init__(self, name, version):
        self.name = name
        self.version = version


class Plugin(object):
    def __init__(self, env=None, version="", **connect_kwargs):
        if env is None:
            env = os.environ

        self.connect_addr = json.loads(env.get(ENV_PLUGIN_CONNECT_ADDRESS))
        self.protocols = env.get(ENV_SUPPORT_RPC_PROTOCOL)
        self.install_version = env.get(ENV_INSTALL_VERSION)

        if self.connect_addr is None or self.protocols is None or self.install_version is None:
            raise NotImplementedError('You must run this program as plugin.')
        if self.install_version != version:
            raise RuntimeError('version not support')

        self.protocol, addr = self.config(self.protocols, self.connect_addr)
        self.schema, self.addr = parse_address(addr)

        self._session = None
        self._client = None
        self._connect_kwargs = connect_kwargs
        self._connect_kwargs['headers'] = self._connect_kwargs.get('headers', {})
        self._connect_kwargs['headers']['Content-Type'] = \
            self._connect_kwargs['headers'].get('Content-Type', 'application/json')
        self._connect_kwargs['headers']['Accept'] = \
            self._connect_kwargs['headers'].get('Accept', 'application/json-rpc')
        self._timeout = self._connect_kwargs.get('timeout')

        self.dispatcher = Dispatcher()
        self.dispatcher.add_method(self._get_plugin_info, 'Plugin.GetPluginInfo')
        self.dispatcher.add_method(self._ping, 'Plugin.Ping')
        self.dispatcher.add_method(self._command, 'Plugin.Command')
        self.dispatcher.add_method(self._get_help, 'Plugin.GetHelp')
        self.dispatcher.add_method(self._list_command, 'Plugin.ListCommand')

    async def ws_connect(self):
        if self.connected:
            raise TransportError('Connection already open.')
        try:
            self._session = aiohttp.ClientSession()
            self._client = await self._session.ws_connect('ws://' + self.addr + '/plugin', **self._connect_kwargs)
            await self._ws_loop()
        except (ClientError, HttpProcessingError, asyncio.TimeoutError) as exc:
            raise TransportError('Error connecting to server', None, exc)
        finally:
            await self.close()

    async def _ws_loop(self):
        msg = None
        try:
            while True:
                msg = await self._client.receive()
                if msg.type == aiohttp.WSMsgType.TEXT:
                    try:
                        data = msg.data
                    except ValueError as exc:
                        raise TransportError('Error Parsing JSON', None, exc)
                    response = JSONRPCResponseManager.handle(data, self.dispatcher)
                    if response:
                        await self._client.send_str(response.json)
                elif msg.type == aiohttp.WSMsgType.CLOSED:
                    break
                elif msg.type == aiohttp.WSMsgType.ERROR:
                    break
        except (ClientError, HttpProcessingError, asyncio.TimeoutError) as exc:
            raise TransportError('Transport Error', None, exc)
        finally:
            await self.close()
            if msg and msg.type == aiohttp.WSMsgType.ERROR:
                raise TransportError('Websocket error detected. Connection closed.')

    async def close(self):
        if self.connected:
            await self._client.close()
        if self._session:
            await self._session.close()
        self._client = None

    @property
    def connected(self):
        return self._client is not None

    @asyncio.coroutine
    def _run(self):
        if self.protocol == 'JsonRPCProtocol':
            yield from self.ws_connect()

    def _ping(self, *args, **kwargs):
        return 'pong'

    def _command(self, *args, **kwargs):
        return self.command(kwargs['name'], kwargs['params'])

    def _get_help(self, *args, **kwargs):
        return self.get_help(args[0])

    def _list_command(self, *args, **kwargs):
        return self.list_command()

    def _get_plugin_info(self, *args, **kwargs):
        return self.get_plugin_info().__dict__

    def run(self):
        asyncio.get_event_loop().run_until_complete(self._run())

    def stop(self, code):
        pass

    def config(self, protocols, connect_addr):
        raise NotImplementedError('Method not implemented!')

    def command(self, name, params):
        raise NotImplementedError('Method not implemented!')

    def get_help(self, name):
        raise NotImplementedError('Method not implemented!')

    def list_command(self):
        raise NotImplementedError('Method not implemented!')

    def get_plugin_info(self):
        raise NotImplementedError('Method not implemented!')
