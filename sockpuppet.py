#!/usr/bin/python

import json
import random
import string
import logging

import tornado.httpclient
import tornado.gen
import tornado.options
import tornado.web
import tornado.websocket


def parse_payload(data):
    try:
        body = None
        frame = json.loads(data)
        if isinstance(frame, list) and len(frame) > 0:
            frame = frame[0]
            body = json.loads(frame.get('body', "{}"))

        return frame, body

    except Exception as ex:
        logging.exception(ex)

    return None, None


@tornado.gen.coroutine
def connect_and_read_websocket():

    url = tornado.options.options.ws_addr
    try:
        logging.info("connecting to: %s", url)
        w = yield tornado.websocket.websocket_connect(url, connect_timeout=5)
        logging.info("connected to %s, waiting for messages", url)
    except Exception as ex:
        logging.error("couldn't connect, err: %s", ex)
        return

    while True:
        payload = yield w.read_message()
        if payload is None:
            logging.error("uh oh, we got disconnected")
            return

        if payload[0] == "o":
            cookie = ''.join(random.choice(string.ascii_letters + string.digits) for _ in range(32))
            msg = json.dumps(['{"action":"login", "client_app":"hermes.push", "cookies":{"nyt-s":"%s"}}' % cookie])
            w.write_message(msg.encode('utf8'))
            logging.info("sent cookie: %s", cookie)

        elif payload[0] == 'h':
            logging.info('ping')

        elif payload[0] == 'a':
            frame, body = parse_payload(payload[1:])
            if not frame:
                logging.warn("unknown paylod: %s", payload.decode('utf8'))
                continue

            if frame['product'] == "core" and frame['project'] == "standard":
                logging.info("Logged in ok")

            elif frame['product'] == "hermes" and frame['project'] == "push" and body:
                logging.warn("NEW! %s: %s", body['label'], body['title'])

        else:
            logging.warn("unknown paylod: %s", payload.decode('utf8'))


if __name__ == '__main__':
    tornado.options.define(name="ws_addr", type=str, help="Address of the websocket host to connect to.")
    tornado.options.parse_command_line()
    if not tornado.options.options.ws_addr:
        tornado.options.print_help()
        exit()

    tornado.ioloop.IOLoop.instance().run_sync(connect_and_read_websocket)
