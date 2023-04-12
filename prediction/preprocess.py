import os
import csv
import re
import sys
from datetime import datetime
from collections import deque

import pendulum

from prediction import HISTORICAL_DATA_DIRECTORY, RAW_DATA_DIRECTORY, generate_lt_sample, generate_st_sample


def save(filename, rows):
    with open(filename, 'w', newline='') as file:
        writer = csv.writer(file)
        writer.writerows(rows)


if __name__ == '__main__':
    pattern = re.compile('P(\d+) -')
    raw_data = {}
    for filename in os.listdir(HISTORICAL_DATA_DIRECTORY):
        match = pattern.match(filename)
        if match is None:
            sys.exit("invalid filename for dump: " + filename)
        with open(os.path.join(HISTORICAL_DATA_DIRECTORY, filename), encoding='utf-8') as file:
            occupancies = {pendulum.parse(raw_occupancy['observationDateTime']): int(raw_occupancy['availableSpotNumber'])
                           for raw_occupancy in csv.DictReader(file)}
            raw_data[int(match.group(1))] = occupancies

    data_long_term, data_short_term, target = [], [], []
    if not os.path.exists(RAW_DATA_DIRECTORY):
        os.mkdir(RAW_DATA_DIRECTORY)
    for parking_id, occupancies in raw_data.items():
        save(os.path.join(RAW_DATA_DIRECTORY, str(
            parking_id) + '.csv'), occupancies.items())
        for dt, spots in occupancies.items():
            long_term_sample = generate_lt_sample(parking_id, dt)
            data_long_term.append(long_term_sample)
            short_term_sample = generate_st_sample(parking_id, dt, occupancies)
            data_short_term.append(short_term_sample)
            target.append(spots)
    save('data-long-term.csv', data_long_term)
    save('data-short-term.csv', data_short_term)
    save('target.csv', [target])
