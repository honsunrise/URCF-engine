import time

import plugin_sdk
import sys

class HelloWorld(plugin_sdk.Plugin):
    def __init__(self):
        super(HelloWorld, self).__init__(version="1.0.0")

    def config(self, protocols, connect_addr):
        return 'JsonRPCProtocol', connect_addr['JsonRPCProtocol']

    def command(self, name, params):
        if name == "echo":
            return params[0]

    def get_help(self, name):
        return "Hello"

    def get_plugin_info(self):
        return plugin_sdk.PluginInfo("HelloWorld", "1.0.0")

    def list_command(self):
        return ["Hello"]


if __name__ == '__main__':
    hello = HelloWorld()
    hello.run()
