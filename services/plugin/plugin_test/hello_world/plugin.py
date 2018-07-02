import time

import plugin_sdk


class HelloWorld(plugin_sdk.Plugin):
    def __init__(self):
        super(HelloWorld, self).__init__(version="1.0.0")

    def config(self):
        return 'JsonRPCProtocol', self.connect_addr['connect_addr']

    def command(self, name):
        if name == "Hello":
            return "World"

    def get_help(self, name):
        return "Hello"

    def list_command(self):
        return ["Hello"]

if __name__ == '__main__':
    # We need to build a health service to work with go-plugin
    hello = HelloWorld()
    hello.run()
