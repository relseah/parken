from datetime import datetime

from flask import Flask, request, abort, make_response
import numpy as np

from prediction import load_model, predict

model = load_model('model.dat')

app = Flask(__name__)


@app.route('/api/prediction')
def prediction():
    try:
        from_time = datetime.fromisoformat(request.args['from'])
        to_time = datetime.fromisoformat(request.args['to'])
    except (KeyError, ValueError):
        abort(400)
    if not 'id' in request.args:
        abort(400)
    return predict(model, request.args['id'], from_time, to_time), {'Access-Control-Allow-Origin': '*'}
