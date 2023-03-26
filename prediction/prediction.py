import csv
from datetime import datetime, timedelta
import pickle


HISTORICAL_DATA_DIRECTORY = 'historical_data'


FEATURE_NAMES = ['parking_id', 'minute', 'hour', 'day_month',
                 'day_year', 'month', 'year']
TARGET_NAMES = ['spots']


def generate_sample(parking_id, time):
    return [parking_id, time.minute, time.hour, time.day,
            time.timetuple().tm_yday, time.month, time.year]


def load_model(filename):
    with open(filename, 'rb') as file:
        return pickle.load(file)


def predict(model, parking_id, from_time, to_time):
    interval = timedelta(minutes=10)
    times = []
    X = []
    while from_time < to_time:
        times.append(from_time)
        X.append(generate_sample(parking_id, from_time))
        from_time += interval
    y = model.predict(X)
    return {time.isoformat(): int(y[i]) for i, time in enumerate(times)}
