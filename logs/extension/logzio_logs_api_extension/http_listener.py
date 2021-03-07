# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import json
import logging
import os
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Event, Thread

""" Demonstrates code to set up an HTTP listener and receive log events
"""

RECEIVER_IP = "0.0.0.0"
RECEIVER_PORT = 4243


def http_server_init(queue):
    logger = get_logger()

    def handler(*args):
        LogsHandler(logger, queue, *args)

    logger.info(f"Initializing HTTP Server on {RECEIVER_IP}:{RECEIVER_PORT}")
    server = HTTPServer((RECEIVER_IP, RECEIVER_PORT), handler)

    # Ensure that the server thread is scheduled so that the server binds to the port
    # and starts to listening before subscribe for the logs and ask for the next event.
    started_event = Event()
    server_thread = Thread(target=serve, daemon=True, args=(started_event, server, logger,))
    server_thread.start()
    rc = started_event.wait(timeout=9)
    if rc is not True:
        raise Exception("server_thread has timedout before starting")


# Server implementation
class LogsHandler(BaseHTTPRequestHandler):
    def __init__(self, logger, queue, *args):
        self.logger = logger
        self.queue = queue
        BaseHTTPRequestHandler.__init__(self, *args)

    def do_POST(self):
        try:
            cl = self.headers.get("Content-Length")
            if cl:
                data_len = int(cl)
            else:
                data_len = 0
            content = self.rfile.read(data_len)
            self.send_response(200)
            self.end_headers()
            batch = json.loads(content.decode("utf-8"))
            self.queue.put(batch)

        except Exception as e:
            self.logger.error(f"Error processing message: {e}")


# Server thread
def serve(started_event, server, logger):
    # Notify that this thread is up and running
    started_event.set()
    try:
        logger.info(f"Serving HTTP Server on {RECEIVER_IP}:{RECEIVER_PORT}")
        server.serve_forever()
    except:
        logger.error(f"Error in HTTP server {sys.exc_info()[0]}")
    finally:
        if server:
            server.shutdown()


def get_logger():
    log_level = os.getenv("LOG_LEVEL", "INFO").upper()
    # validate entered value, fallback to INFO
    if log_level not in ["DEBUG", "INFO", "WARNING", "WARNING", "ERROR", "CRITICAL"]:
        log_level = "INFO"
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.getLevelName(log_level))
    handler = logging.StreamHandler()
    formatter = logging.Formatter('%(levelname)s %(asctime)s %(module)s %(process)d %(thread)d %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    return logger
