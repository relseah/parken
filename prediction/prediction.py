import csv
from datetime import datetime, timedelta
import pickle

import pendulum
import numpy.ma as ma

HISTORICAL_DATA_DIRECTORY = 'historical-data'
RAW_DATA_DIRECTORY = 'raw-data'

FEATURE_NAMES_LONG_TERM = ['parking_id', 'minute', 'hour', 'day_month',
                           'day_year', 'month', 'year']
FEATURE_NAMES_SHORT_TERM = FEATURE_NAMES_LONG_TERM + \
    ['lag_{}h'.format(h) for h in range(1, 25)]

TARGET_NAMES = ['spots']


def generate_lt_sample(parking_id, time):
    return [parking_id, time.minute, time.hour, time.day,
            time.timetuple().tm_yday, time.month, time.year]


def generate_st_sample(parking_id, dt, occupancies):
    return generate_lt_sample(parking_id, dt) + \
        [occupancies.get(dt.subtract(hours=h), -1)
         for h in range(1, 25)]


def mask_samples(samples):
    mask = [[datapoint == -1 for datapoint in sample] for sample in samples]
    return ma.array(samples, mask=mask)


def load_model(filename):
    with open(filename, 'rb') as file:
        return pickle.load(file)


def predict(model, parking_id, from_dt, to_dt):
    datetimes = []
    X = []
    while from_dt < to_dt:
        datetimes.append(from_dt)
        X.append(generate_lt_sample(parking_id, from_dt))
        from_dt = from_dt.add(minutes=10)
    y = model.predict(X)
    return {str(dt): int(y[i]) for i, dt in enumerate(datetimes)}


def predict_today(model, parking_id, occupancies):
    datetimes = []
    X = []
    from_dt = pendulum.datetime(2023, 4, 10)
    to_dt = from_dt.add(days=1)
    while from_dt < to_dt:
        datetimes.append(from_dt)
        X.append(generate_st_sample(parking_id, from_dt, occupancies))
        from_dt = from_dt.add(minutes=10)
    X = mask_samples(X)
    y = model.predict(X)
    return {str(dt): int(y[i]) for i, dt in enumerate(datetimes)}
