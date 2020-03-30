import flask
import logging
import json
from flask import request, jsonify
from flask_sockets import Sockets
from logging.config import dictConfig

dictConfig({
    'version': 1,
    'formatters': {'default': {
        'format': '[%(asctime)s] %(levelname)s in %(module)s: %(message)s',
    }},
    'handlers': {'wsgi': {
        'class': 'logging.StreamHandler',
        'stream': 'ext://flask.logging.wsgi_errors_stream',
        'formatter': 'default'
    }},
    'root': {
        'level': 'INFO',
        'handlers': ['wsgi']
    }
})
log = logging.getLogger('api')
app = flask.Flask(__name__)
app.config['SECRET_KEY'] = 'secret!'
app.config["DEBUG"] = True

sockets = Sockets(app)


clients =[]
host ='local.appspot.com'
@app.route('/',methods=['GET'])
def home():
    return "hi", 200 

@app.route('/ntf/<msg>', methods=['GET'])
def ntf(msg):
    for c in clients:
        if not c.closed:
            log.info('ntf' + msg)
            c.send('blah')
    return {'Status':'OK'}

@app.route('/service/json/1/<service>', methods=['GET'])
def servie_locator(service):
    log.info(f"Locating: {service}")
    return {'Host':host, 'Status':'OK'}


@app.route('/token/json/2/device/new', methods=['POST'])
def device_new():
    d = json.loads(request.get_data().decode())
    log.info(d)
    code = d["code"]
    device_desc = d["deviceDesc"]
    device_id = d["deviceID"]
    return "some device token"

@app.route('/token/json/2/user/new', methods=['POST'])
def user_new():
    auth = request.headers["Authorization"]
    app.logger.info(f"Got auth token: {auth}")
    return "some user token"

@sockets.route('/notifications/ws/json/1')
def wsstuff(ws):
    global clients
    clients.append(ws)
    while not ws.closed:
        message = ws.receive()
        print(message)
    clients.remove(ws)

@app.route('/service/json/1/document-storage', methods=['GET'])
def storage():
    return jsonify({'Host':host, 'Status':'OK'})

@app.route('/document-storage/json/2/upload/request', methods=['PUT'])
def upload_doc_request():
    b = request.get_json()
    log.info(b)
    if b: 
        id = b[0]['ID']
    else:
        id = 'somenewid'
    response = [{
        #'BlobURLPut':f'https://{host}/upload?id={id}',
        'BlobURLPut':f'http://localhost:8000/upload?id={id}',
        'ID':id,
        'Message':'',
        'Success':True,
        'Version':1
        }]
    return jsonify(response)

@app.route('/upload', methods=['PUT'])
def upload():
    app.logger.info('uploading')
    b = request.args.get('id')
    data= request.data
    with open('temp.zip','wb') as f:
        f.write(data)
    return jsonify({}, 200)

@app.route('/download', methods=['GET'])
def download():
    with open('temp.zip','rb') as f:
        data = f.read()
    return data


@app.route('/document-storage/json/2/upload/update-status', methods=['PUT'])
def update_document_status():
    b = request.get_json()
    app.logger.info(b)
    return ""

@app.route('/document-storage/json/2/delete', methods=['PUT'])
def delete():
    r = request.get_json()[0]
    app.logger.warn('deleting' + r['ID'])
    return jsonify({})

@app.route('/document-storage/json/2/docs', methods=['GET'])
def get_documents():
    withBlob = request.args.get('withBlob')
    doc_id = request.args.get('docId')
    
    b = {
        "BlobURLGet": f"https://{host}/download?id={doc_id}",
        "BlobURLGetExpires": "0001-01-01T00:00:00Z",
        "Bookmarked": False,
        "CurrentPage": 0,
        "ID": "01c8689d-b135-468a-924d-0b79456bc6ae",
        "Message": "",
        "ModifiedClient": "2020-03-27T20:04:46.303303Z",
        "Parent": "",
        "Success": True,
        "Type": "DocumentType",
        "Version": 4,
        "VissibleName": "Stuff"
    }

    return jsonify([b])

@app.route('/api/v2/document', methods=["POST"])
def document_email():
    v = request.form
    log.info(v)
    for u in v.keys():
        print(u)
    for f in request.files:
        print(f)
    return {}

@app.route('/api/v1/page', methods=["POST"])
def document_hwr():
    v = request.get_json()
    #log.info(v)
    with open('blah.json','r') as f:
        content = f.read()
    return content

@app.errorhandler(404)
def stuff(e):
    app.logger.info('got 404')
    return jsonify({'Error':'not found'}), 404

from gevent import pywsgi
from geventwebsocket.handler import WebSocketHandler
if __name__ == '__main__':
    gunicorn_logger = logging.getLogger('gunicorn.error')
    app.logger.handlers = gunicorn_logger.handlers
    logging.handlers = gunicorn_logger.handlers
    app.logger.setLevel(gunicorn_logger.level)
    server = pywsgi.WSGIServer(('0.0.0.0', 3000), app, handler_class=WebSocketHandler)
    server.serve_forever()
