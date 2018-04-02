import time

import plugin_sdk


class HelloWorld(plugin_sdk.Plugin):
    def config(self):
        pass

    def command(self, name):
        if name == "Hello":
            return "World"

    def get_help(self, name):
        return "Hello"

    def list_command(self):
        return ["Hello"]

def serve():
    # We need to build a health service to work with go-plugin
    hello = HelloWorld()
    hello.serve()
    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        hello.stop(0)


if __name__ == '__main__':
    serve()
