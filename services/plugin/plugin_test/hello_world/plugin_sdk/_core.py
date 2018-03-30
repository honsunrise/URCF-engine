import os
from concurrent import futures

import grpc
from grpc_health.v1 import health_pb2, health_pb2_grpc
from grpc_health.v1.health import HealthServicer

from . import command_pb2_grpc
from . import plugin_interface_pb2_grpc

ENV_PLUGIN_LISTENER_ADDRESS = "ENV_PLUGIN_LISTENER_ADDRESS"
ENV_ALLOW_PLUGIN_RPC_PROTOCOL = "ENV_ALLOW_PLUGIN_RPC_PROTOCOL"
ENV_REQUEST_VERSION = "ENV_REQUEST_VERSION"

MSG_COREVERSION = "CoreVersion"
MSG_VERSION = "Version"
MSG_ADDRESS = "Address"
MSG_RPC_PROTOCOL = "RPCProtocol"

GRPCProtocol = 1


class _CommandServicer(command_pb2_grpc.CommandInterfaceServicer):
    """Implementation of Command service."""

    def __init__(self, handler):
        self.handler = handler

    def Command(self, request, context):
        self.handler.command(request.name)

    def GetHelp(self, request, context):
        self.handler.get_help(request.name)


class _PluginServicer(plugin_interface_pb2_grpc.PluginInterfaceServicer):
    """Implementation of Plugin service."""

    def __init__(self, server, handler):
        self._server = server
        self._handler = handler

    def Initialization(self, request, context):
        pass

    def Deploy(self, request, context):
        if request.name == "command":
            command_pb2_grpc.add_CommandInterfaceServicer_to_server(_CommandServicer(self._handler), self._server)

    def UnInitialization(self, request, context):
        pass


class Plugin(object):
    def __init__(self, env=None):
        if env == None:
            env = os.environ
        self.listener_addr = env.get('ENV_PLUGIN_LISTENER_ADDRESS')
        self.rpc_protocol = env.get('ENV_ALLOW_PLUGIN_RPC_PROTOCOL')
        self.version = env.get('ENV_REQUEST_VERSION')
        if self.listener_addr == None or self.rpc_protocol == None or self.version == None:
            raise NotImplementedError('You must run this program as plugin.')

    def serve(self):
        # We need to build a health service to work with go-plugin
        health = HealthServicer()
        health.set("command", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))
        # Start the server.
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
        plugin_interface_pb2_grpc.add_PluginInterfaceServicer_to_server(_PluginServicer(server, self), server)
        health_pb2_grpc.add_HealthServicer_to_server(health, server)
        server.add_insecure_port(self.listener_addr)
        server.start()
        dataOut = os.fdopen(3)
        print("%s: %s" % (MSG_COREVERSION, "1.0.0"), file=dataOut, flush=True)
        print("%s: %s" % (MSG_VERSION, "1.0.0"), file=dataOut, flush=True)
        print("%s: %s" % (MSG_ADDRESS, self.listener_addr), file=dataOut, flush=True)
        print("%s: %s" % (MSG_RPC_PROTOCOL, "1"), file=dataOut, flush=True)

    def stop(self, code):
        pass

    def config(self):
        raise NotImplementedError('Method not implemented!')

    def command(self, name):
        raise NotImplementedError('Method not implemented!')

    def get_help(self, name):
        raise NotImplementedError('Method not implemented!')
