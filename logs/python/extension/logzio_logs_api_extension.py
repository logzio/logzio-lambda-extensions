#!/bin/sh
''''exec python -u -- "$0" ${1+"$@"} # '''
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: MIT-0

import concurrent.futures
import gzip
import json
import logging
import logging.config
import os
import urllib
import sys

sys.path.append("/opt/python")

import requests
from logzio_logs_api_extension.http_listener import http_server_init, RECEIVER_PORT
from logzio_logs_api_extension.logs_api_client import LogsAPIClient
from logzio_logs_api_extension.extensions_api_client import ExtensionsAPIClient
from queue import Queue

MAX_BULK_SIZE_IN_BYTES = 10 * 1024 * 1024  # 10 MB
MAX_LOG_SIZE_IN_BYTES = 500000
TIMEOUT = int(os.getenv("THREAD_TIMEOUT", 5))  # seconds

SHIPPER_NAME = "logs-lambda-extension"
EXTENSION_VERSION = "0.0.1"

class LogsAPIHTTPExtension():
    def __init__(self, agent_name, registration_body, subscription_body, logger):
        self.logger = logger
        self.logger.info(f"Initializing LogsAPIExternalExtension {agent_name}")
        self.logzio_token = os.environ["LOGZIO_LOGS_TOKEN"]
        self.logzio_listener = self.get_logzio_listener()
        self.agent_name = agent_name
        self.queue = Queue()
        self.logs_api_client = LogsAPIClient()
        self.extensions_api_client = ExtensionsAPIClient()

        # Register early so Runtime could start in parallel
        self.agent_id = self.extensions_api_client.register(self.agent_name, registration_body)

        # Start listening before Logs API registration
        http_server_init(self.queue)
        self.logs_api_client.subscribe(self.agent_id, subscription_body)

    def run_forever(self):
        self.logger.debug(f"Serving LogzioLogsAPIHTTPExternalExtension {self.agent_name}")
        self.logger.debug("Creating new Session for this run")
        session = self.create_session()
        while True:
            # Process the received batches if any.
            self.logger.debug(f"Current queue size: {self.queue.qsize()}")
            if not self.queue.empty():
                with concurrent.futures.ThreadPoolExecutor(max_workers=self.queue.qsize()) as executor:
                    futures = []
                    while not self.queue.empty():
                        batch = self.queue.get_nowait()
                        futures.append(executor.submit(self.parse_and_send, batch, session))
                    for future in concurrent.futures.as_completed(futures, timeout=TIMEOUT):
                        is_batch_successful = future.result()
                        if is_batch_successful:
                            self.logger.info("Batch sent without errors")
                        else:
                            self.logger.warning("Some errors occurred while sending batch")
            resp = self.extensions_api_client.next(self.agent_id)

    def parse_and_send(self, batch, session):
        self.logger.debug(f"batch: {batch}")
        send_timeout = TIMEOUT
        error_occurred = False
        bulks = self.parse_batch(batch)
        bulks_futures = []
        with concurrent.futures.ThreadPoolExecutor(max_workers=len(bulks)) as executor:
            for bulk in bulks:
                bulks_futures.append(executor.submit(self.send_batch_to_logzio, bulk, session))
            for future in concurrent.futures.as_completed(bulks_futures, timeout=send_timeout):
                status_code, start_index, end_index = future.result()
                if status_code != 200:
                    self.logger.debug(f"Error occurred for batch: batch start index: {start_index}, batch end index: {end_index}")
                    error_occurred = True
                    # Bad request
                    if status_code == 400:
                        self.logger.error(f"Logz.io listener returned {status_code}.")
                        corrupted_batch = batch[start_index:end_index]
                        if start_index == end_index:
                            corrupted_batch = batch[start_index]
                        self.logger.error(f"{corrupted_batch} caused 400 response")
                    # In the following cases, we return the batch to the queue to try sending it again
                    else:
                        msg = f"Logz.io listener returned {status_code}. Batch will not be returned to queue. " \
                                  f"Logs in the batch may be dropped."
                        if status_code == 401:
                            msg = f"Logz.io listener returned {status_code}. The token query string parameter is " \
                                  f"missing or not valid. Make sure you’re using the right account token."
                        elif status_code == 500:
                            msg = f"Logz.io listener returned {status_code}. Returning batch to queue."

                        self.return_batch_to_queue(batch, msg, start_index, end_index)
                else:
                    self.logger.info("Bulk sent successfully to Logz.io!")
        return not error_occurred

    def return_batch_to_queue(self, batch, msg, start, end):
        return_batch = batch[start:end]
        self.logger.error(msg)
        if start == end:
            return_batch = [batch[start]]
        self.logger.debug(f"Returning part of batch ({start}, {end}): {return_batch}")
        self.queue.put(return_batch)

    def parse_batch(self, batch):
        # parse batch to a logz.io format
        batch_is_done = False
        current_batch = batch
        bulks = []
        while not batch_is_done:
            request_data = ""
            bulk_start_index = batch.index(current_batch[0])
            for log in current_batch:
                separator = "\n"
                if log is current_batch[-1]:
                    separator = ""
                new_log_line = self.parse_log(log, separator)
                if new_log_line is None:
                    continue
                size_of_new_line = sys.getsizeof(new_log_line)
                if size_of_new_line >= MAX_LOG_SIZE_IN_BYTES:
                    self.logger.error(f"Log line: {new_log_line} size ({size_of_new_line} bytes) is larger than "
                                        f"allowed. Dropping log.")
                    continue
                if sys.getsizeof(request_data) + size_of_new_line <= MAX_BULK_SIZE_IN_BYTES:
                    request_data = request_data + new_log_line
                    if log == current_batch[-1]:
                        self.logger.debug("Finished with batch")
                        bulks.append({"bulk": request_data, "start_index_in_batch": bulk_start_index,
                                      "end_index_in_batch": batch.index(log)})
                        batch_is_done = True
                else:
                    # if request_data exceeded MAX_BULK_SIZE_IN_BYTES we'll need to split batch into bulks
                    self.logger.info("Batch exceeds allowed bulk size. Breaking batch to several bulks to send to "
                                     "Logz.io")
                    bulks.append({"bulk": request_data, "start_index_in_batch": bulk_start_index,
                                  "end_index_in_batch": batch.index(log) - 1})
                    current_batch = batch[batch.index(log):]
                    break
        return bulks

    def parse_log(self, log, separator):
        # drop logs that contain only whitespace
        if log["record"] == 1 and log["record"].isspace():
            self.logger.debug("Dropping new line log")
            return None
        new_log = {"@timestamp": log["time"], "type": "logs-lambda-extension-python",
                   "lambda.log.type": log["type"]}

        if type(log["record"]) is str:
            new_log["message"] = log["record"]
        else:
            for key, value in log["record"].items():
                new_log["lambda.log.{}".format(key)] = value
        return f"{json.dumps(new_log)}{separator}"

    def get_logzio_listener(self):
        # Prioritize custom listener, if exists
        if "LOGZIO_CUSTOM_LISTENER" in os.environ and os.environ["LOGZIO_CUSTOM_LISTENER"] != "":
            custom_listener = os.environ["LOGZIO_CUSTOM_LISTENER"]
            self.logger.info(f"Using custom listener: {custom_listener}")
            return custom_listener

        base_listener = "https://listener.logz.io:8071"
        region = os.getenv("LOGZIO_REGION", "")  # defaults to us region
        if region == "us" or region == "":
            return base_listener
        return base_listener.replace("listener", f"listener-{region}")

    def send_batch_to_logzio(self, bulk_obj, session):
        try:
            self.logger.info("Preparing to send logs")
            url = "{}/?token={}".format(self.logzio_listener, self.logzio_token)
            compressed = gzip.compress(bulk_obj["bulk"].encode('utf-8'))
            res = session.post(url=url, data=compressed)
            if res.status_code != 200:
                self.logger.error(f"Error sending logs to Logz.io: {res.status_code} {res.text}")
            return res.status_code, bulk_obj["start_index_in_batch"], bulk_obj["end_index_in_batch"]
        except urllib.error.HTTPError as e:
            body = e.read().decode()
            self.logger.error(f"Error occurred while sending batch to Logz.io. Response from Logz.io: {body}")
        except Exception as e:
            self.logger.error(f"Error occurred while sending batch to Logz.io: {e}")
        return None

    def create_session(self):
        session = requests.Session()
        session.headers["content-encoding"] = "gzip"
        session.headers["content-type"] = "application/json"
        session.headers["logzio-shipper"] = f"{SHIPPER_NAME}/v{EXTENSION_VERSION}/0,0"
        return session


