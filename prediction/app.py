from datetime import datetime
import os
import csv

from flask import Flask, request, abort, make_response
import numpy as np
import pendulum

from prediction import RAW_DATA_DIRECTORY, load_model, predict, predict_today

model_lt = load_model('model-long-term.dat')
model_st = load_model('model-short-term.dat')
raw_data = {}
for filename in os.listdir(RAW_DATA_DIRECTORY):
    with open(os.path.join(RAW_DATA_DIRECTORY, filename)) as f:
        parking_id = int(filename[:filename.index('.')])
        raw_data[parking_id] = {pendulum.parse(
            row[0]): row[1] for row in csv.reader(f)}

app = Flask(__name__)


@ app.route('/api/prediction')
def prediction():
    try:
        parking_id = int(request.args['id'])
        from_dt = pendulum.parse(request.args['from'])
        to_dt = pendulum.parse(request.args['to'])
    except (KeyError, ValueError):
        abort(400)
    return predict(model_lt, parking_id, from_dt, to_dt), {'Access-Control-Allow-Origin': '*'}


@ app.route('/api/prediction-today')
def prediction_today():
    try:
        parking_id = int(request.args['id'])
    except (KeyError, ValueError):
        abort(400)
    return predict_today(model_st, parking_id, raw_data[parking_id]), {'Access-Control-Allow-Origin': '*'}


@ app.route('/api/true-occupancy')
def true_occupancy():
    with open('showcase/raw-data.csv') as f:
        occupancies = {str(pendulum.parse(
            row[0])): row[1] for row in csv.reader(f)}
    return occupancies, {'Access-Control-Allow-Origin': '*'}
