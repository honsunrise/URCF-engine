import time

import plugin_sdk


class HelloWorld(plugin_sdk.Plugin):
    def __init__(self):
        super(HelloWorld, self).__init__(version="1.0.0")

    def config(self, protocols, connect_addr):
        return 'JsonRPCProtocol', connect_addr['JsonRPCProtocol']

    def command(self, name):
        if name == "Hello":
            return "World"

    def get_help(self, name):
        return "Hello"

    def get_plugin_info(self):
        return plugin_sdk.PluginInfo("HelloWorld", "1.0.0")

    def list_command(self):
        return ["Hello"]


if __name__ == '__main__':
    # We need to build a health service to work with go-plugin
    hello = HelloWorld()
    hello.run()
    time.sleep(2)