# Register for the INVOKE events from the EXTENSIONS API
_REGISTRATION_BODY = {
    "events": ["INVOKE", "SHUTDOWN"],
}

# Subscribe to platform logs and receive them on ${local_ip}:4243 via HTTP protocol.

TIMEOUT_MS = 1000  # Maximum time (in milliseconds) that a batch would be buffered.
MAX_BYTES = 262144  # Maximum size in bytes that the logs would be buffered in memory.
MAX_ITEMS = 10000  # Maximum number of events that would be buffered in memory.


def main():
    logger = get_logger()
    subscription_body = get_subscription_body()
    logger.debug(f"Starting Extensions {_REGISTRATION_BODY} {subscription_body}")
    # Note: Agent name has to be file name to register as an external extension
    ext = LogsAPIHTTPExtension(os.path.basename(__file__), _REGISTRATION_BODY, subscription_body, logger)
    ext.run_forever()


def get_lambda_log_types():
    types = ["function"]
    if os.getenv("ENABLE_PLATFORM_LOGS", "false") == "true":
        types.append("platform")
    if os.getenv("ENABLE_EXTENSION_LOGS", "false") == "true":
        types.append("extension")
    return types


def get_subscription_body():
    subscription_body = {
        "destination": {
            "protocol": "HTTP",
            "URI": f"http://sandbox:{RECEIVER_PORT}",
        },
        "types": get_lambda_log_types(),
        "buffering": {
            "timeoutMs": TIMEOUT_MS,
            "maxBytes": MAX_BYTES,
            "maxItems": MAX_ITEMS
        }
    }

    return subscription_body


def get_logger():
    log_level = os.getenv("LOGS_EXT_LOG_LEVEL", "INFO").upper()
    # validate entered value, fallback to INFO
    if log_level not in ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]:
        log_level = "INFO"
    logger = logging.getLogger(__name__)
    logger.setLevel(logging.getLevelName(log_level))
    handler = logging.StreamHandler()
    formatter = logging.Formatter('%(levelname)s %(asctime)s %(module)s %(thread)d %(message)s')
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    return logger


if __name__ == "__main__":
    main()
